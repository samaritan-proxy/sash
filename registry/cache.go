// Copyright 2019 Samaritan Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registry

import (
	"context"
	"math/rand"
	"reflect"
	"sync"
	"time"

	backoff "github.com/cenkalti/backoff/v3"
	"github.com/samaritan-proxy/sash/logger"
	"github.com/samaritan-proxy/sash/model"
)

//go:generate mockgen -source ../model/service.go -destination mock_registry_test.go -package registry

var (
	defaultBackoffInitialInterval     = 100 * time.Millisecond
	defaultBackoffRandomizationFactor = 0.2
	defaultBackoffMultiplier          = 1.6
	defaultBackoffMaxInterval         = time.Second
	defaultBackoffMaxRetries          = 10
)

type cacheOptions struct {
	syncFreq   time.Duration
	syncJitter float64
}

func defaultBackOff() *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = defaultBackoffInitialInterval
	b.RandomizationFactor = defaultBackoffRandomizationFactor
	b.Multiplier = defaultBackoffMultiplier
	b.MaxInterval = defaultBackoffMaxInterval
	b.Reset()
	return b
}

func defaultCacheOptions() *cacheOptions {
	return &cacheOptions{
		syncFreq:   5 * time.Second,
		syncJitter: 0.2,
	}
}

// cacheOption sets the options for cache.
type cacheOption func(o *cacheOptions)

// SyncFreq sets the sync frequency.
func SyncFreq(freq time.Duration) cacheOption {
	return func(o *cacheOptions) {
		o.syncFreq = freq
	}
}

// SyncJitter sets the jitter for sync frequency.
func SyncJitter(jitter float64) cacheOption {
	return func(o *cacheOptions) {
		o.syncJitter = jitter
	}
}

// Cache is used to cache all registered services from the underlying registry,
// and provides a event dispatch mechanism which means the caller could receive
// and handle the service and instance change event.
type Cache interface {
	model.ServiceRegistry
	// RegisterServiceEventHandler registers a handler to handle service event.
	RegisterServiceEventHandler(handler ServiceEventHandler)
	// RegisterInstanceEventHandler registers a handler to handle instance event.
	RegisterInstanceEventHandler(handler InstanceEventHandler)
}

// NewCache creates a cache container for service registry.
func NewCache(r model.ServiceRegistry, opts ...cacheOption) Cache {
	c, ok := r.(Cache)
	// if already implements the Cache interface, return it immediately.
	if ok {
		return c
	}
	return newCache(r, opts...)
}

// cache is an implementation of Cache.
type cache struct {
	rwMu    sync.RWMutex
	options *cacheOptions
	r       model.ServiceRegistry

	services    map[string]*model.Service
	svcEvtHdls  []ServiceEventHandler
	instEvtHdls []InstanceEventHandler
}

func newCache(r model.ServiceRegistry, opts ...cacheOption) *cache {
	// init options
	o := defaultCacheOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &cache{
		r:           r,
		options:     o,
		services:    make(map[string]*model.Service),
		svcEvtHdls:  make([]ServiceEventHandler, 0, 1),
		instEvtHdls: make([]InstanceEventHandler, 0, 1),
	}
}

func (c *cache) List() ([]string, error) {
	c.rwMu.RLock()
	defer c.rwMu.RUnlock()
	names := make([]string, 0, len(c.services))
	for name := range c.services {
		names = append(names, name)
	}
	// TODO: return a temp error when it's under the first sync.
	return names, nil
}

func (c *cache) Get(name string) (*model.Service, error) {
	c.rwMu.RLock()
	defer c.rwMu.RUnlock()
	service := c.services[name]
	// TODO: return a temp error when it's under the first sync.
	return service, nil
}

// RegisterServiceEventHandler registers a handler to handle service event.
// It is not goroutine-safe, should call it before execute Run.
func (c *cache) RegisterServiceEventHandler(handler ServiceEventHandler) {
	c.svcEvtHdls = append(c.svcEvtHdls, handler)
}

// RegisterInstanceEventHandler registers a handler to handle instance event.
// It is not goroutine-safe, should call it before execute Run.
func (c *cache) RegisterInstanceEventHandler(handler InstanceEventHandler) {
	c.instEvtHdls = append(c.instEvtHdls, handler)
}

// Run runs the cache container until the context is canceled or deadline exceeded.
func (c *cache) Run(ctx context.Context) {
	maxSyncInterval := time.Duration(float64(c.options.syncFreq) * (1 + c.options.syncJitter))
	b := defaultBackOff()
	b.MaxInterval = maxSyncInterval
	b.Reset()

	for {
		startTime := time.Now()
		err := c.Sync(ctx)
		select {
		case <-ctx.Done():
			return
		default:
		}

		// calculate the interval before next sync.
		var interval time.Duration
		if err != nil {
			interval = b.NextBackOff()
			logger.Warnf("Sync services failed: %v, retry after %s", err, interval)
		} else {
			// reset the backoff
			b.Reset()
			d := float64(c.options.syncFreq) * (1 + c.options.syncJitter*(rand.Float64()*2-1))
			interval = time.Duration(d)
			logger.Debugf("Sync services succeed, cost: %s, do it again after %s", time.Since(startTime), interval)
		}

		t := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (c *cache) Sync(ctx context.Context) error {
	names, err := c.r.List()
	if err != nil {
		return err
	}

	// remove the outdated services.
	c.deleteOutdatedServices(names)

	eb := defaultBackOff()
	// the total retry time is under six seconds.
	b := backoff.WithMaxRetries(eb, uint64(defaultBackoffMaxRetries))
	// TODO: improve the performance with concurrency.
	for _, name := range names {
		b.Reset()

		var (
			err     error
			service *model.Service
		)
	Retry:
		for {
			service, err = c.r.Get(name)
			if err == nil {
				break
			}

			d := b.NextBackOff()
			// exceeded the max retry times
			if d == backoff.Stop {
				break
			}

			t := time.NewTimer(d)
			select {
			case <-ctx.Done():
				break Retry
			case <-t.C:
			}
		}

		// NOTE: it's also acceptable to skip this error, and sync others.
		if err != nil {
			return err
		}
		c.addOrUpdateService(service)
	}
	return nil
}

func (c *cache) deleteOutdatedServices(newServiceNames []string) {
	// build map to speed up lookup
	m := make(map[string]struct{}, len(newServiceNames))
	for _, name := range newServiceNames {
		m[name] = struct{}{}
	}

	var needDelete []*model.Service
	c.rwMu.RLock()
	for name, service := range c.services {
		_, ok := m[name]
		if !ok {
			needDelete = append(needDelete, service)
		}
	}
	c.rwMu.RUnlock()

	for _, service := range needDelete {
		c.deleteService(service)
	}
}

func (c *cache) deleteService(service *model.Service) {
	c.rwMu.Lock()
	delete(c.services, service.Name)
	c.rwMu.Unlock()
	event := &ServiceEvent{
		Type:    EventDelete,
		Service: service,
	}
	c.dispatchServiceEvent(event)
}

func (c *cache) addOrUpdateService(newService *model.Service) {
	name := newService.Name
	c.rwMu.RLock()
	oldService, ok := c.services[name]
	c.rwMu.RUnlock()
	if !ok {
		c.addService(newService)
		return
	}
	c.updateService(oldService, newService)
}

func (c *cache) addService(service *model.Service) {
	c.rwMu.Lock()
	c.services[service.Name] = service
	c.rwMu.Unlock()
	event := &ServiceEvent{
		Type:    EventAdd,
		Service: service,
	}
	c.dispatchServiceEvent(event)
}

func (c *cache) updateService(oldService, newService *model.Service) {
	serviceName := oldService.Name
	oldInstances := oldService.Instances
	newInstances := newService.Instances

	// deleted
	var deleted []*model.ServiceInstance
	for addr, instance := range oldInstances {
		_, ok := newInstances[addr]
		if !ok {
			deleted = append(deleted, instance)
		}
	}

	// added or updated
	var added, updated []*model.ServiceInstance
	for addr, newInstance := range newInstances {
		oldInstance, ok := oldInstances[addr]
		if !ok {
			added = append(added, newInstance)
			continue
		}

		isEqual := isInstanceEqual(oldInstance, newInstance)
		if !isEqual {
			updated = append(updated, newInstance)
		}
	}

	c.rwMu.Lock()
	c.services[serviceName] = newService
	c.rwMu.Unlock()

	// should emit add event first.
	c.addInstances(serviceName, added)
	c.updateInstances(serviceName, updated)
	c.deleteInstances(serviceName, deleted)
}

func (c *cache) addInstances(serviceName string, instances []*model.ServiceInstance) {
	if len(instances) == 0 {
		return
	}
	event := &InstanceEvent{
		Type:        EventAdd,
		ServiceName: serviceName,
		Instances:   instances,
	}
	c.dispatchInstanceEvent(event)
}

func (c *cache) updateInstances(serviceName string, instances []*model.ServiceInstance) {
	if len(instances) == 0 {
		return
	}
	event := &InstanceEvent{
		Type:        EventUpdate,
		ServiceName: serviceName,
		Instances:   instances,
	}
	c.dispatchInstanceEvent(event)
}

func (c *cache) deleteInstances(serviceName string, instances []*model.ServiceInstance) {
	if len(instances) == 0 {
		return
	}
	event := &InstanceEvent{
		Type:        EventDelete,
		ServiceName: serviceName,
		Instances:   instances,
	}
	c.dispatchInstanceEvent(event)
}

func isInstanceEqual(inst1, inst2 *model.ServiceInstance) bool {
	// TODO: move it to the model package.
	if inst1.Addr != inst2.Addr {
		return false
	}
	if inst1.State != inst2.State {
		return false
	}
	if !reflect.DeepEqual(inst1.Meta, inst1.Meta) {
		return false
	}
	return true
}

func (c *cache) dispatchServiceEvent(event *ServiceEvent) {
	for i := 0; i < len(c.svcEvtHdls); i++ {
		handler := c.svcEvtHdls[i]
		handler(event)
	}
}

func (c *cache) dispatchInstanceEvent(event *InstanceEvent) {
	for i := 0; i < len(c.instEvtHdls); i++ {
		handler := c.instEvtHdls[i]
		handler(event)
	}
}

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

var (
	defaultBackoffInitialInterval     = 100 * time.Millisecond
	defaultBackoffRandomizationFactor = 0.2
	defaultBackoffMultiplier          = 1.6
	defaultBackoffMaxInterval         = time.Second
)

type cacheOptions struct {
	syncFreq   time.Duration
	syncJitter float64
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

// Cache is used to cache all registerd services from the underlying registry,
// and provides a event dispatch mechanism which means the caller could receive
// and handle the service and instance change event.
type Cache interface {
	model.ServiceRegistry
	// RegisterServiceEventHandler registers a handler to handle servcie event.
	RegisterServiceEventHandler(handler ServiceEventHandler)
	// RegisterInstanceEventHandler registers a handler to handle instance event.
	RegisterInstanceEventHandler(handler InstanceEventHandler)
}

// Newcache creates a cache container for service registry.
func Newcache(r model.ServiceRegistry, opts ...cacheOption) Cache {
	c, ok := r.(Cache)
	// If already implements the Cache interface, return it immediately.
	if ok {
		return c
	}
	return newCache(r, opts...)
}

// cache is an implementation of Cache.
type cache struct {
	rwMu sync.RWMutex
	r    model.ServiceRegistry

	options     *cacheOptions
	svcEvtHdls  []ServiceEventHandler
	instEvtHdls []InstanceEventHandler

	services map[string]*model.Service
}

func newCache(r model.ServiceRegistry, opts ...cacheOption) *cache {
	// init options
	o := defaultCacheOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &cache{
		r:       r,
		options: o,
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

// RegisterServiceEventHandler registers a handler to handle servcie event.
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
	b := &backoff.ExponentialBackOff{
		InitialInterval:     defaultBackoffInitialInterval,
		RandomizationFactor: defaultBackoffRandomizationFactor,
		Multiplier:          defaultBackoffMultiplier,
		MaxInterval:         maxSyncInterval,
	}

	for {
		startTime := time.Now()
		err := c.sync(ctx)
		select {
		case <-ctx.Done():
			return
		default:
		}

		// calculate the interval before next sync.
		var interval time.Duration
		if err != nil {
			interval = b.NextBackOff()
			logger.Warnf("Sync services from the underlying registry failed: %v, retry after %d", err, interval)
		} else {
			// reset the backoff
			b.Reset()
			d := float64(c.options.syncFreq) * (1 + c.options.syncJitter*(rand.Float64()*2-1))
			interval = time.Duration(d)
			logger.Debugf("Sync services succeed, cost: %s, will do it again after %d", time.Since(startTime), interval)
		}

		t := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (c *cache) sync(ctx context.Context) error {
	// load
	newServices, err := c.loadFromRegistry(ctx)
	if err != nil {
		return err
	}

	// diff
	c.diffServices(c.services, newServices)

	// save
	c.rwMu.Lock()
	c.services = newServices
	c.rwMu.Unlock()
	return nil
}

func (c *cache) diffServices(oldServices, newServices map[string]*model.Service) {
	// deleted
	for name, service := range oldServices {
		_, ok := newServices[name]
		if !ok {
			c.handleServiceDelete(service)
		}
	}

	// added or updated
	for name, newService := range newServices {
		oldService, ok := oldServices[name]
		if !ok {
			c.handleServiceAdd(newService)
			continue
		}

		// diff them
		c.diffService(oldService, newService)
	}
}

func (c *cache) diffService(oldService, newService *model.Service) {
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
	for addr, newInstace := range newInstances {
		oldInstance, ok := oldInstances[addr]
		if !ok {
			added = append(added, newInstace)
			continue
		}

		isEqual := isInstanceEqual(oldInstance, newInstace)
		if !isEqual {
			updated = append(updated, newInstace)
		}
	}

	// should handle add event first.
	if len(added) != 0 {
		c.handleInstanceAdd(serviceName, added)
	}
	if len(updated) != 0 {
		c.handleInstanceUpdate(serviceName, updated)
	}
	if len(deleted) != 0 {
		c.handleInstanceDelete(serviceName, deleted)
	}
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

func (c *cache) loadFromRegistry(ctx context.Context) (map[string]*model.Service, error) {
	names, err := c.r.List()
	if err != nil {
		return nil, err
	}

	eb := &backoff.ExponentialBackOff{
		InitialInterval:     defaultBackoffInitialInterval,
		RandomizationFactor: defaultBackoffRandomizationFactor,
		Multiplier:          defaultBackoffMultiplier,
		MaxInterval:         defaultBackoffMaxInterval,
	}
	// the total retry time is under six seconds.
	b := backoff.WithMaxRetries(eb, 10)
	services := make(map[string]*model.Service, len(names))
	// TODO: improve the performace with concurrency.
	for _, name := range names {
		b.Reset()

		var (
			err     error
			service *model.Service
		)
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
				break
			case <-t.C:
			}
		}

		if err != nil {
			return nil, err
		}
		services[name] = service
	}
	return services, nil
}

func (c *cache) handleServiceAdd(service *model.Service) {
	event := &ServiceEvent{
		Type:    EventAdd,
		Service: service,
	}
	c.handleServiceEvent(event)
}

func (c *cache) handleServiceDelete(service *model.Service) {
	event := &ServiceEvent{
		Type:    EventDelete,
		Service: service,
	}
	c.handleServiceEvent(event)
}

func (c *cache) handleServiceEvent(event *ServiceEvent) {
	for i := len(c.svcEvtHdls) - 1; i >= 0; i-- {
		handler := c.svcEvtHdls[i]
		handler(event)
	}
}

func (c *cache) handleInstanceAdd(serviceName string, instances []*model.ServiceInstance) {
	event := &InstanceEvent{
		Type:        EventAdd,
		ServiceName: serviceName,
		Instances:   instances,
	}
	c.handleInstanceEvent(event)
}

func (c *cache) handleInstanceUpdate(serviceName string, instances []*model.ServiceInstance) {
	event := &InstanceEvent{
		Type:        EventUpdate,
		ServiceName: serviceName,
		Instances:   instances,
	}
	c.handleInstanceEvent(event)
}

func (c *cache) handleInstanceDelete(serviceName string, instances []*model.ServiceInstance) {
	event := &InstanceEvent{
		Type:        EventDelete,
		ServiceName: serviceName,
		Instances:   instances,
	}
	c.handleInstanceEvent(event)
}

func (c *cache) handleInstanceEvent(event *InstanceEvent) {
	for i := len(c.instEvtHdls) - 1; i >= 0; i-- {
		handler := c.instEvtHdls[i]
		handler(event)
	}
}

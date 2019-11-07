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
	"time"

	backoff "github.com/cenkalti/backoff/v3"
	"github.com/samaritan-proxy/sash/logger"
	"github.com/samaritan-proxy/sash/model"
)

type cacheOptions struct {
	syncFreq   time.Duration
	syncJitter float64
}

func newDefaultCacheOptions() *cacheOptions {
	return &cacheOptions{}
}

type CacheOption func(o *cacheOptions)

type Cache struct {
	model.ServiceRegistry

	options     *cacheOptions
	svcEvtHdls  []ServiceEventHandler
	instEvtHdls []InstanceEventHandler

	services map[string]*model.Service
}

func NewCache(registry model.ServiceRegistry, opts ...CacheOption) *Cache {
	// init options
	o := newDefaultCacheOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &Cache{
		ServiceRegistry: registry,
		options:         o,
	}
}

func (c *Cache) RegisterServiceEventHandler(handler ServiceEventHandler) {
	c.svcEvtHdls = append(c.svcEvtHdls, handler)
}

func (c *Cache) RegisterInstanceEventHandler(handler InstanceEventHandler) {
	c.instEvtHdls = append(c.instEvtHdls, handler)
}

func (c *Cache) Run(ctx context.Context) {
	maxSyncInterval := time.Duration(float64(c.options.syncFreq) * (1 + c.options.syncJitter))
	b := &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.2,
		Multiplier:          1.6,
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

func (c *Cache) sync(ctx context.Context) error {
	// load
	newServices, err := c.loadFromRegistry(ctx)
	if err != nil {
		return err
	}

	// diff
	c.diffServices(c.services, newServices)

	// save
	c.services = newServices
	return nil
}

func (c *Cache) diffServices(oldServices, newServices map[string]*model.Service) {
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

func (c *Cache) diffService(oldService, newService *model.Service) {
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

func (c *Cache) loadFromRegistry(ctx context.Context) (map[string]*model.Service, error) {
	names, err := c.List()
	if err != nil {
		return nil, err
	}

	eb := &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.2,
		Multiplier:          1.6,
		MaxInterval:         time.Second,
	}
	// the total retry time is under six seconds.
	b := backoff.WithMaxRetries(eb, 10)
	services := make(map[string]*model.Service, len(names))
	for _, name := range names {
		b.Reset()

		var (
			err     error
			service *model.Service
		)
		for {
			service, err = c.Get(name)
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

func (c *Cache) handleServiceAdd(service *model.Service) {
	event := &ServiceEvent{
		Type:    EventAdd,
		Service: service,
	}
	c.handleServiceEvent(event)
}

func (c *Cache) handleServiceDelete(service *model.Service) {
	event := &ServiceEvent{
		Type:    EventDelete,
		Service: service,
	}
	c.handleServiceEvent(event)
}

func (c *Cache) handleServiceEvent(event *ServiceEvent) {
	for i := len(c.svcEvtHdls) - 1; i >= 0; i-- {
		handler := c.svcEvtHdls[i]
		handler(event)
	}
}

func (c *Cache) handleInstanceAdd(serviceName string, instances []*model.ServiceInstance) {
	event := &InstanceEvent{
		Type:        EventAdd,
		ServiceName: serviceName,
		Instances:   instances,
	}
	c.handleInstanceEvent(event)
}

func (c *Cache) handleInstanceUpdate(serviceName string, instances []*model.ServiceInstance) {
	event := &InstanceEvent{
		Type:        EventUpdate,
		ServiceName: serviceName,
		Instances:   instances,
	}
	c.handleInstanceEvent(event)
}

func (c *Cache) handleInstanceDelete(serviceName string, instances []*model.ServiceInstance) {
	event := &InstanceEvent{
		Type:        EventDelete,
		ServiceName: serviceName,
		Instances:   instances,
	}
	c.handleInstanceEvent(event)
}

func (c *Cache) handleInstanceEvent(event *InstanceEvent) {
	for i := len(c.instEvtHdls) - 1; i >= 0; i-- {
		handler := c.instEvtHdls[i]
		handler(event)
	}
}

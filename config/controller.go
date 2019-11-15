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

package config

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/samaritan-proxy/sash/logger"
)

const (
	NamespaceService       = "service"
	TypeServiceProxyConfig = "proxy-config"
	TypeServiceDependence  = "dependence"
)

var InterestedNSAndType = map[string][]string{
	NamespaceService: {TypeServiceProxyConfig, TypeServiceDependence},
}

// Controller is used to store configuration information.
type Controller struct {
	sync.Mutex
	store    Store
	updateCh chan struct{}

	cache    atomic.Value // *Config
	interval time.Duration

	evtHdls []EventHandler

	stop chan struct{}
	wg   sync.WaitGroup
}

// NewController return a new Controller.
func NewController(store Store, interval time.Duration) *Controller {
	c := &Controller{
		store:    store,
		updateCh: make(chan struct{}, 1),
		interval: interval,
		stop:     make(chan struct{}),
		cache:    atomic.Value{},
	}
	return c
}

func (c *Controller) loadCache() *Cache {
	return c.cache.Load().(*Cache)
}

func (c *Controller) storeCache(cfg *Cache) {
	c.cache.Store(cfg)
}

func (c *Controller) trySubscribe(namespace ...string) error {
	ss, ok := c.store.(SubscribableStore)
	if !ok {
		return nil
	}
	for _, ns := range namespace {
		if err := ss.Subscribe(ns); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) fetchAll() (*Cache, error) {
	cache := NewCache()
	for ns, types := range InterestedNSAndType {
		for _, typ := range types {
			keys, err := c.store.GetKeys(ns, typ)
			switch err {
			case nil:
			case ErrNamespaceNotExist, ErrTypeNotExist:
				continue
			default:
				return nil, err
			}
			for _, key := range keys {
				value, err := c.store.Get(ns, typ, key)
				if err != nil {
					return nil, err
				}
				cache.Set(ns, typ, key, value)
			}
		}
	}
	return cache, nil
}

// Start start the controller.
func (c *Controller) Start() error {
	if err := c.store.Start(); err != nil {
		return err
	}
	// init Controller
	cache, err := c.fetchAll()
	if err != nil {
		return err
	}
	c.storeCache(cache)
	for ns := range InterestedNSAndType {
		if err := c.trySubscribe(ns); err != nil {
			return err
		}
	}
	c.wg.Add(2)
	go c.trigger()
	go c.loop()
	return nil
}

// Stop stop the controller.
func (c *Controller) Stop() {
	close(c.stop)
	c.wg.Wait()
	c.store.Stop()
}

func (c *Controller) triggerUpdate() {
	select {
	case c.updateCh <- struct{}{}:
	default:
	}
}

func (c *Controller) trigger() {
	ticker := time.NewTicker(c.interval)
	defer func() {
		ticker.Stop()
		c.wg.Done()
	}()

	var ch <-chan struct{}
	if ss, ok := c.store.(SubscribableStore); ok {
		ch = ss.Event()
	}

	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
		case <-ch:
		}
		c.triggerUpdate()
	}
}

func (c *Controller) diff(that *Cache) {
	add, update, del := c.loadCache().Diff(that)
	do := func(event *Event) {
		for _, hdl := range c.evtHdls {
			hdl(event)
		}
	}
	for _, cfg := range add {
		do(NewEvent(EventAdd, cfg))
	}
	for _, cfg := range update {
		do(NewEvent(EventUpdate, cfg))
	}
	for _, cfg := range del {
		do(NewEvent(EventDelete, cfg))
	}
}

func (c *Controller) loop() {
	defer c.wg.Done()
	for {
		select {
		case <-c.stop:
			return
		case <-c.updateCh:
			c.doUpdate()
		}
	}
}

func (c *Controller) doUpdate() {
	newConf, err := c.fetchAll()
	if err != nil {
		logger.Warnf("failed to fetch config from store, err: %v", err)
		return
	}
	c.diff(newConf)
	c.storeCache(newConf)
}

// RegisterEventHandler registers a handler to handle config event.
// It is not goroutine-safe, should call it before execute Run.
func (c *Controller) RegisterEventHandler(handler EventHandler) {
	c.evtHdls = append(c.evtHdls, handler)
}

// Get return config data by namespace, type and key.
func (c *Controller) Get(namespace, typ, key string) ([]byte, error) {
	return c.loadCache().Get(namespace, typ, key)
}

// Set set config data by namespace, type and key.
func (c *Controller) Set(namespace, typ, key string, value []byte) error {
	return c.store.Set(namespace, typ, key, value)
}

// Del del config data by namespace, type and key.
func (c *Controller) Del(namespace, typ, key string) error {
	return c.store.Del(namespace, typ, key)
}

// Exist return true if config data is exist.
func (c *Controller) Exist(namespace, typ, key string) bool {
	_, err := c.Get(namespace, typ, key)
	return err == nil
}

// Keys return all key by namespace and type
func (c *Controller) Keys(namespace, typ string) ([]string, error) {
	return c.loadCache().Keys(namespace, typ)
}

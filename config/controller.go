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
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v3"

	"github.com/samaritan-proxy/sash/logger"
	"github.com/samaritan-proxy/sash/utils"
)

const (
	NamespaceService       = "service"
	TypeServiceProxyConfig = "proxy-config"
	TypeServiceDependency  = "dependency"

	NamespaceSamaritan    = "samaritan"
	TypeSamaritanInstance = "instance"
)

var InterestedNSAndType = map[string][]string{
	NamespaceService:   {TypeServiceProxyConfig, TypeServiceDependency},
	NamespaceSamaritan: {TypeSamaritanInstance},
}

var (
	defaultBackoffInitialInterval            = 100 * time.Millisecond
	defaultBackoffRandomizationFactor        = 0.2
	defaultBackoffMultiplier                 = 1.6
	defaultBackoffMaxInterval                = time.Second
	defaultBackoffMaxRetries          uint64 = 10

	errCancelled = errors.New("retry is cancelled")
)

type controllerOptions struct {
	syncInterval time.Duration
}

func defaultControllerOptions() *controllerOptions {
	return &controllerOptions{
		syncInterval: time.Second * 10,
	}
}

type controllerOption func(o *controllerOptions)

func SyncInterval(interval time.Duration) controllerOption {
	return func(o *controllerOptions) {
		o.syncInterval = interval
	}
}

// Controller is used to store configuration information.
type Controller struct {
	sync.Mutex
	options *controllerOptions
	store   Store

	cache    atomic.Value // *Config
	updateCh chan struct{}
	evtHdls  atomic.Value //[]EventHandler

	dep      *DependenciesController
	inst     *InstancesController
	proxycfg *ProxyConfigsController

	initFinish bool
	stop       chan struct{}
	wg         sync.WaitGroup
}

// NewController return a new Controller.
func NewController(store Store, opts ...controllerOption) *Controller {
	o := defaultControllerOptions()
	for _, opt := range opts {
		opt(o)
	}

	c := &Controller{
		store:    store,
		options:  o,
		updateCh: make(chan struct{}, 1),
		stop:     make(chan struct{}),
		cache:    atomic.Value{},
	}
	c.dep = newDependenciesController(c)
	c.inst = newInstancesController(c)
	c.proxycfg = newProxyConfigController(c)
	return c
}

func (c *Controller) Dependencies() *DependenciesController {
	return c.dep
}

func (c *Controller) Instances() *InstancesController {
	return c.inst
}

func (c *Controller) ProxyConfigs() *ProxyConfigsController {
	return c.proxycfg
}

func (c *Controller) loadCache() *Cache {
	cache, _ := c.cache.Load().(*Cache)
	return cache
}

func (c *Controller) storeCache(cfg *Cache) {
	c.cache.Store(cfg)
}

func (c *Controller) loadEvtHdls() []EventHandler {
	hdls, _ := c.evtHdls.Load().([]EventHandler)
	return hdls
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

func (c *Controller) doRetry(fn func() (interface{}, error)) (res interface{}, err error) {
	b := utils.NewExponentialBackoffBuilder().
		InitialInterval(defaultBackoffInitialInterval).
		RandomizationFactor(defaultBackoffRandomizationFactor).
		Multiplier(defaultBackoffMultiplier).
		MaxInterval(defaultBackoffMaxInterval).
		MaxRetries(defaultBackoffMaxRetries).
		Build()
	for {
		res, err = fn()
		if err == nil {
			return
		}
		d := b.NextBackOff()
		if d == backoff.Stop {
			break
		}
		select {
		case <-c.stop:
			return nil, errCancelled
		case <-time.NewTimer(d).C:
		}
	}
	return
}

func (c *Controller) getKeysWithRetry(ns, typ string) (keys []string, err error) {
	res, err := c.doRetry(func() (i interface{}, e error) {
		i, e = c.store.GetKeys(ns, typ)
		switch e {
		case ErrNotExist:
			return nil, nil
		default:
			return
		}
	})
	if err != nil {
		return nil, err
	}
	keys, _ = res.([]string)
	return
}

func (c *Controller) getValueWithRetry(ns, typ, key string) ([]byte, *Metadata, error) {
	res, err := c.doRetry(func() (i interface{}, e error) {
		b, meta, e := c.store.Get(ns, typ, key)
		switch e {
		case nil:
			return [2]interface{}{b, meta}, nil
		case ErrNotExist:
			return nil, nil
		default:
			return nil, e
		}
	})
	if err != nil || res == nil {
		return nil, nil, err
	}
	value, _ := res.([2]interface{})[0].([]byte)
	metadata, _ := res.([2]interface{})[1].(*Metadata)
	return value, metadata, nil
}

func (c *Controller) fetchAll() (*Cache, error) {
	cache := NewCache()
	for ns, types := range InterestedNSAndType {
		for _, typ := range types {
			keys, err := c.getKeysWithRetry(ns, typ)
			if err != nil {
				return nil, err
			}
			if keys == nil {
				continue
			}
			for _, key := range keys {
				value, _, err := c.getValueWithRetry(ns, typ, key)
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
	for ns := range InterestedNSAndType {
		if err := c.trySubscribe(ns); err != nil {
			return err
		}
	}
	c.wg.Add(2)
	go c.triggerLoop()
	go c.loop()
	c.triggerUpdate()
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

func (c *Controller) triggerLoop() {
	ticker := time.NewTicker(c.options.syncInterval)
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

func (c *Controller) diffCache(that *Cache) {
	cur := c.loadCache()
	if cur == nil {
		cur = NewCache()
	}

	add, update, del := cur.Diff(that)
	dispatchEvent := func(event *Event) {
		for _, hdl := range c.loadEvtHdls() {
			hdl(event)
		}
	}
	for _, cfg := range add {
		dispatchEvent(NewEvent(EventAdd, cfg))
	}
	for _, cfg := range update {
		dispatchEvent(NewEvent(EventUpdate, cfg))
	}
	for _, cfg := range del {
		dispatchEvent(NewEvent(EventDelete, cfg))
	}
}

func (c *Controller) loop() {
	defer c.wg.Done()
	for {
		select {
		case <-c.stop:
			return
		case <-c.updateCh:
			newConf, err := c.fetchAll()
			if err != nil {
				logger.Warnf("failed to load config, err: %v", err)
				continue
			}
			c.diffCache(newConf)
			c.storeCache(newConf)
		}
	}
}

// RegisterEventHandler registers a handler to handle config event.
func (c *Controller) RegisterEventHandler(handler EventHandler) {
	c.Lock()
	defer c.Unlock()
	oldEvtHdls := c.loadEvtHdls()
	newEvtHdls := make([]EventHandler, len(oldEvtHdls), len(oldEvtHdls)+1)
	copy(newEvtHdls, oldEvtHdls)
	newEvtHdls = append(newEvtHdls, handler)
	c.evtHdls.Store(newEvtHdls)
}

// Get return config data by namespace, type and key.
func (c *Controller) Get(namespace, typ, key string) ([]byte, *Metadata, error) {
	return c.store.Get(namespace, typ, key)
}

// GetCache return config data by namespace, type and key from cache.
func (c *Controller) GetCache(namespace, typ, key string) ([]byte, error) {
	return c.loadCache().Get(namespace, typ, key)
}

// Add add config data by namespace, type and key.
func (c *Controller) Add(namespace, typ, key string, value []byte) error {
	return c.store.Add(namespace, typ, key, value)
}

// Update update config data by namespace, type and key.
func (c *Controller) Update(namespace, typ, key string, value []byte) error {
	return c.store.Update(namespace, typ, key, value)
}

// Del del config data by namespace, type and key.
func (c *Controller) Del(namespace, typ, key string) error {
	return c.store.Del(namespace, typ, key)
}

// Exist return true if config data is exist.
func (c *Controller) Exist(namespace, typ, key string) bool {
	_, _, err := c.Get(namespace, typ, key)
	return err == nil
}

// Keys return all key by namespace and type.
func (c *Controller) Keys(namespace, typ string) ([]string, error) {
	return c.store.GetKeys(namespace, typ)
}

// Keys return all key by namespace and type from cache.
func (c *Controller) KeysCached(namespace, typ string) ([]string, error) {
	return c.loadCache().Keys(namespace, typ)
}

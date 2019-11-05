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
	"github.com/samaritan-proxy/sash/model"
)

const (
	NamespaceService       = "service"
	TypeServiceProxyConfig = "proxy-config"
	TypeServiceDependence  = "dependence"

	svcDependenceSep = byte(',')
)

var interestedNSAndType = map[string][]string{
	NamespaceService: {TypeServiceProxyConfig, TypeServiceDependence},
}

// Controller is used to store configuration information.
type Controller struct {
	store    Store
	updateCh chan struct{}
	cache    atomic.Value // map[uint32]*rawConf
	interval time.Duration

	// for test
	cfgUpdateHdl func(namespace, typ, key string, value []byte)
	cfgAddHdl    func(namespace, typ, key string, value []byte)
	cfgDelHdl    func(namespace, typ, key string)

	svcCfgEvtHdls []EventHandler

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
	c.cfgUpdateHdl = c.handleCfgUpdate
	c.cfgAddHdl = c.handleCfgAdd
	c.cfgDelHdl = c.handleCfgDel
	return c
}

func (c *Controller) loadCache() map[uint32]*rawConf {
	return c.cache.Load().(map[uint32]*rawConf)
}

func (c *Controller) storeCache(m map[uint32]*rawConf) {
	c.cache.Store(m)
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

func (c *Controller) fetchAll() (map[uint32]*rawConf, error) {
	res := make(map[uint32]*rawConf)
	for ns, types := range interestedNSAndType {
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
				node := newRawConf(ns, typ, key, value)
				res[node.Hashcode()] = node
			}
		}
	}
	return res, nil
}

// Start start the controller.
func (c *Controller) Start() error {
	if err := c.store.Start(); err != nil {
		return err
	}
	// init Controller
	conf, err := c.fetchAll()
	if err != nil {
		return err
	}
	c.storeCache(conf)
	for ns := range interestedNSAndType {
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

func (c *Controller) handleCfgUpdate(namespace, typ, key string, value []byte) {
	if namespace == NamespaceService && typ == TypeServiceProxyConfig {
		cfg, err := unmarshalSvcConfig(value)
		if err != nil {
			logger.Warnf("failed to unmarshal svc config, err: %v", err)
			return
		}
		evt := NewSvcConfigEvent(EventUpdate, model.NewServiceConfig(key, cfg))
		for _, hdl := range c.svcCfgEvtHdls {
			hdl(evt)
		}
	}
}

func (c *Controller) handleCfgAdd(namespace, typ, key string, value []byte) {
	if namespace == NamespaceService && typ == TypeServiceProxyConfig {
		cfg, err := unmarshalSvcConfig(value)
		if err != nil {
			logger.Warnf("failed to unmarshal svc config, err: %v", err)
			return
		}
		evt := NewSvcConfigEvent(EventAdd, model.NewServiceConfig(key, cfg))
		for _, hdl := range c.svcCfgEvtHdls {
			hdl(evt)
		}
	}
}

func (c *Controller) handleCfgDel(namespace, typ, key string) {
	if namespace == NamespaceService && typ == TypeServiceProxyConfig {
		evt := NewSvcConfigEvent(EventDelete, model.NewServiceConfig(key, nil))
		for _, hdl := range c.svcCfgEvtHdls {
			hdl(evt)
		}
	}
}

func (c *Controller) diff(newConf map[uint32]*rawConf) {
	allKeys := make(map[uint32]struct{})
	for k := range c.loadCache() {
		allKeys[k] = struct{}{}
	}
	for k := range newConf {
		allKeys[k] = struct{}{}
	}
	for k := range allKeys {
		newConf, newConfExist := newConf[k]
		oldConf, oldConfExist := c.loadCache()[k]
		switch {
		case newConfExist && oldConfExist:
			if oldConf.Equal(newConf) {
				continue
			}
			c.cfgUpdateHdl(newConf.Namespace, newConf.Type, newConf.Key, newConf.Value)
		case newConfExist && !oldConfExist: //Add
			c.cfgAddHdl(newConf.Namespace, newConf.Type, newConf.Key, newConf.Value)
		case !newConfExist && oldConfExist: // Remove
			c.cfgDelHdl(oldConf.Namespace, oldConf.Type, oldConf.Key)
		}
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
				logger.Warnf("failed to fetch config from store, err: %v", err)
				continue
			}
			c.diff(newConf)
			c.storeCache(newConf)
		}
	}
}

// RegisterEventHandler registers a handler to handle config event.
// It is not goroutine-safe, should call it before execute Run.
func (c *Controller) RegisterEventHandler(handler EventHandler) {
	c.svcCfgEvtHdls = append(c.svcCfgEvtHdls, handler)
}

func (c *Controller) get(namespace, typ, key string) ([]byte, error) {
	// TODO: build and use index
	hashcode := newRawConf(namespace, typ, key, nil).Hashcode()
	node, ok := c.loadCache()[hashcode]
	if !ok {
		return nil, ErrKeyNotExist
	}
	return node.Value, nil
}

func (c *Controller) set(namespace, typ, key string, value []byte) error {
	return c.store.Set(namespace, typ, key, value)
}

func (c *Controller) del(namespace, typ, key string) error {
	return c.store.Del(namespace, typ, key)
}

func (c *Controller) exist(namespace, typ, key string) bool {
	// TODO: build and use index
	hashcode := newRawConf(namespace, typ, key, nil).Hashcode()
	_, ok := c.loadCache()[hashcode]
	return ok
}

// GetSvcDependence return the service config of the input service.
func (c *Controller) GetSvcConfig(service string) (*model.ServiceConfig, error) {
	b, err := c.get(NamespaceService, TypeServiceProxyConfig, service)
	switch err {
	case nil:
	case ErrNamespaceNotExist,
		ErrTypeNotExist,
		ErrKeyNotExist:
		// TODO: return default config
	default:
		return nil, err
	}
	cfg, err := unmarshalSvcConfig(b)
	if err != nil {
		logger.Warnf("failed to get svc config, err: %v", err)
		// TODO: return default config
	}

	return model.NewServiceConfig(service, cfg), nil
}

// GetSvcDependence return all dependencies of input service.
func (c *Controller) GetSvcDependence(service string) (*model.ServiceDependence, error) {
	b, err := c.get(NamespaceService, TypeServiceDependence, service)
	switch err {
	case nil:
	case ErrNamespaceNotExist,
		ErrTypeNotExist,
		ErrKeyNotExist:
		return model.NewServiceDependence(service, nil), nil
	default:
		return nil, err
	}

	deps, err := unmarshalSvcDependence(b)
	if err != nil {
		return nil, err
	}

	return model.NewServiceDependence(service, deps), nil
}

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

type config struct {
	Namespaces map[string]*namespace
}

type namespace struct {
	Name  string
	Types map[string]*typ
}

type typ struct {
	Name    string
	Configs map[string]*RawConf
}

// Controller is used to store configuration information.
type Controller struct {
	sync.Mutex
	store    Store
	updateCh chan struct{}

	index    atomic.Value // *config
	cache    atomic.Value // map[uint32]*rawConf
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

func (c *Controller) updateCacheAndIndex(m map[uint32]*RawConf) {
	c.Lock()
	defer c.Unlock()
	c.storeCache(m)
	c.storeIndex(m)
}

func (c *Controller) loadCache() map[uint32]*RawConf {
	return c.cache.Load().(map[uint32]*RawConf)
}

func (c *Controller) storeCache(m map[uint32]*RawConf) {
	c.cache.Store(m)
}

func (c *Controller) loadIndex() *config {
	return c.index.Load().(*config)
}

func (c *Controller) storeIndex(m map[uint32]*RawConf) {
	index := &config{
		Namespaces: make(map[string]*namespace),
	}
	for _, config := range m {
		if _, ok := index.Namespaces[config.Namespace]; !ok {
			index.Namespaces[config.Namespace] = &namespace{
				Name:  config.Namespace,
				Types: make(map[string]*typ),
			}
		}
		ns := index.Namespaces[config.Namespace]
		if _, ok := ns.Types[config.Type]; !ok {
			ns.Types[config.Type] = &typ{
				Name:    config.Type,
				Configs: make(map[string]*RawConf),
			}
		}
		typ := ns.Types[config.Type]
		typ.Configs[config.Key] = config
	}
	c.index.Store(index)
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

func (c *Controller) fetchAll() (map[uint32]*RawConf, error) {
	res := make(map[uint32]*RawConf)
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
				node := NewRawConf(ns, typ, key, value)
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
	c.updateCacheAndIndex(conf)
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

func (c *Controller) diff(newConf map[uint32]*RawConf) {
	allKeys := make(map[uint32]struct{})
	for k := range c.loadCache() {
		allKeys[k] = struct{}{}
	}
	for k := range newConf {
		allKeys[k] = struct{}{}
	}
	for k := range allKeys {
		var (
			newConf, newConfExist = newConf[k]
			oldConf, oldConfExist = c.loadCache()[k]

			evt *Event
		)

		switch {
		case newConfExist && oldConfExist:
			if oldConf.Equal(newConf) {
				continue
			}
			evt = NewEvent(EventUpdate, newConf)
		case newConfExist && !oldConfExist: //Add
			evt = NewEvent(EventAdd, newConf)
		case !newConfExist && oldConfExist: // Remove
			evt = NewEvent(EventDelete, oldConf)
		}

		for _, hdl := range c.evtHdls {
			hdl(evt)
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
	c.updateCacheAndIndex(newConf)
}

// RegisterEventHandler registers a handler to handle config event.
// It is not goroutine-safe, should call it before execute Run.
func (c *Controller) RegisterEventHandler(handler EventHandler) {
	c.evtHdls = append(c.evtHdls, handler)
}

func (c *Controller) get(namespace, typ, key string) ([]byte, error) {
	index := c.loadIndex()
	ns, ok := index.Namespaces[namespace]
	if !ok {
		return nil, ErrNamespaceNotExist
	}
	t, ok := ns.Types[typ]
	if !ok {
		return nil, ErrTypeNotExist
	}
	v, ok := t.Configs[key]
	if !ok {
		return nil, ErrKeyNotExist
	}
	return v.Value, nil
}

func (c *Controller) Get(namespace, typ, key string) ([]byte, error) {
	return c.get(namespace, typ, key)
}

func (c *Controller) Set(namespace, typ, key string, value []byte) error {
	return c.store.Set(namespace, typ, key, value)
}

func (c *Controller) Del(namespace, typ, key string) error {
	return c.store.Del(namespace, typ, key)
}

func (c *Controller) Exist(namespace, typ, key string) bool {
	_, err := c.get(namespace, typ, key)
	return err == nil
}

func (c *Controller) Keys(namespace, typ string) ([]string, error) {
	index := c.loadIndex()
	ns, ok := index.Namespaces[namespace]
	if !ok {
		return nil, ErrNamespaceNotExist
	}
	t, ok := ns.Types[typ]
	if !ok {
		return nil, ErrTypeNotExist
	}
	keys := make([]string, 0, len(t.Configs))
	for k := range t.Configs {
		keys = append(keys, k)
	}
	return keys, nil
}

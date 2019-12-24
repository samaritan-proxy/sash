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

package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/samaritan-proxy/sash/api"
	"github.com/samaritan-proxy/sash/config"
	cfgmem "github.com/samaritan-proxy/sash/config/memory"
	cfgzk "github.com/samaritan-proxy/sash/config/zk"
	"github.com/samaritan-proxy/sash/discovery"
	"github.com/samaritan-proxy/sash/internal/zk"
	"github.com/samaritan-proxy/sash/logger"
	"github.com/samaritan-proxy/sash/model"
	"github.com/samaritan-proxy/sash/registry"
	regmem "github.com/samaritan-proxy/sash/registry/memory"
	regzk "github.com/samaritan-proxy/sash/registry/zk"
)

func init() {
	parseConfig()
}

var (
	signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	inst    *instance
)

type instance struct {
	wg     sync.WaitGroup
	ctx    context.Context
	cancel func()

	// FIXME: remove this
	SvcReg model.ServiceRegistry
	Reg    registry.Cache
	Cfg    *config.Controller
	Api    *api.Server
	Xds    *discovery.Server
}

func getDurationFromQuery(v url.Values, key string) (d time.Duration) {
	value := v.Get(key)
	if len(value) == 0 {
		return
	}
	dur, err := time.ParseDuration(value)
	if err != nil {
		return
	}
	return dur
}

func buildZKConfig(u *url.URL) *zk.ConnConfig {
	var (
		pwd, _         = u.User.Password()
		connectTimeout = getDurationFromQuery(u.Query(), "connect_timeout")
		sessionTimeout = getDurationFromQuery(u.Query(), "session_timeout")
	)

	return &zk.ConnConfig{
		Hosts:          strings.Split(u.Host, ","),
		User:           u.User.Username(),
		Pwd:            pwd,
		ConnectTimeout: connectTimeout,
		SessionTimeout: sessionTimeout,
	}
}

func buildServiceRegistry() (model.ServiceRegistry, error) {
	if len(rootCfg.Registry.Endpoint) == 0 {
		logger.Warnf("discovery will run in in-memory mode, only for test")
		return regmem.NewRegistry(), nil
	}
	u, err := url.Parse(rootCfg.Registry.Endpoint)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "zk":
		return regzk.NewDiscoveryClient(buildZKConfig(u), u.Path)
	default:
		return nil, fmt.Errorf("unknown type of scheme when init service registry: %s", u.Scheme)
	}
}

func buildConfigStore() (config.Store, error) {
	if len(rootCfg.Config.Endpoint) == 0 {
		logger.Warnf("config store will run in in-memory mode, only for test")
		return cfgmem.NewMemStore(), nil
	}
	u, err := url.Parse(rootCfg.Config.Endpoint)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "zk":
		return cfgzk.New(buildZKConfig(u), u.Path)
	default:
		return nil, fmt.Errorf("unknown type of scheme when init config store: %s", u.Scheme)
	}
}

func buildListener(addr string) (net.Listener, error) {
	return net.Listen("tcp", addr)
}

func initInstance() error {
	if inst != nil {
		return nil
	}

	serviceRegistry, err := buildServiceRegistry()
	if err != nil {
		return err
	}

	store, err := buildConfigStore()
	if err != nil {
		return err
	}

	ln, err := buildListener(rootCfg.XdsRPC.Bind)
	if err != nil {
		return err
	}

	var (
		ctx, cancel = context.WithCancel(context.TODO())
		reg         = registry.NewCache(
			serviceRegistry,
			registry.SyncFreq(rootCfg.Registry.SyncFreq),
			registry.SyncJitter(rootCfg.Registry.SyncJitter))
		ctl = config.NewController(store, config.Interval(rootCfg.Config.Interval))
		api = api.New(rootCfg.API.Bind, reg, ctl)
		xds = discovery.NewServer(ln, reg, ctl)
	)

	inst = &instance{
		ctx:    ctx,
		cancel: cancel,
		SvcReg: serviceRegistry,
		Reg:    reg,
		Cfg:    ctl,
		Api:    api,
		Xds:    xds,
	}
	return nil
}

func (i *instance) startComponent(fn func()) {
	i.wg.Add(1)
	go func() {
		defer i.wg.Done()
		fn()
	}()
}

func (i *instance) Start() error {
	i.startComponent(func() { i.SvcReg.Run(i.ctx) })
	i.startComponent(func() { i.Reg.Run(i.ctx) })
	if err := i.Cfg.Start(); err != nil {
		return err
	}
	logger.Infof("HTTP API serve at: %s", rootCfg.API.Bind)
	i.startComponent(func() {
		if err := i.Api.Start(); err != nil {
			logger.Fatal(err)
		}
	})
	logger.Infof("XDS RPC serve at: %s", rootCfg.XdsRPC.Bind)
	i.startComponent(func() {
		if err := i.Xds.Serve(); err != nil {
			logger.Fatal(err)
		}
	})
	return nil
}

func (i *instance) Stop() {
	i.cancel()
	i.Cfg.Stop()
	i.Api.Stop()
	i.Xds.Stop()
	i.wg.Wait()
}

func handleSignal() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, signals...)
	s := <-signalCh
	logger.Info("Signal received: ", s)
	inst.Stop()
}

func main() {
	logger.SetLevel(rootCfg.LogLevel)
	logger.Debugf("rootConfig: %+v", rootCfg)
	if err := initInstance(); err != nil {
		logger.Fatal(err)
	}
	if err := inst.Start(); err != nil {
		logger.Fatal(err)
	}
	handleSignal()
}

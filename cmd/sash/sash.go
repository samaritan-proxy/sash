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
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	parseFlags()
}

var (
	signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
)

func newServiceRegistry() (model.ServiceRegistry, error) {
	switch typ := bootstrap.Registry.Type; typ {
	case "memory":
		logger.Warnf("discovery will run in in-memory mode, only for test")
		return regmem.NewRegistry(), nil
	case "zk":
		return regzk.NewDiscoveryClient(bootstrap.Registry.Spec.(*zk.ConnConfig))
	default:
		return nil, fmt.Errorf("unknown type of service registry: %s", typ)
	}
}

func newConfigStore() (config.Store, error) {
	switch typ := bootstrap.ConfigStore.Type; typ {
	case "memory":
		logger.Warnf("config store will run in in-memory mode, only for test")
		return cfgmem.NewMemStore(), nil
	case "zk":
		return cfgzk.New(bootstrap.Registry.Spec.(*zk.ConnConfig))
	default:
		return nil, fmt.Errorf("unknown type of config store: %s", typ)
	}
}

func newListener(addr string) (net.Listener, error) {
	return net.Listen("tcp", addr)
}

func main() {
	logger.SetLevel(bootstrap.LogLevel)
	logger.Debugf("rootConfig: %+v", bootstrap)

	wg := sync.WaitGroup{}

	serviceRegistry, err := newServiceRegistry()
	if err != nil {
		logger.Fatal(err)
	}

	store, err := newConfigStore()
	if err != nil {
		logger.Fatal(err)
	}

	ln, err := newListener(bootstrap.Discovery.Bind)
	if err != nil {
		logger.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.TODO())

	reg := registry.NewCache(
		serviceRegistry,
		registry.SyncFreq(bootstrap.Registry.SyncFreq),
		registry.SyncJitter(bootstrap.Registry.SyncJitter))
	ctl := config.NewController(store, config.Interval(bootstrap.ConfigStore.SyncFreq))
	api := api.New(bootstrap.API.Bind, reg, ctl)
	rpc := discovery.NewServer(ln, reg, ctl)

	wg.Add(1)
	go func() {
		defer wg.Done()
		serviceRegistry.Run(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		reg.Run(ctx)
	}()

	if err := ctl.Start(); err != nil {
		logger.Fatal(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Infof("HTTP API serve at: %s", bootstrap.API.Bind)
		if err := api.Start(); err != nil {
			logger.Fatal(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Infof("Discovery RPC serve at: %s", bootstrap.Discovery.Bind)
		if err := rpc.Serve(); err != nil {
			logger.Fatal(err)
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, signals...)
	s := <-signalCh
	logger.Info("Signal received: ", s)

	cancel()
	api.Stop()
	rpc.Stop()
	ctl.Stop()

	wg.Wait()
}

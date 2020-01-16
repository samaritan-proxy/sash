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
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/samaritan-proxy/sash/api"
	"github.com/samaritan-proxy/sash/config"
	"github.com/samaritan-proxy/sash/discovery"
	"github.com/samaritan-proxy/sash/logger"
	"github.com/samaritan-proxy/sash/registry"
)

var (
	b = &Bootstrap{
		ConfigStore: ConfigStore{
			SyncFreq: time.Second * 5,
		},
		Registry: Registry{
			SyncFreq:   time.Second * 5,
			SyncJitter: 0.1,
		},
		API:       API{Bind: ":8882"},
		Discovery: Discovery{Bind: ":9090"},
		LogLevel:  "info",
	}

	configFile string
)

func parseFlags() {
	flag.StringVar(&configFile, "c", "./config.yaml", "config file")
	flag.Parse()

	f, err := os.Open(configFile)
	if err != nil {
		logger.Fatal(err)
	}

	if err = yaml.NewDecoder(f).Decode(&b); err != nil {
		logger.Fatal(err)
	}

	// TODO: verify and fix the bootstrap info.
}

func init() {
	parseFlags()
}

func initDiscoveryServer(b *Bootstrap, reg registry.Cache, cfg *config.Controller) *discovery.Server {
	l, err := net.Listen("tcp", b.Discovery.Bind)
	if err != nil {
		log.Fatal(err)
	}
	s := discovery.NewServer(l, reg, cfg)
	return s
}

func initAPIServer(b *Bootstrap, reg registry.Cache, cfg *config.Controller) *api.Server {
	l, err := net.Listen("tcp", b.API.Bind)
	if err != nil {
		log.Fatal(err)
	}
	s := api.New(l, reg, cfg)
	return s
}

func main() {
	logger.SetLevel(b.LogLevel)
	// TODO: make the print info more pretty
	logger.Debugf("bootstrap: %+v", b)

	regCtl := initRegistryController(b)
	cfgCtl := initConfigController(b)
	ds := initDiscoveryServer(b, regCtl, cfgCtl)
	as := initAPIServer(b, regCtl, cfgCtl)
	ctx, cancel := context.WithCancel(context.Background())

	if err := cfgCtl.Start(); err != nil {
		log.Fatal(err)
	}
	defer cfgCtl.Stop()
	go regCtl.Run(ctx)
	go ds.Serve()
	go as.Serve()
	defer func() {
		ds.Stop()
		as.Shutdown()
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	s := <-signalCh
	logger.Info("Signal received: ", s)
	cancel()
}

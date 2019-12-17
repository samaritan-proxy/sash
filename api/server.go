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

package api

import (
	"context"
	"net/http"
	"time"

	"github.com/samaritan-proxy/sash/config"
	"github.com/samaritan-proxy/sash/logger"
	"github.com/samaritan-proxy/sash/registry"
)

type serverOptions struct {
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type ServerOption func(o *serverOptions)

func Addr(addr string) ServerOption {
	return func(o *serverOptions) {
		o.Addr = addr
	}
}

func ReadTimeout(d time.Duration) ServerOption {
	return func(o *serverOptions) {
		o.ReadTimeout = d
	}
}

func ReadHeaderTimeout(d time.Duration) ServerOption {
	return func(o *serverOptions) {
		o.ReadHeaderTimeout = d
	}
}

func WriteTimeout(d time.Duration) ServerOption {
	return func(o *serverOptions) {
		o.WriteTimeout = d
	}
}

func IdleTimeout(d time.Duration) ServerOption {
	return func(o *serverOptions) {
		o.IdleTimeout = d
	}
}

type Server struct {
	hs      *http.Server
	options *serverOptions

	reg registry.Cache

	rawCtl      *config.Controller
	depsCtl     *config.DependenciesController
	proxyCfgCtl *config.ProxyConfigsController
	instCtl     *config.InstancesController
}

func New(reg registry.Cache, ctl *config.Controller, opts ...ServerOption) *Server {
	options := new(serverOptions)
	for _, opt := range opts {
		opt(options)
	}
	s := &Server{
		reg:         reg,
		rawCtl:      ctl,
		depsCtl:     ctl.Dependencies(),
		proxyCfgCtl: ctl.ProxyConfigs(),
		instCtl:     ctl.Instances(),
		options:     options,
		hs: &http.Server{
			Addr:              options.Addr,
			ReadTimeout:       options.ReadTimeout,
			ReadHeaderTimeout: options.ReadHeaderTimeout,
			WriteTimeout:      options.WriteTimeout,
			IdleTimeout:       options.IdleTimeout,
		},
	}
	s.hs.Handler = s.genRouter()
	return s
}

func (s *Server) Start() error {
	go func() {
		switch err := s.hs.ListenAndServe(); err {
		case nil, http.ErrServerClosed:
		default:
			logger.Warnf("http.Server.ListenAndServe got a unexpected error: %s")
		}
	}()
	return nil
}

func (s *Server) Stop() {
	//nolint:lostcancel
	ctx, _ := context.WithTimeout(context.TODO(), time.Second)
	if err := s.hs.Shutdown(ctx); err != nil {
		logger.Warnf("failed to stop server, err: %v", err)
	}
}

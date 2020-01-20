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
	"net"
	"net/http"
	"time"

	"github.com/samaritan-proxy/sash/config"
	"github.com/samaritan-proxy/sash/logger"
	"github.com/samaritan-proxy/sash/registry"
)

type serverOptions struct {
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type ServerOption func(o *serverOptions)

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
	l       net.Listener
	hs      *http.Server
	options *serverOptions

	reg         registry.Cache
	rawCtl      *config.Controller
	depsCtl     *config.DependenciesController
	proxyCfgCtl *config.ProxyConfigsController
	instCtl     *config.InstancesController
}

func New(l net.Listener, reg registry.Cache, ctl *config.Controller, opts ...ServerOption) *Server {
	options := new(serverOptions)
	for _, opt := range opts {
		opt(options)
	}
	s := &Server{
		l:           l,
		reg:         reg,
		rawCtl:      ctl,
		depsCtl:     ctl.Dependencies(),
		proxyCfgCtl: ctl.ProxyConfigs(),
		instCtl:     ctl.Instances(),
		options:     options,
		hs: &http.Server{
			ReadTimeout:       options.ReadTimeout,
			ReadHeaderTimeout: options.ReadHeaderTimeout,
			WriteTimeout:      options.WriteTimeout,
			IdleTimeout:       options.IdleTimeout,
		},
	}
	s.hs.Handler = s.genRouter()
	return s
}

func (s *Server) Addr() string {
	return s.l.Addr().String()
}

func (s *Server) Serve() error {
	logger.Infof("API server listening on %s...", s.Addr())
	switch err := s.hs.Serve(s.l); err {
	case nil, http.ErrServerClosed:
		return nil
	default:
		logger.Warnf("http.Server.ListenAndServe got a unexpected error: %s")
		return err
	}
}

func (s *Server) Shutdown() {
	ctx, _ := context.WithTimeout(context.TODO(), time.Second) //nolint:lostcancel
	if err := s.hs.Shutdown(ctx); err != nil {
		logger.Warnf("Error when shutdowning the api server: %v", err)
	}
}

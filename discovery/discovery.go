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

package discovery

import (
	"net"
	"time"

	"github.com/samaritan-proxy/samaritan-api/go/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/samaritan-proxy/sash/config"
	"github.com/samaritan-proxy/sash/logger"
	"github.com/samaritan-proxy/sash/registry"
)

type serverOptions struct {
	// TODO: add fields, such as credentials.
}

func defaultServerOptions() *serverOptions {
	return &serverOptions{}
}

type ServerOption func(o *serverOptions)

// Server is an implementation of api.DiscoveryServiceServer.
type Server struct {
	l       net.Listener
	options *serverOptions

	g *grpc.Server

	eds *endpointDiscoveryServer
	cds *configDiscoveryServer
	dds *dependencyDiscoveryServer
}

// NewServer creates a discovery server.
func NewServer(l net.Listener, reg registry.Cache, ctl *config.Controller, opts ...ServerOption) *Server {
	o := defaultServerOptions()
	for _, opt := range opts {
		opt(o)
	}

	eds := newEndpointDiscoveryServer(reg)
	cds := newConfigDiscoveryServer(ctl)
	dds := newDependencyDiscoveryServer(ctl)
	s := &Server{
		l:       l,
		options: o,
		eds:     eds,
		cds:     cds,
		dds:     dds,
	}

	g := grpc.NewServer(s.grpcOptions()...)
	api.RegisterDiscoveryServiceServer(g, s)
	s.g = g
	return s
}

func (s *Server) grpcOptions() []grpc.ServerOption {
	options := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    30 * time.Second,
			Timeout: 10 * time.Second,
		}),
	}
	return options
}

func (s *Server) Serve() error {
	logger.Infof("Discovery server listening on %s...", s.l.Addr())
	return s.g.Serve(s.l)
}

// Stop stops the server.
func (s *Server) Stop() {
	s.g.Stop()
}

// StreamDependencies returns all dependencies of the given instance.
func (s *Server) StreamDependencies(req *api.DependencyDiscoveryRequest, stream api.DiscoveryService_StreamDependenciesServer) (err error) {
	return s.dds.StreamDependencies(req, stream)
}

// StreamSvcConfigs receives a stream of service subscription/unsubscription, and responds with a stream
// of the updated service configs.
func (s *Server) StreamSvcConfigs(stream api.DiscoveryService_StreamSvcConfigsServer) (err error) {
	return s.cds.StreamSvcConfigs(stream)
}

// StreamSvcEndpoints receives a stream of service subscription/unsubscription, and responds with a stream
// of the changed service endpoints.
func (s *Server) StreamSvcEndpoints(stream api.DiscoveryService_StreamSvcEndpointsServer) (err error) {
	return s.eds.StreamSvcEndpoints(stream)
}

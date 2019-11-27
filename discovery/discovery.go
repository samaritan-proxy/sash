package discovery

import (
	"net"
	"time"

	"github.com/samaritan-proxy/samaritan-api/go/api"
	"github.com/samaritan-proxy/sash/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
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
	addr    string
	options *serverOptions

	g   *grpc.Server
	eds *endpointDiscoveryServer
}

// NewServer creates a discovery server.
func NewServer(addr string, reg registry.Cache, opts ...ServerOption) *Server {
	o := defaultServerOptions()
	for _, opt := range opts {
		opt(o)
	}

	eds := newEndpointDiscoveryServer(reg)
	s := &Server{
		addr:    addr,
		options: o,
		eds:     eds,
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
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	return s.g.Serve(l)
}

// Stop stops the server.
func (s *Server) Stop() {
	s.g.Stop()
}

// StreamSvcs returns all dependencies of the given isntance.
func (s *Server) StreamSvcs(req *api.SvcDiscoveryRequest, stream api.DiscoveryService_StreamSvcsServer) (err error) {
	// TODO: implement it
	return
}

// StreamSvcConfigs receives a stream of service subscription/unsubscription, and responds with a stream
// of the updated service configs.
func (s *Server) StreamSvcConfigs(stream api.DiscoveryService_StreamSvcConfigsServer) (err error) {
	// TODO: implement it
	return
}

// StreamSvcEndpoints receives a stream of service subscription/unsubscription, and responds with a stream
// of the changed service ednpoints.
func (s *Server) StreamSvcEndpoints(stream api.DiscoveryService_StreamSvcEndpointsServer) (err error) {
	return s.eds.StreamSvcEndpoints(stream)
}

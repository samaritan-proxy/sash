package discovery

import (
	"sync"

	"github.com/samaritan-proxy/samaritan-api/go/api"

	"github.com/samaritan-proxy/sash/registry"
)

type endpointClient struct {
	s *endpointServer

	// subscribed service names
	subscribed map[string]interface{}
}

func newEndpointClient(s *endpointServer) *endpointClient {
	return &endpointClient{
		s:          s,
		subscribed: make(map[string]interface{}, 8),
	}
}

func (c *endpointClient) Subscribe(svcNames ...string) {
	for _, svcName := range svcNames {
		_, ok := c.subscribed[svcName]
		if ok {
			continue
		}

		c.subscribed[svcName] = struct{}{}
		c.s.Subscribe(svcName, c)
	}
}

func (c *endpointClient) Unsubscribe(svcNames ...string) {
	for _, svcName := range svcNames {
		_, ok := c.subscribed[svcName]
		if !ok {
			continue
		}
		c.unsubscribe(svcName)
	}
}

func (c *endpointClient) UnsubscribeAll() {
	for svcName := range c.subscribed {
		c.unsubscribe(svcName)
	}
}

func (c *endpointClient) unsubscribe(svcName string) {
	c.s.Unsubscribe(svcName, c)
	delete(c.subscribed, svcName)
}

type endpointClients map[*endpointClient]interface{}

type endpointServer struct {
	sync.RWMutex
	reg registry.Cache

	subscribers map[string]endpointClients // service_name: endpoint_clients
}

func (s *endpointServer) StreamSvcEndpoints(stream api.DiscoveryService_StreamSvcEndpointsServer) error {
	c := newEndpointClient(s)

	go func() {
		for {
			req, err := stream.Recv()
			if err != nil {
				return
			}

			c.Subscribe(req.SvcNamesSubscribe...)
			c.Unsubscribe(req.SvcNamesUnsubscribe...)
		}
	}()

	return nil
}

func (s *endpointServer) Subscribe(svcName string, c *endpointClient) {
	s.Lock()
	defer s.Unlock()
	subscribers, ok := s.subscribers[svcName]
	if !ok {
		subscribers = endpointClients{}
		s.subscribers[svcName] = subscribers
	}
	subscribers[c] = struct{}{}
}

func (s *endpointServer) Unsubscribe(svcName string, c *endpointClient) {
	s.Lock()
	defer s.Unlock()
	// Go runtime never shrink map after elemenets removal, refer to: https://github.com/golang/go/issues/20135
	// FIXME: To prevent OOM after long running, we should add some memchainsm recycle the memory.
	subscribers, ok := s.subscribers[svcName]
	if !ok {
		return
	}
	delete(subscribers, c)
}

package discovery

import (
	"sync"

	"github.com/samaritan-proxy/samaritan-api/go/api"
	"github.com/samaritan-proxy/samaritan-api/go/common"
	"github.com/samaritan-proxy/samaritan-api/go/config/service"

	"github.com/samaritan-proxy/sash/model"
	"github.com/samaritan-proxy/sash/registry"
)

type endpointEvent struct {
	SvcName string
	Added   []*service.Endpoint
	Updated []*service.Endpoint
	Removed []*service.Endpoint
}

func newEndpointEvent(svcName string, added, updated, removed []*model.ServiceInstance) *endpointEvent {
	event := &endpointEvent{
		SvcName: svcName,
		Added:   make([]*service.Endpoint, len(added)),
		Updated: make([]*service.Endpoint, len(updated)),
		Removed: make([]*service.Endpoint, len(removed)),
	}

	for i, inst := range added {
		event.Added[i] = toEndpoint(inst)
	}
	for i, inst := range updated {
		event.Updated[i] = toEndpoint(inst)
	}
	for i, inst := range removed {
		event.Removed[i] = toEndpoint(inst)
	}

	return event
}

func toEndpoint(inst *model.ServiceInstance) *service.Endpoint {
	endpoint := &service.Endpoint{
		Address: &common.Address{
			Ip:   "",
			Port: 0,
		},
		// TODO: attach type and state
	}
	return endpoint
}

type endpointClient struct {
	s      *endpointServer
	stream api.DiscoveryService_StreamSvcEndpointsServer

	subscribed map[string]struct{} // subscribed services.
	eventCh    chan *endpointEvent

	quit chan struct{}
}

func newEndpointClient(s *endpointServer, stream api.DiscoveryService_StreamSvcEndpointsServer) *endpointClient {
	return &endpointClient{
		s:          s,
		stream:     stream,
		subscribed: make(map[string]struct{}, 8),
		eventCh:    make(chan *endpointEvent, 64),
		quit:       make(chan struct{}),
	}
}

func (c *endpointClient) Serve() {
	recvDone := make(chan struct{})
	defer func() {
		close(c.quit)
		// wait recv goroutine done
		<-recvDone
		// unsubscribe all the services.
		c.unsubscribeAll()
	}()

	go func() {
		defer close(recvDone)
		for {
			req, err := c.stream.Recv()
			if err != nil {
				// TODO: add log
				return
			}

			c.subscribe(req.SvcNamesSubscribe...)
			c.unsubscribe(req.SvcNamesUnsubscribe...)
		}
	}()

	var event *endpointEvent
	for {
		select {
		case event = <-c.eventCh:
		case <-recvDone:
			return
		}

		resp := &api.SvcEndpointDiscoveryResponse{
			SvcName: event.SvcName,
			Added:   event.Added,
			Removed: event.Removed,
		}
		if err := c.stream.Send(resp); err != nil {
			// TODO: add log
			return
		}
	}
}

func (c *endpointClient) subscribe(svcNames ...string) {
	for _, svcName := range svcNames {
		_, ok := c.subscribed[svcName]
		if ok {
			continue
		}

		c.subscribed[svcName] = struct{}{}
		c.s.Subscribe(svcName, c)
	}
}

func (c *endpointClient) unsubscribe(svcNames ...string) {
	for _, svcName := range svcNames {
		_, ok := c.subscribed[svcName]
		if !ok {
			continue
		}
		c.s.Unsubscribe(svcName, c)
		delete(c.subscribed, svcName)
	}
}

func (c *endpointClient) unsubscribeAll() {
	for svcName := range c.subscribed {
		c.unsubscribe(svcName)
	}
}

func (c *endpointClient) SendEvent(event *endpointEvent) {
	select {
	case c.eventCh <- event:
	case <-c.quit:
	}
}

func (c *endpointClient) Event() <-chan *endpointEvent {
	return c.eventCh
}

type endpointClients map[*endpointClient]interface{}

type endpointServer struct {
	sync.RWMutex
	reg registry.Cache

	subscribers map[string]endpointClients // service_name: endpoint_clients
}

func newEndpointServer(reg registry.Cache) *endpointServer {
	s := &endpointServer{
		reg:         reg,
		subscribers: make(map[string]endpointClients),
	}

	// register handler that handles registry event
	s.reg.RegisterServiceEventHandler(s.handleRegServiceEvent)
	s.reg.RegisterInstanceEventHandler(s.handleRegInstanceEvent)

	return s
}

func (s *endpointServer) handleRegServiceEvent(event *registry.ServiceEvent) {
	svc := event.Service

	// convert instances map to slice
	insts := make([]*model.ServiceInstance, 0, len(svc.Instances))
	for _, inst := range svc.Instances {
		insts = append(insts, inst)
	}

	var added, updated, deleted []*model.ServiceInstance
	switch event.Type {
	case registry.EventAdd:
		added = insts
	case registry.EventDelete:
		deleted = insts
	}

	svcName := svc.Name
	etEvent := newEndpointEvent(svcName, added, updated, deleted)
	s.dispatchEvent(etEvent)
}

func (s *endpointServer) handleRegInstanceEvent(event *registry.InstanceEvent) {
	var added, updated, deleted []*model.ServiceInstance
	switch event.Type {
	case registry.EventAdd:
		added = event.Instances
	case registry.EventUpdate:
		updated = event.Instances
	case registry.EventDelete:
		deleted = event.Instances
	}

	svcName := event.ServiceName
	etEvent := newEndpointEvent(svcName, added, updated, deleted)
	s.dispatchEvent(etEvent)
}

func (s *endpointServer) dispatchEvent(event *endpointEvent) {
	// TODO: improve map performance.
	s.RLock()
	defer s.RUnlock()
	subscribers := s.subscribers[event.SvcName]
	for subscriber := range subscribers {
		subscriber.SendEvent(event)
	}
}

func (s *endpointServer) Subscribe(svcName string, c *endpointClient) {
	s.Lock()
	defer s.Unlock()
	subscribers, ok := s.subscribers[svcName]
	if !ok {
		subscribers = endpointClients{}
		s.subscribers[svcName] = subscribers
	}

	if _, ok := subscribers[c]; ok {
		return
	}

	subscribers[c] = struct{}{}
	// send all instances when first subscribe
	svc, err := s.reg.Get(svcName)
	// error only can be service not found.
	if err != nil {
		return
	}

	// convert instances map to slice
	insts := make([]*model.ServiceInstance, 0, len(svc.Instances))
	for _, inst := range svc.Instances {
		insts = append(insts, inst)
	}

	event := newEndpointEvent(svc.Name, insts, nil, nil)
	c.SendEvent(event)
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

func (s *endpointServer) StreamSvcEndpoints(stream api.DiscoveryService_StreamSvcEndpointsServer) {
	c := newEndpointClient(s, stream)
	c.Serve()
}

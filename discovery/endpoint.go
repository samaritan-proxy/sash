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
	"sync"

	"github.com/samaritan-proxy/samaritan-api/go/api"
	"github.com/samaritan-proxy/samaritan-api/go/common"
	"github.com/samaritan-proxy/samaritan-api/go/config/service"
	"google.golang.org/grpc/peer"

	"github.com/samaritan-proxy/sash/logger"
	"github.com/samaritan-proxy/sash/model"
	"github.com/samaritan-proxy/sash/registry"
)

//go:generate mockgen -package $GOPACKAGE -self_package github.com/samaritan-proxy/sash/$GOPACKAGE --destination ./mock_endpoint_test.go github.com/samaritan-proxy/samaritan-api/go/api DiscoveryService_StreamSvcEndpointsServer

type endpointEvent struct {
	SvcName string
	Added   []*service.Endpoint
	Updated []*service.Endpoint
	Removed []*service.Endpoint
}

func newEndpointEvent(svcName string, added, updated, removed []*model.ServiceInstance) *endpointEvent {
	event := &endpointEvent{SvcName: svcName}

	if len(added) > 0 {
		event.Added = make([]*service.Endpoint, len(added))
	}
	if len(updated) > 0 {
		event.Updated = make([]*service.Endpoint, len(updated))
	}
	if len(removed) > 0 {
		event.Removed = make([]*service.Endpoint, len(removed))
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
			Ip:   inst.IP,
			Port: uint32(inst.Port),
		},
		// TODO: attach type and state
	}
	return endpoint
}

func instsMapToSlice(instsMap map[string]*model.ServiceInstance) []*model.ServiceInstance {
	insts := make([]*model.ServiceInstance, 0, len(instsMap))
	for _, inst := range instsMap {
		insts = append(insts, inst)
	}
	return insts
}

type (
	endpointSubHandler   func(svcName string, session *endpointDiscoverySession)
	endpointUnsubHandler func(svcName string, session *endpointDiscoverySession)
)

type endpointDiscoverySession struct {
	stream api.DiscoveryService_StreamSvcEndpointsServer
	remote *peer.Peer

	subscribed map[string]struct{} // subscribed services.
	subHdlr    endpointSubHandler
	unsubHdlr  endpointUnsubHandler
	eventCh    chan *endpointEvent

	quit chan struct{}
}

func newEndpointDiscoverySession(stream api.DiscoveryService_StreamSvcEndpointsServer) *endpointDiscoverySession {
	remote, _ := peer.FromContext(stream.Context())
	return &endpointDiscoverySession{
		stream:     stream,
		remote:     remote,
		subscribed: make(map[string]struct{}, 8),
		eventCh:    make(chan *endpointEvent, 64),
		quit:       make(chan struct{}),
	}
}

func (session *endpointDiscoverySession) SetSubscribeHandler(hdlr endpointSubHandler) {
	session.subHdlr = hdlr
}

func (session *endpointDiscoverySession) SetUnsubscribeHandler(hdlr endpointUnsubHandler) {
	session.unsubHdlr = hdlr
}

func (session *endpointDiscoverySession) Serve() {
	recvDone := make(chan struct{})
	defer func() {
		close(session.quit)
		// wait recv goroutine done
		<-recvDone
		// unsubscribe all the services.
		session.unsubscribeAll()
		logger.Debugf("Endpoint discovery session %s exit", session.remote.Addr)
	}()

	go func() {
		defer close(recvDone)
		for {
			req, err := session.stream.Recv()
			if err != nil {
				logger.Warnf("Read from service endpoints stream %s failed: %v", session.remote.Addr, err)
				return
			}

			session.subscribe(req.SvcNamesSubscribe...)
			session.unsubscribe(req.SvcNamesUnsubscribe...)
		}
	}()

	var event *endpointEvent
	for {
		select {
		case event = <-session.eventCh:
		case <-recvDone:
			return
		}

		resp := &api.SvcEndpointDiscoveryResponse{
			SvcName: event.SvcName,
			Added:   event.Added,
			Removed: event.Removed,
			// TODO: attach updated endpoints.
		}
		if err := session.stream.Send(resp); err != nil {
			logger.Warnf("Send to service endpoints stream %s failed: %v", session.remote.Addr, err)
			return
		}
	}
}

func (session *endpointDiscoverySession) subscribe(svcNames ...string) {
	for _, svcName := range svcNames {
		_, ok := session.subscribed[svcName]
		if ok {
			continue
		}

		if session.subHdlr != nil {
			session.subHdlr(svcName, session)
		}
		session.subscribed[svcName] = struct{}{}
	}
}

func (session *endpointDiscoverySession) unsubscribe(svcNames ...string) {
	for _, svcName := range svcNames {
		_, ok := session.subscribed[svcName]
		if !ok {
			continue
		}
		if session.unsubHdlr != nil {
			session.unsubHdlr(svcName, session)
		}
		delete(session.subscribed, svcName)
	}
}

func (session *endpointDiscoverySession) unsubscribeAll() {
	for svcName := range session.subscribed {
		session.unsubscribe(svcName)
	}
}

func (session *endpointDiscoverySession) SendEvent(event *endpointEvent) {
	select {
	case session.eventCh <- event:
	case <-session.quit:
	}
}

type endpointDiscoverySessions map[*endpointDiscoverySession]interface{}

type endpointDiscoveryServer struct {
	sync.RWMutex
	reg registry.Cache

	subscribers map[string]endpointDiscoverySessions // service: sessions
}

func newEndpointDiscoveryServer(reg registry.Cache) *endpointDiscoveryServer {
	s := &endpointDiscoveryServer{
		reg:         reg,
		subscribers: make(map[string]endpointDiscoverySessions),
	}

	// register handler that handles registry event
	s.reg.RegisterServiceEventHandler(s.handleRegServiceEvent)
	s.reg.RegisterInstanceEventHandler(s.handleRegInstanceEvent)

	return s
}

func (s *endpointDiscoveryServer) handleRegServiceEvent(event *registry.ServiceEvent) {
	svc := event.Service
	insts := instsMapToSlice(svc.Instances)

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

func (s *endpointDiscoveryServer) handleRegInstanceEvent(event *registry.InstanceEvent) {
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

func (s *endpointDiscoveryServer) dispatchEvent(event *endpointEvent) {
	// TODO: improve map performance.
	s.RLock()
	defer s.RUnlock()
	subscribers := s.subscribers[event.SvcName]
	for subscriber := range subscribers {
		subscriber.SendEvent(event)
	}
}

func (s *endpointDiscoveryServer) handleSubscribe(svcName string, c *endpointDiscoverySession) {
	s.Lock()
	defer s.Unlock()
	subscribers, ok := s.subscribers[svcName]
	if !ok {
		subscribers = endpointDiscoverySessions{}
		s.subscribers[svcName] = subscribers
	}

	if _, ok := subscribers[c]; ok {
		return
	}

	subscribers[c] = struct{}{}
	// send all instances when first subscribe
	svc, _ := s.reg.Get(svcName)
	// service doesn't exist.
	if svc == nil {
		return
	}

	insts := instsMapToSlice(svc.Instances)
	event := newEndpointEvent(svc.Name, insts, nil, nil)
	c.SendEvent(event)
}

func (s *endpointDiscoveryServer) handleUnsubscribe(svcName string, c *endpointDiscoverySession) {
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

// Subscribers returns all subscribers. It's only for test, and not goroutine-safe.
func (s *endpointDiscoveryServer) Subscribers() map[string]endpointDiscoverySessions {
	return s.subscribers
}

func (s *endpointDiscoveryServer) StreamSvcEndpoints(stream api.DiscoveryService_StreamSvcEndpointsServer) (err error) {
	session := newEndpointDiscoverySession(stream)
	session.SetSubscribeHandler(s.handleSubscribe)
	session.SetUnsubscribeHandler(s.handleUnsubscribe)
	session.Serve()
	return
}

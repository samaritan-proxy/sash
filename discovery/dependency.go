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

//go:generate mockgen -package $GOPACKAGE -self_package github.com/samaritan-proxy/sash/$GOPACKAGE --destination ./mock_dependency_test.go github.com/samaritan-proxy/samaritan-api/go/api DiscoveryService_StreamDependenciesServer

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/samaritan-proxy/samaritan-api/go/api"
	"github.com/samaritan-proxy/samaritan-api/go/config/service"
	"google.golang.org/grpc/peer"

	"github.com/samaritan-proxy/sash/config"
	"github.com/samaritan-proxy/sash/logger"
)

type dependencyUnsubHandler func(svcName string, session *dependencyDiscoverySession)

type dependencyDiscoverySession struct {
	instID string
	stream api.DiscoveryService_StreamDependenciesServer
	remote *peer.Peer

	eventCh   chan *config.Event
	unsubHdlr dependencyUnsubHandler

	lastDeps []string

	quit chan struct{}
}

func newDependencyDiscoverySession(instID string, stream api.DiscoveryService_StreamDependenciesServer) *dependencyDiscoverySession {
	remote, _ := peer.FromContext(stream.Context())
	return &dependencyDiscoverySession{
		instID:  instID,
		stream:  stream,
		remote:  remote,
		eventCh: make(chan *config.Event, 64),
		quit:    make(chan struct{}),
	}
}

func diffSlice(a, b []string) (add, del []string) {
	setA := make(map[string]struct{}, len(add))
	setB := make(map[string]struct{}, len(del))
	all := make(map[string]struct{}, len(add)+len(del))
	for _, ele := range a {
		setA[ele] = struct{}{}
		all[ele] = struct{}{}
	}
	for _, ele := range b {
		setB[ele] = struct{}{}
		all[ele] = struct{}{}
	}
	for k := range all {
		_, inA := setA[k]
		_, inB := setB[k]
		switch {
		case inA && !inB:
			del = append(del, k)
		case !inA && inB:
			add = append(add, k)
		}
	}
	return
}

func buildServices(svcNames ...string) (services []*service.Service) {
	for _, svcName := range svcNames {
		services = append(services, &service.Service{
			Name: svcName,
		})
	}
	return
}

func (s *dependencyDiscoverySession) buildRespFromEvt(event *config.Event) (*api.DependencyDiscoveryResponse, error) {
	deps := make([]string, 0, 4)
	if err := json.Unmarshal(event.Config.Value, &deps); err != nil {
		return nil, err
	}
	var add, del []string
	switch event.Type {
	case config.EventAdd:
		add = deps
	case config.EventUpdate:
		add, del = diffSlice(s.lastDeps, deps)
	case config.EventDelete:
		del = s.lastDeps
	}
	s.lastDeps = deps
	return &api.DependencyDiscoveryResponse{
		Added:   buildServices(add...),
		Removed: buildServices(del...),
	}, nil
}

func (s *dependencyDiscoverySession) SetUnsubscribeHandler(hdlr dependencyUnsubHandler) {
	s.unsubHdlr = hdlr
}

func (s *dependencyDiscoverySession) Serve() {
	defer func() {
		close(s.quit)
		if s.unsubHdlr != nil {
			s.unsubHdlr(s.instID, s)
		}
		logger.Debugf("Dependency discovery session %s exit", s.remote.Addr)
	}()

	for {
		var event *config.Event
		select {
		case event = <-s.eventCh:
		case <-s.stream.Context().Done():
			return
		}
		resp, err := s.buildRespFromEvt(event)
		if err != nil {
			logger.Warnf("failed dependency response from config event, err: %v", err)
			continue
		}
		if err := s.stream.Send(resp); err != nil {
			logger.Warnf("Send to dependency stream %s failed: %v", s.remote.Addr, err)
			return
		}
	}
}

func (s *dependencyDiscoverySession) SendEvent(event *config.Event) {
	select {
	case s.eventCh <- event:
	case <-s.quit:
	}
}

type dependencyDiscoveryServer struct {
	sync.RWMutex
	ctl *config.Controller

	subscribers map[string]*dependencyDiscoverySession
}

func newDependencyDiscoveryServer(ctl *config.Controller) *dependencyDiscoveryServer {
	s := &dependencyDiscoveryServer{
		ctl:         ctl,
		subscribers: make(map[string]*dependencyDiscoverySession),
	}

	s.ctl.RegisterEventHandler(s.handleConfigEvent)
	return s
}

func (s *dependencyDiscoveryServer) handleConfigEvent(evt *config.Event) {
	if evt == nil ||
		evt.Config.Namespace != config.NamespaceService ||
		evt.Config.Type != config.TypeServiceDependency {
		return
	}
	subscriber, ok := s.subscribers[evt.Config.Key]
	if !ok {
		return
	}
	subscriber.SendEvent(evt)
}

func (s *dependencyDiscoveryServer) handleUnsubscribe(instName string, d *dependencyDiscoverySession) {
	s.Lock()
	defer s.Unlock()
	// Go runtime never shrink map after elements removal, refer to: https://github.com/golang/go/issues/20135
	// FIXME: To prevent OOM after long running, we should add some memchainsm recycle the memory.
	delete(s.subscribers, instName)
}

func (s *dependencyDiscoveryServer) StreamDependencies(req *api.DependencyDiscoveryRequest, stream api.DiscoveryService_StreamDependenciesServer) error {
	if err := req.Instance.Validate(); err != nil {
		return err
	}

	instID := req.Instance.Id

	s.Lock()
	_, ok := s.subscribers[instID]
	if ok {
		s.Unlock()
		return fmt.Errorf("instID: %s, already registered", instID)
	}
	session := newDependencyDiscoverySession(req.Instance.Id, stream)
	session.SetUnsubscribeHandler(s.handleUnsubscribe)
	s.subscribers[req.Instance.Id] = session
	s.Unlock()
	session.Serve()
	return nil
}

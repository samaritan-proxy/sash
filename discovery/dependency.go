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

func buildServices(svcNames ...string) (services []*service.Service) {
	for _, svcName := range svcNames {
		services = append(services, &service.Service{
			Name: svcName,
		})
	}
	return
}

type dependencyEvent struct {
	Add []*service.Service
	Del []*service.Service
}

func buildDependencyEvent(add, del []string) *dependencyEvent {
	return &dependencyEvent{
		Add: buildServices(add...),
		Del: buildServices(del...),
	}
}

func unmarshalServiceDeps(b []byte) ([]string, error) {
	deps := make([]string, 0, 4)
	if err := json.Unmarshal(b, &deps); err != nil {
		return nil, err
	}
	return deps, nil
}

type dependencyDiscoverySession struct {
	instID string
	stream api.DiscoveryService_StreamDependenciesServer
	remote *peer.Peer

	eventCh chan *dependencyEvent

	quit chan struct{}
}

func newDependencyDiscoverySession(instID string, stream api.DiscoveryService_StreamDependenciesServer) *dependencyDiscoverySession {
	remote, _ := peer.FromContext(stream.Context())
	return &dependencyDiscoverySession{
		instID:  instID,
		stream:  stream,
		remote:  remote,
		eventCh: make(chan *dependencyEvent, 16),
		quit:    make(chan struct{}),
	}
}

func (s *dependencyDiscoverySession) Serve() {
	defer func() {
		close(s.quit)
		logger.Debugf("Dependency discovery session %s exit", s.remote.Addr)
	}()

	for {
		select {
		case <-s.stream.Context().Done():
			return
		case event := <-s.eventCh:
			err := s.stream.Send(&api.DependencyDiscoveryResponse{
				Added:   event.Add,
				Removed: event.Del,
			})
			if err != nil {
				logger.Warnf("Send to dependency stream %s failed: %v", s.remote.Addr, err)
				return
			}
		}
	}
}

func (s *dependencyDiscoverySession) SendEvent(event *dependencyEvent) {
	select {
	case s.eventCh <- event:
	case <-s.quit:
	}
}

type dependencyDiscoverySessions map[*dependencyDiscoverySession]interface{}

type dependencyDiscoveryServer struct {
	sync.RWMutex
	ctl *config.Controller

	dependencies map[string][]string // serviceName, dependencies
	subscribers  map[string]dependencyDiscoverySessions
}

func newDependencyDiscoveryServer(ctl *config.Controller) *dependencyDiscoveryServer {
	s := &dependencyDiscoveryServer{
		ctl:          ctl,
		dependencies: make(map[string][]string),
		subscribers:  make(map[string]dependencyDiscoverySessions),
	}

	s.ctl.RegisterEventHandler(s.handleConfigEvent)
	return s
}

func (s *dependencyDiscoveryServer) buildDepEvtFromCfgEvt(cfgEvt *config.Event) (*dependencyEvent, error) {
	svcName := cfgEvt.Config.Key
	deps, err := unmarshalServiceDeps(cfgEvt.Config.Value)
	if err != nil {
		return nil, err
	}
	var add, del []string
	switch cfgEvt.Type {
	case config.EventAdd:
		add = deps
	case config.EventUpdate:
		add, del = diffSlice(s.dependencies[svcName], deps)
	case config.EventDelete:
		del = s.dependencies[svcName]
	}
	s.dependencies[svcName] = deps
	return buildDependencyEvent(add, del), nil
}

func (s *dependencyDiscoveryServer) handleConfigEvent(evt *config.Event) {
	if evt == nil ||
		evt.Config.Namespace != config.NamespaceService ||
		evt.Config.Type != config.TypeServiceDependency {
		return
	}

	svcName := evt.Config.Key

	s.Lock()
	defer s.Unlock()
	depEvt, err := s.buildDepEvtFromCfgEvt(evt)
	if err != nil {
		logger.Warnf("failed to build dependency event of service[%s], err: %v", svcName, err)
		return
	}
	subscribers, ok := s.subscribers[svcName]
	if !ok {
		return
	}
	for subscriber := range subscribers {
		subscriber.SendEvent(depEvt)
	}
}

func (s *dependencyDiscoveryServer) regSession(belongSvc string, session *dependencyDiscoverySession) {
	s.Lock()
	defer s.Unlock()
	subscribers, ok := s.subscribers[belongSvc]
	if !ok {
		subscribers = dependencyDiscoverySessions{}
		s.subscribers[belongSvc] = subscribers
	}
	subscribers[session] = struct{}{}
}

func (s *dependencyDiscoveryServer) unRegSession(belongSvc string, session *dependencyDiscoverySession) {
	s.Lock()
	defer s.Unlock()
	subscribers, ok := s.subscribers[belongSvc]
	if !ok {
		return
	}
	delete(subscribers, session)
	if len(subscribers) == 0 {
		delete(s.subscribers, belongSvc)
	}
}

func (s *dependencyDiscoveryServer) StreamDependencies(req *api.DependencyDiscoveryRequest, stream api.DiscoveryService_StreamDependenciesServer) error {
	if err := req.Instance.Validate(); err != nil {
		return err
	}

	instID := req.Instance.Id
	belongSvc := req.Instance.Belong

	if len(belongSvc) == 0 {
		return fmt.Errorf("instID: %s, Belong is null", instID)
	}

	session := newDependencyDiscoverySession(req.Instance.Id, stream)
	s.regSession(belongSvc, session)
	defer s.unRegSession(belongSvc, session)

	if b, err := s.ctl.GetCache(config.NamespaceService, config.TypeServiceDependency, belongSvc); err == nil {
		if deps, err := unmarshalServiceDeps(b); err == nil {
			s.Lock()
			s.dependencies[belongSvc] = deps
			s.Unlock()
			session.SendEvent(buildDependencyEvent(deps, nil))
		}
	}

	session.Serve()
	return nil
}

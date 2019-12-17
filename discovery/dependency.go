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

type dependencyDiscoverySession struct {
	instID string
	stream api.DiscoveryService_StreamDependenciesServer
	remote *peer.Peer

	eventCh chan *config.DependencyEvent

	quit chan struct{}
}

func newDependencyDiscoverySession(instID string, stream api.DiscoveryService_StreamDependenciesServer) *dependencyDiscoverySession {
	remote, _ := peer.FromContext(stream.Context())
	return &dependencyDiscoverySession{
		instID:  instID,
		stream:  stream,
		remote:  remote,
		eventCh: make(chan *config.DependencyEvent, 16),
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
				Added:   buildServices(event.Add...),
				Removed: buildServices(event.Del...),
			})
			if err != nil {
				logger.Warnf("Send to dependency stream %s failed: %v", s.remote.Addr, err)
				return
			}
		}
	}
}

func (s *dependencyDiscoverySession) SendEvent(event *config.DependencyEvent) {
	select {
	case s.eventCh <- event:
	case <-s.quit:
	}
}

type dependencyDiscoverySessions map[*dependencyDiscoverySession]interface{}

type dependencyDiscoveryServer struct {
	sync.RWMutex
	depCtl *config.DependenciesController

	dependencies map[string][]string // serviceName, dependencies
	subscribers  map[string]dependencyDiscoverySessions
}

func newDependencyDiscoveryServer(ctl *config.Controller) *dependencyDiscoveryServer {
	s := &dependencyDiscoveryServer{
		depCtl:       ctl.Dependencies(),
		dependencies: make(map[string][]string),
		subscribers:  make(map[string]dependencyDiscoverySessions),
	}

	s.depCtl.RegisterEventHandler(s.handleConfigEvent)
	return s
}

func (s *dependencyDiscoveryServer) handleConfigEvent(evt *config.DependencyEvent) {
	s.Lock()
	defer s.Unlock()
	for subscriber := range s.subscribers[evt.ServiceName] {
		subscriber.SendEvent(evt)
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

	if dep, err := s.depCtl.GetCache(belongSvc); err == nil {
		session.SendEvent(&config.DependencyEvent{
			ServiceName: dep.ServiceName,
			Add:         dep.Dependencies,
		})
	}
	session.Serve()
	return nil
}

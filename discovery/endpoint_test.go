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
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/samaritan-proxy/samaritan-api/go/api"
	"github.com/samaritan-proxy/samaritan-api/go/common"
	"github.com/samaritan-proxy/samaritan-api/go/config/service"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/peer"

	"github.com/samaritan-proxy/sash/model"
	"github.com/samaritan-proxy/sash/registry"
)

func makeEndpoint(ip string, port uint32) *service.Endpoint {
	return &service.Endpoint{
		Address: &common.Address{
			Ip:   ip,
			Port: port,
		},
	}
}
func makeEndpointEvent(svcName string, added, updated, removed []*service.Endpoint) *endpointEvent {
	return &endpointEvent{
		SvcName: svcName,
		Added:   added,
		Updated: updated,
		Removed: removed,
	}
}

func TestNewEndpointEvent(t *testing.T) {
	svcName := "foo"
	added := []*model.ServiceInstance{
		model.NewServiceInstance("127.0.0.1", 8888),
	}
	updated := []*model.ServiceInstance{
		model.NewServiceInstance("127.0.0.1", 8889),
	}
	removed := []*model.ServiceInstance{
		model.NewServiceInstance("127.0.0.1", 9000),
	}
	event := newEndpointEvent(svcName, added, updated, removed)
	expected := &endpointEvent{
		SvcName: "foo",
		Added:   []*service.Endpoint{makeEndpoint("127.0.0.1", 8888)},
		Updated: []*service.Endpoint{makeEndpoint("127.0.0.1", 8889)},
		Removed: []*service.Endpoint{makeEndpoint("127.0.0.1", 9000)},
	}
	assert.Equal(t, expected, event)
}

func makeSvcEndpointsStream(ctrl *gomock.Controller) *MockDiscoveryService_StreamSvcEndpointsServer {
	stream := NewMockDiscoveryService_StreamSvcEndpointsServer(ctrl)
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	// attach peer info
	stream.EXPECT().Context().Return(
		peer.NewContext(context.TODO(), &peer.Peer{Addr: addr}),
	)
	return stream
}

func TestEndpointDiscoverySessionSubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := makeSvcEndpointsStream(ctrl)
	session := newEndpointDiscoverySession(stream)

	// set subscribe handler
	calls := make(map[string]int)
	subHdlr := func(svcName string, session *endpointDiscoverySession) {
		calls[svcName]++
	}
	session.SetSubscribeHandler(subHdlr)

	session.subscribe("foo", "foo", "bar")
	// assert called
	assert.Equal(t, 1, calls["foo"])
	assert.Equal(t, 1, calls["bar"])
}

func TestEndpointDiscoverySessionUnsubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := makeSvcEndpointsStream(ctrl)
	session := newEndpointDiscoverySession(stream)

	// set unsubscribe handler
	calls := make(map[string]int)
	unsubHdlr := func(svcName string, session *endpointDiscoverySession) {
		calls[svcName]++
	}
	session.SetUnsubscribeHandler(unsubHdlr)

	session.subscribe("foo", "bar")
	session.unsubscribe("foo", "foo", "bar", "zoo")
	// assert called
	assert.Equal(t, 1, calls["foo"])
	assert.Equal(t, 1, calls["bar"])
	assert.Equal(t, 0, calls["zoo"])
}

func TestEndpointDiscoverySessionUnsubscribeAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := makeSvcEndpointsStream(ctrl)
	session := newEndpointDiscoverySession(stream)

	// set subscribe handler
	calls := make(map[string]int)
	unsubHdlr := func(svcName string, session *endpointDiscoverySession) {
		calls[svcName]++
	}
	session.SetUnsubscribeHandler(unsubHdlr)

	session.subscribe("foo", "bar")
	session.unsubscribeAll()
	assert.Equal(t, 1, calls["foo"])
	assert.Equal(t, 1, calls["bar"])
}

func TestEndpointDiscoverySessionSendFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := makeSvcEndpointsStream(ctrl)
	quit := make(chan struct{})
	stream.EXPECT().Send(gomock.Any()).DoAndReturn(func(resp *api.SvcEndpointDiscoveryResponse) error {
		// abort the stream on error.
		close(quit)
		return io.ErrShortWrite
	})
	stream.EXPECT().Recv().
		DoAndReturn(func() (*api.SvcEndpointDiscoveryRequest, error) {
			// wait the stream closed
			<-quit
			return nil, io.EOF
		})

	session := newEndpointDiscoverySession(stream)
	time.AfterFunc(time.Millisecond*100, func() {
		session.SendEvent(newEndpointEvent("foo", nil, nil, nil))
	})
	session.Serve()
}

func TestEndpointDiscoverySessionRecvFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := makeSvcEndpointsStream(ctrl)
	quit := make(chan struct{})
	abortStream := func() {
		close(quit)
	}
	stream.EXPECT().Recv().
		DoAndReturn(func() (*api.SvcEndpointDiscoveryRequest, error) {
			// wait the stream closed
			<-quit
			return nil, io.EOF
		})

	session := newEndpointDiscoverySession(stream)
	time.AfterFunc(time.Millisecond*100, abortStream)
	session.Serve()
}

func TestEndpointDiscoverySessionCleanSubscriptionOnExit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := makeSvcEndpointsStream(ctrl)
	quit := make(chan struct{})
	abortStream := func() {
		close(quit)
	}
	times := 0
	stream.EXPECT().Recv().
		DoAndReturn(func() (*api.SvcEndpointDiscoveryRequest, error) {
			times++
			if times < 2 {
				return &api.SvcEndpointDiscoveryRequest{
					SvcNamesSubscribe: []string{"foo", "bar", "zoo"},
				}, nil
			}
			// wait the stream closed
			<-quit
			return nil, io.EOF
		}).Times(2)

	session := newEndpointDiscoverySession(stream)
	calls := make(map[string]int)
	unsubHdlr := func(svcName string, session *endpointDiscoverySession) {
		calls[svcName]++
	}
	session.SetUnsubscribeHandler(unsubHdlr)

	time.AfterFunc(time.Millisecond*100, abortStream)
	session.Serve()
	// assert unsubscription calls
	assert.Equal(t, 1, calls["foo"])
	assert.Equal(t, 1, calls["zoo"])
	assert.Equal(t, 1, calls["bar"])
}

func makeRegistryCache(ctrl *gomock.Controller) *registry.MockCache {
	reg := registry.NewMockCache(ctrl)
	reg.EXPECT().RegisterServiceEventHandler(gomock.Any())
	reg.EXPECT().RegisterInstanceEventHandler(gomock.Any())
	return reg
}

func TestEndpointDiscoveryServerHandleSubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := makeSvcEndpointsStream(ctrl)
	session := newEndpointDiscoverySession(stream)

	reg := makeRegistryCache(ctrl)
	// foo exists, but bar don't
	fooSvc := model.NewService(
		"foo",
		model.NewServiceInstance("127.0.0.1", 8888),
		model.NewServiceInstance("127.0.0.1", 9999),
	)
	reg.EXPECT().Get("foo").Return(fooSvc, nil)
	reg.EXPECT().Get("bar").Return(nil, nil)

	s := newEndpointDiscoveryServer(reg)
	s.handleSubscribe("foo", session)
	s.handleSubscribe("foo", session) // duplicate subscribe
	s.handleSubscribe("bar", session)

	// assert subscribers
	subscribers := s.Subscribers()
	assert.Equal(t, 2, len(subscribers))
	assert.Equal(t, 1, len(subscribers["foo"]))
	assert.Equal(t, 1, len(subscribers["bar"]))
	// assert events
	assert.Equal(t, 1, len(session.eventCh))
}

func TestEndpointDiscoveryServerHandleUnsubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream1 := makeSvcEndpointsStream(ctrl)
	session1 := newEndpointDiscoverySession(stream1)
	stream2 := makeSvcEndpointsStream(ctrl)
	session2 := newEndpointDiscoverySession(stream2)

	reg := makeRegistryCache(ctrl)
	reg.EXPECT().Get("foo").Return(nil, nil)
	s := newEndpointDiscoveryServer(reg)
	s.handleSubscribe("foo", session1)

	s.handleUnsubscribe("foo", session1)
	// non-existent subscriptions
	s.handleUnsubscribe("foo", session2)
	s.handleUnsubscribe("bar", session1)

	// assert subscribers
	assert.Equal(t, 1, len(s.Subscribers()))
	assert.Equal(t, 0, len(s.Subscribers()["foo"]))
}

func TestEndpointDiscoveryServerHandleRegServiceEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reg := makeRegistryCache(ctrl)
	reg.EXPECT().Get(gomock.Any()).Return(nil, nil)
	s := newEndpointDiscoveryServer(reg)

	// subscribe foo
	stream := makeSvcEndpointsStream(ctrl)
	session := newEndpointDiscoverySession(stream)
	s.handleSubscribe("foo", session)

	tests := []struct {
		name     string
		event    *registry.ServiceEvent
		expected *endpointEvent
	}{
		{
			name: "add",
			event: makeSvcEvent(
				registry.EventAdd,
				model.NewService("foo", model.NewServiceInstance("127.0.0.1", 8888)),
			),
			expected: makeEndpointEvent(
				"foo",
				[]*service.Endpoint{makeEndpoint("127.0.0.1", 8888)},
				nil, nil,
			),
		},
		{
			name: "delete",
			event: makeSvcEvent(
				registry.EventDelete,
				model.NewService("foo", model.NewServiceInstance("127.0.0.1", 8888)),
			),
			expected: makeEndpointEvent(
				"foo", nil, nil,
				[]*service.Endpoint{makeEndpoint("127.0.0.1", 8888)},
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s.handleRegServiceEvent(test.event)
			etEvent := <-session.eventCh
			assert.Equal(t, test.expected, etEvent)
		})
	}
}

func makeSvcEvent(typ registry.EventType, svc *model.Service) *registry.ServiceEvent {
	return &registry.ServiceEvent{
		Type:    typ,
		Service: svc,
	}
}

func makeInstEvent(svcName string, typ registry.EventType, insts ...*model.ServiceInstance) *registry.InstanceEvent {
	return &registry.InstanceEvent{
		Type:        typ,
		ServiceName: svcName,
		Instances:   insts,
	}
}

func TestEndpointDiscoveryServerHandleRegInstanceEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reg := makeRegistryCache(ctrl)
	reg.EXPECT().Get(gomock.Any()).Return(nil, nil)
	s := newEndpointDiscoveryServer(reg)

	// subscribe foo
	stream := makeSvcEndpointsStream(ctrl)
	session := newEndpointDiscoverySession(stream)
	s.handleSubscribe("foo", session)

	tests := []struct {
		name     string
		event    *registry.InstanceEvent
		expected *endpointEvent
	}{
		{
			name: "add",
			event: makeInstEvent(
				"foo", registry.EventAdd,
				model.NewServiceInstance("127.0.0.1", 8888),
				model.NewServiceInstance("127.0.0.1", 8889),
			),
			expected: makeEndpointEvent(
				"foo",
				[]*service.Endpoint{
					makeEndpoint("127.0.0.1", 8888),
					makeEndpoint("127.0.0.1", 8889),
				},
				nil, nil,
			),
		},
		{
			name: "update",
			event: makeInstEvent(
				"foo", registry.EventUpdate,
				model.NewServiceInstance("127.0.0.1", 8889),
			),
			expected: makeEndpointEvent(
				"foo",
				nil,
				[]*service.Endpoint{makeEndpoint("127.0.0.1", 8889)},
				nil,
			),
		},
		{
			name: "delete",
			event: makeInstEvent(
				"foo", registry.EventDelete,
				model.NewServiceInstance("127.0.0.1", 8888),
			),
			expected: makeEndpointEvent(
				"foo", nil, nil,
				[]*service.Endpoint{makeEndpoint("127.0.0.1", 8888)},
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s.handleRegInstanceEvent(test.event)
			etEvent := <-session.eventCh
			assert.Equal(t, test.expected, etEvent)
		})
	}
}

func TestEndpointDiscoveryServerStreamSvcEndpoints(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reg := makeRegistryCache(ctrl)
	s := newEndpointDiscoveryServer(reg)

	stream := makeSvcEndpointsStream(ctrl)
	quit := make(chan struct{})
	abortStream := func() {
		close(quit)
	}
	stream.EXPECT().Recv().DoAndReturn(func() (*api.SvcEndpointDiscoveryRequest, error) {
		<-quit
		return nil, io.EOF
	})

	time.AfterFunc(time.Millisecond*100, abortStream)
	assert.NoError(t, s.StreamSvcEndpoints(stream))
}

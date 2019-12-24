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
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/peer"

	"github.com/samaritan-proxy/sash/config"
	"github.com/samaritan-proxy/sash/config/memory"
)

func makeSvcConfigsStream(ctrl *gomock.Controller) *MockDiscoveryService_StreamSvcConfigsServer {
	stream := NewMockDiscoveryService_StreamSvcConfigsServer(ctrl)
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	// attach peer info
	stream.EXPECT().Context().Return(
		peer.NewContext(context.TODO(), &peer.Peer{Addr: addr}),
	)
	return stream
}

func TestConfigDiscoverySessionSubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := makeSvcConfigsStream(ctrl)
	session := newConfigDiscoverySession(stream)

	// set subscribe handler
	calls := make(map[string]int)
	subHdlr := func(svcName string, session *configDiscoverySession) {
		calls[svcName]++
	}
	session.SetSubscribeHandler(subHdlr)

	session.handleSubscribe("foo", "foo", "bar")
	// assert called
	assert.Equal(t, 1, calls["foo"])
	assert.Equal(t, 1, calls["bar"])
}

func TestConfigDiscoverySessionUnsubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := makeSvcConfigsStream(ctrl)
	session := newConfigDiscoverySession(stream)

	// set unsubscribe handler
	calls := make(map[string]int)
	unsubHdlr := func(svcName string, session *configDiscoverySession) {
		calls[svcName]++
	}
	session.SetUnsubscribeHandler(unsubHdlr)

	session.handleSubscribe("foo", "bar")
	session.handleUnsubscribe("foo", "foo", "bar", "zoo")
	// assert called
	assert.Equal(t, 1, calls["foo"])
	assert.Equal(t, 1, calls["bar"])
	assert.Equal(t, 0, calls["zoo"])
}

func TestConfigDiscoverySessionUnsubscribeAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := makeSvcConfigsStream(ctrl)
	session := newConfigDiscoverySession(stream)

	// set subscribe handler
	calls := make(map[string]int)
	unsubHdlr := func(svcName string, session *configDiscoverySession) {
		calls[svcName]++
	}
	session.SetUnsubscribeHandler(unsubHdlr)

	session.handleSubscribe("foo", "bar")
	session.unsubscribeAll()
	assert.Equal(t, 1, calls["foo"])
	assert.Equal(t, 1, calls["bar"])
}

func TestConfigDiscoverySessionSendFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := makeSvcConfigsStream(ctrl)
	streamQuitCh := make(chan struct{})
	stream.EXPECT().Send(gomock.Any()).DoAndReturn(func(resp *api.SvcConfigDiscoveryResponse) error {
		// abort the stream on error.
		close(streamQuitCh)
		return io.ErrShortWrite
	})
	stream.EXPECT().Recv().
		DoAndReturn(func() (*api.SvcConfigDiscoveryRequest, error) {
			// wait the stream closed
			<-streamQuitCh
			return nil, io.EOF
		})

	session := newConfigDiscoverySession(stream)
	time.AfterFunc(time.Millisecond*100, func() {
		session.SendEvent(&config.ProxyConfigEvent{
			Type: config.EventDelete,
			ProxyConfig: &config.ProxyConfig{
				ServiceName: "foo",
				Config:      nil,
			},
		})
	})
	session.Serve()
}

func TestConfigDiscoverySessionRecvFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := makeSvcConfigsStream(ctrl)
	streamQuitCh := make(chan struct{})
	abortStream := func() {
		close(streamQuitCh)
	}
	stream.EXPECT().Recv().
		DoAndReturn(func() (*api.SvcConfigDiscoveryRequest, error) {
			// wait the stream closed
			<-streamQuitCh
			return nil, io.EOF
		})

	session := newConfigDiscoverySession(stream)
	time.AfterFunc(time.Millisecond*100, abortStream)
	session.Serve()
}

func TestConfigDiscoverySessionCleanSubscriptionOnExit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := makeSvcConfigsStream(ctrl)
	streamQuitCh := make(chan struct{})
	abortStream := func() {
		close(streamQuitCh)
	}
	times := 0
	stream.EXPECT().Recv().
		DoAndReturn(func() (*api.SvcConfigDiscoveryRequest, error) {
			times++
			if times < 2 {
				return &api.SvcConfigDiscoveryRequest{
					SvcNamesSubscribe: []string{"foo", "bar", "zoo"},
				}, nil
			}
			// wait the stream closed
			<-streamQuitCh
			return nil, io.EOF
		}).Times(2)

	session := newConfigDiscoverySession(stream)
	calls := make(map[string]int)
	unsubHdlr := func(svcName string, session *configDiscoverySession) {
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

func TestConfigDiscoveryServerHandleSubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := makeSvcConfigsStream(ctrl)
	session := newConfigDiscoverySession(stream)

	ctl := config.NewController(memory.NewMemStore(), config.Interval(time.Millisecond))
	assert.NoError(t, ctl.Start())
	defer ctl.Stop()
	assert.NoError(t, ctl.Add(config.NamespaceService, config.TypeServiceProxyConfig, "foo", nil))
	assert.NoError(t, ctl.Add(config.NamespaceService, config.TypeServiceProxyConfig, "bar", []byte{0, 0, 0, 0}))

	time.Sleep(time.Millisecond * 10)

	s := newConfigDiscoveryServer(ctl)
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

func TestConfigDiscoveryServerHandleUnsubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream1 := makeSvcConfigsStream(ctrl)
	session1 := newConfigDiscoverySession(stream1)
	stream2 := makeSvcConfigsStream(ctrl)
	session2 := newConfigDiscoverySession(stream2)

	ctl := config.NewController(memory.NewMemStore(), config.Interval(time.Second))
	assert.NoError(t, ctl.Start())
	defer ctl.Stop()
	time.Sleep(time.Millisecond * 1500)
	assert.NoError(t, ctl.Add(config.NamespaceService, config.TypeServiceProxyConfig, "foo", nil))

	s := newConfigDiscoveryServer(ctl)
	s.handleSubscribe("foo", session1)

	s.handleUnsubscribe("foo", session1)
	// non-existent subscriptions
	s.handleUnsubscribe("foo", session2)
	s.handleUnsubscribe("bar", session1)

	// assert subscribers
	assert.Equal(t, 1, len(s.Subscribers()))
	assert.Equal(t, 0, len(s.Subscribers()["foo"]))
}

func TestConfigDiscoveryServerStreamSvcConfigs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctl := config.NewController(memory.NewMemStore(), config.Interval(time.Second))
	assert.NoError(t, ctl.Start())
	defer ctl.Stop()
	time.Sleep(time.Millisecond * 1500)
	s := newConfigDiscoveryServer(ctl)

	stream := makeSvcConfigsStream(ctrl)
	streamQuitCh := make(chan struct{})
	abortStream := func() {
		close(streamQuitCh)
	}
	stream.EXPECT().Recv().DoAndReturn(func() (*api.SvcConfigDiscoveryRequest, error) {
		<-streamQuitCh
		return nil, io.EOF
	})

	time.AfterFunc(time.Millisecond*100, abortStream)
	assert.NoError(t, s.StreamSvcConfigs(stream))
}

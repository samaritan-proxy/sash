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
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/samaritan-proxy/samaritan-api/go/api"
	"github.com/samaritan-proxy/samaritan-api/go/common"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/peer"

	"github.com/samaritan-proxy/sash/config"
	"github.com/samaritan-proxy/sash/config/memory"
)

func TestDiffSlice(t *testing.T) {
	cases := []struct {
		A, B     []string
		Add, Del []string
	}{
		{
			A:   []string{},
			B:   []string{},
			Add: []string{},
			Del: []string{},
		},
		{
			A:   nil,
			B:   nil,
			Add: []string{},
			Del: []string{},
		},
		{
			A:   []string{"A", "B"},
			B:   []string{"B", "C"},
			Add: []string{"C"},
			Del: []string{"A"},
		},
	}
	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			add, del := diffSlice(c.A, c.B)
			assert.ElementsMatch(t, c.Add, add)
			assert.ElementsMatch(t, c.Del, del)
		})
	}
}

func makeDependenciesStream(ctrl *gomock.Controller) (*MockDiscoveryService_StreamDependenciesServer, func()) {
	stream := NewMockDiscoveryService_StreamDependenciesServer(ctrl)
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	// attach peer info
	ctx, cancel := context.WithCancel(context.TODO())
	stream.EXPECT().Context().Return(
		peer.NewContext(ctx, &peer.Peer{Addr: addr}),
	).AnyTimes()
	return stream, cancel
}

func TestDependencyDiscoverySessionBuildRespFromEvt(t *testing.T) {
	ctrl := gomock.NewController(t)
	stream, _ := makeDependenciesStream(ctrl)
	s := newDependencyDiscoverySession("foo", stream)

	cases := []struct {
		Event    *config.Event
		LastDeps []string
		Response *api.DependencyDiscoveryResponse
		IsError  bool
	}{
		{
			Event: config.NewEvent(
				config.EventAdd,
				config.NewRawConf(
					config.NamespaceService, config.TypeServiceDependency, "svc", []byte(`["foo", "bar"]`),
				),
			),
			LastDeps: []string{},
			Response: &api.DependencyDiscoveryResponse{
				Added:   buildServices("foo", "bar"),
				Removed: nil,
			},
			IsError: false,
		},
		{
			Event: config.NewEvent(
				config.EventUpdate,
				config.NewRawConf(
					config.NamespaceService, config.TypeServiceDependency, "svc", []byte(`["1", "2"]`),
				),
			),
			LastDeps: []string{"1", "3"},
			Response: &api.DependencyDiscoveryResponse{
				Added:   buildServices("2"),
				Removed: buildServices("3"),
			},
			IsError: false,
		},
		{
			Event: config.NewEvent(
				config.EventDelete,
				config.NewRawConf(
					config.NamespaceService, config.TypeServiceDependency, "svc", []byte(`[]`)),
			),
			LastDeps: []string{"foo", "bar"},
			Response: &api.DependencyDiscoveryResponse{
				Added:   nil,
				Removed: buildServices("foo", "bar"),
			},
			IsError: false,
		},
		{
			Event: config.NewEvent(
				config.EventAdd,
				config.NewRawConf(
					config.NamespaceService, config.TypeServiceDependency, "svc", nil),
			),
			IsError: true,
		},
	}
	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			s.lastDeps = c.LastDeps
			evt, err := s.buildRespFromEvt(c.Event)
			if c.IsError {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, c.Response, evt)
		})
	}

}

func TestDependenciesDiscoverySessionServeWithRepeat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		stream, cancel = makeDependenciesStream(ctrl)
		ctl            = config.NewController(memory.NewMemStore(), config.Interval(time.Millisecond*100))
		server         = newDependencyDiscoveryServer(ctl)

		errCount    int32
		repeatCount = 3
		done        = make(chan struct{})
	)

	for i := 0; i < repeatCount; i++ {
		go func() {
			err := server.StreamDependencies(&api.DependencyDiscoveryRequest{
				Instance: &common.Instance{
					Id: "foo",
				},
			}, stream)
			if err != nil {
				atomic.AddInt32(&errCount, 1)
				return
			}
			close(done)
		}()
	}
	time.Sleep(time.Millisecond)
	cancel()
	<-done
	assert.EqualValues(t, repeatCount-1, atomic.LoadInt32(&errCount))
}

func TestDependenciesDiscoverySessionServe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream, cancel := makeDependenciesStream(ctrl)
	stream.EXPECT().Send(gomock.Any()).DoAndReturn(func(resp *api.DependencyDiscoveryResponse) error {
		assert.ElementsMatch(t, buildServices("svc1", "svc2"), resp.Added)
		return nil
	}).AnyTimes()

	ctl := config.NewController(memory.NewMemStore(), config.Interval(time.Millisecond*100))
	assert.NoError(t, ctl.Start())
	defer ctl.Stop()
	time.Sleep(time.Second)

	server := newDependencyDiscoveryServer(ctl)
	done := make(chan struct{})

	go func() {
		err := server.StreamDependencies(&api.DependencyDiscoveryRequest{
			Instance: &common.Instance{
				Id: "foo",
			},
		}, stream)
		assert.NoError(t, err)
		close(done)
	}()
	time.Sleep(time.Millisecond * 100)
	assert.NoError(t, ctl.Set(config.NamespaceService, config.TypeServiceDependency, "foo", []byte(`["svc1", "svc2"]`)))
	time.Sleep(time.Millisecond * 100)
	cancel()
	<-done
}

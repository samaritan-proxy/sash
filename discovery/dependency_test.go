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
	"sync"
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

func TestDependencyDiscoverySessionServe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		stream, cancel = makeDependenciesStream(ctrl)
		session        = newDependencyDiscoverySession("inst_0", stream)
		serveDone      = make(chan struct{})
	)
	stream.EXPECT().Send(gomock.Any()).DoAndReturn(func(resp *api.DependencyDiscoveryResponse) error {
		assert.Equal(t, buildServices("add_1", "add_2"), resp.Added)
		assert.Equal(t, buildServices("del_1", "del_2"), resp.Removed)
		return nil
	})
	go func() {
		session.Serve()
		close(serveDone)
	}()

	session.SendEvent(&config.DependencyEvent{
		ServiceName: "inst_0",
		Add:         []string{"add_1", "add_2"},
		Del:         []string{"del_1", "del_2"},
	})
	time.Sleep(time.Millisecond * 100)

	cancel()

	select {
	case <-time.NewTicker(time.Second).C:
		t.Fatal("close timeout")
	case <-serveDone:
	}
}

func TestDependencyDiscoverySessionServeWithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		stream, _ = makeDependenciesStream(ctrl)
		session   = newDependencyDiscoverySession("inst_0", stream)
		serveDone = make(chan struct{})
	)
	stream.EXPECT().Send(gomock.Any()).Return(io.ErrUnexpectedEOF)
	go func() {
		session.Serve()
		close(serveDone)
	}()

	session.SendEvent(&config.DependencyEvent{
		ServiceName: "inst_0",
		Add:         []string{"add_1", "add_2"},
		Del:         []string{"del_1", "del_2"},
	})

	select {
	case <-time.NewTicker(time.Second).C:
		t.Fatal("close timeout")
	case <-serveDone:
	}
}

func TestDependenciesDiscoverySessionServeInitPush(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := memory.NewMemStore()

	ctl := config.NewController(store, config.Interval(time.Millisecond))
	assert.NoError(t, ctl.Start())
	defer ctl.Stop()

	server := newDependencyDiscoveryServer(ctl)

	stream1, cancel1 := makeDependenciesStream(ctrl)
	stream1.EXPECT().Send(gomock.Any()).Return(nil)
	stream2, cancel2 := makeDependenciesStream(ctrl)

	time.Sleep(time.Millisecond * 50)

	assert.NoError(t, store.Add(config.NamespaceService, config.TypeServiceDependency, "svc", []byte(`["dep_1", "dep_2"]`)))

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Millisecond * 10)
		defer ticker.Stop()
		for {
			<-ticker.C
			if _, err := server.depCtl.GetCache("svc"); err == nil {
				close(done)
				return
			}
		}
	}()
	timer := time.NewTimer(time.Second)
Loop:
	for {
		select {
		case <-timer.C:
			t.Fatal("timeout")
		case <-done:
			break Loop
		}
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.StreamDependencies(&api.DependencyDiscoveryRequest{Instance: &common.Instance{
			Id:     "inst_0",
			Belong: "svc",
		}}, stream1)
		assert.NoError(t, err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.StreamDependencies(&api.DependencyDiscoveryRequest{Instance: &common.Instance{
			Id:     "inst_1",
			Belong: "foo",
		}}, stream2)
		assert.NoError(t, err)
	}()

	time.Sleep(time.Millisecond * 100)

	cancel1()
	cancel2()
	wg.Wait()
}

func TestDependenciesDiscoverySessionServe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctl := config.NewController(memory.NewMemStore(), config.Interval(time.Millisecond))
	assert.NoError(t, ctl.Start())
	defer ctl.Stop()
	time.Sleep(time.Millisecond * 10)

	stream1, cancel1 := makeDependenciesStream(ctrl)
	stream1.EXPECT().Send(gomock.Any()).DoAndReturn(func(resp *api.DependencyDiscoveryResponse) error {
		assert.Equal(t, buildServices("svc1", "svc2"), resp.Added)
		return nil
	})

	stream2, cancel2 := makeDependenciesStream(ctrl)

	server := newDependencyDiscoveryServer(ctl)

	wg := sync.WaitGroup{}

	cases := []struct {
		Instance *common.Instance
		Stream   *MockDiscoveryService_StreamDependenciesServer
		IsError  bool
	}{
		{
			Instance: &common.Instance{
				Id:     "inst1",
				Belong: "svc",
			},
			Stream:  stream1,
			IsError: false,
		},
		{
			Instance: &common.Instance{
				Id: "inst2",
			},
			Stream:  stream2,
			IsError: true,
		},
	}
	for _, c := range cases {
		wg.Add(1)
		go func(wg *sync.WaitGroup, instance *common.Instance, stream *MockDiscoveryService_StreamDependenciesServer, isError bool) {
			defer wg.Done()
			err := server.StreamDependencies(&api.DependencyDiscoveryRequest{Instance: instance}, stream)
			assert.Equal(t, isError, err != nil)
		}(&wg, c.Instance, c.Stream, c.IsError)
	}

	time.Sleep(time.Millisecond * 100)
	assert.NoError(t, ctl.Add(config.NamespaceService, config.TypeServiceDependency, "svc", []byte(`["svc1", "svc2"]`)))
	time.Sleep(time.Millisecond * 100)
	cancel1()
	cancel2()
	wg.Wait()

}

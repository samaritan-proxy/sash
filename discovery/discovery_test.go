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
	"log"
	"net"
	"testing"

	"github.com/samaritan-proxy/samaritan-api/go/api"
	"github.com/samaritan-proxy/samaritan-api/go/common"
	"github.com/samaritan-proxy/samaritan-api/go/config/service"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/samaritan-proxy/sash/config"
	cfgmem "github.com/samaritan-proxy/sash/config/memory"
	"github.com/samaritan-proxy/sash/model"
	"github.com/samaritan-proxy/sash/registry"
	regmem "github.com/samaritan-proxy/sash/registry/memory"
)

func TestStreamSvcEndpoints(t *testing.T) {
	reg := regmem.NewRegistry(
		model.NewService("foo", model.NewServiceInstance("127.0.0.1", 8888)),
	)
	regCache := registry.NewCache(reg)
	ctrl := config.NewController(cfgmem.NewStore())

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	s := NewServer(l, regCache, ctrl)

	// start cache controller
	ctx, stopCache := context.WithCancel(context.TODO())
	go regCache.Run(ctx)
	defer stopCache()
	// start server
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		s.Serve() //nolint:errcheck
	}()
	defer func() {
		s.Stop()
		<-serverDone
	}()

	// create client
	conn, err := grpc.Dial(l.Addr().String(), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := api.NewDiscoveryServiceClient(conn)
	stream, err := client.StreamSvcEndpoints(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	defer stream.CloseSend() //nolint:errcheck

	// send
	req := &api.SvcEndpointDiscoveryRequest{
		SvcNamesSubscribe:   []string{"foo"},
		SvcNamesUnsubscribe: []string{"bar"},
	}
	err = stream.Send(req)
	assert.NoError(t, err)
	// recv
	resp, err := stream.Recv()
	assert.NoError(t, err)
	assert.Equal(t, "foo", resp.SvcName)
	assert.Equal(t, 1, len(resp.Added))
	assert.Equal(t, 0, len(resp.Removed))
	assert.Equal(t, &service.Endpoint{
		Address: &common.Address{
			Ip:   "127.0.0.1",
			Port: 8888,
		}}, resp.Added[0])
}

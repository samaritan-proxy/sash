package discovery

import (
	context "context"
	"log"
	"net"
	"testing"

	"github.com/samaritan-proxy/samaritan-api/go/api"
	"github.com/samaritan-proxy/samaritan-api/go/common"
	"github.com/samaritan-proxy/samaritan-api/go/config/service"
	"github.com/samaritan-proxy/sash/model"
	"github.com/samaritan-proxy/sash/registry"
	"github.com/samaritan-proxy/sash/registry/memory"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestDiscoveryServerStreamSvcEndpoints(t *testing.T) {
	reg := memory.NewRegistry(
		model.NewService("foo", model.NewServiceInstance("127.0.0.1", 8888)),
	)
	regCache := registry.NewCache(reg)
	ctx, cancel := context.WithCancel(context.TODO())
	go regCache.Run(ctx)
	defer cancel()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	s := NewServer(l, regCache)
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		s.Serve()
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
	defer stream.CloseSend()

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

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
	context "context"
	"net"

	gomock "github.com/golang/mock/gomock"
	"github.com/samaritan-proxy/samaritan-api/go/api"
	"google.golang.org/grpc/peer"
)

func makeSvcEndpointsStream(ctrl *gomock.Controller) api.DiscoveryService_StreamSvcEndpointsServer {
	stream := NewMockDiscoveryService_StreamSvcEndpointsServer(ctrl)
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	// attach peer info
	stream.EXPECT().Context().Return(
		peer.NewContext(context.TODO(), &peer.Peer{Addr: addr}),
	)
	return stream
}

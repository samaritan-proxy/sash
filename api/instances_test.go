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

package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/samaritan-proxy/sash/config"
)

func TestHandleGetAllInstances(t *testing.T) {
	cases := []struct {
		Instances []*config.Instance
		ReqURI    string
		Resp      string
		IsError   bool
	}{
		{
			Instances: []*config.Instance{
				{
					Metadata: config.Metadata{
						CreateTime: time.Time{},
						UpdateTime: time.Time{},
					},
					ID:            "inst_1",
					Hostname:      "test_host",
					IP:            "1.1.1.1",
					Port:          12345,
					Version:       "0.0.1",
					BelongService: "test_svc",
				},
			},
			ReqURI: "/api/instances",
			Resp: `{
				"data": [
					{
						"create_time":"0001-01-01T00:00:00Z",
						"update_time":"0001-01-01T00:00:00Z",
						"id": "inst_1",
						"hostname": "test_host",
						"ip": "1.1.1.1",
						"port": 12345,
						"version": "0.0.1",
						"belong_service": "test_svc"
					}
				],
				"page_num": 0,
				"page_size": 10,
				"total":1
			}`,
			IsError: false,
		},
		{
			Instances: []*config.Instance{
				{
					Metadata: config.Metadata{
						CreateTime: time.Time{},
						UpdateTime: time.Time{},
					},
					ID:            "inst_1",
					Hostname:      "test_host",
					IP:            "1.1.1.1",
					Port:          12345,
					Version:       "0.0.1",
					BelongService: "test_svc",
				},
				{
					Metadata: config.Metadata{
						CreateTime: time.Time{},
						UpdateTime: time.Time{},
					},
					ID:            "inst_2",
					Hostname:      "test_host",
					IP:            "1.1.1.1",
					Port:          12345,
					Version:       "0.0.1",
					BelongService: "test_svc",
				},
			},
			ReqURI: "/api/instances?id=inst_1",
			Resp: `{
				"data": [
					{
						"create_time":"0001-01-01T00:00:00Z",
						"update_time":"0001-01-01T00:00:00Z",
						"id": "inst_1",
						"hostname": "test_host",
						"ip": "1.1.1.1",
						"port": 12345,
						"version": "0.0.1",
						"belong_service": "test_svc"
					}
				],
				"page_num": 0,
				"page_size": 10,
				"total":1
			}`,
			IsError: false,
		},
		{
			Instances: []*config.Instance{
				{
					Metadata: config.Metadata{
						CreateTime: time.Time{},
						UpdateTime: time.Time{},
					},
					ID:            "inst_1",
					Hostname:      "test_host",
					IP:            "1.1.1.1",
					Port:          12345,
					Version:       "0.0.1",
					BelongService: "test_svc",
				},
				{
					Metadata: config.Metadata{
						CreateTime: time.Time{},
						UpdateTime: time.Time{},
					},
					ID:            "inst_foo",
					Hostname:      "test_host",
					IP:            "1.1.1.1",
					Port:          12345,
					Version:       "0.0.1",
					BelongService: "test_svc",
				},
			},
			ReqURI: "/api/instances?id=re%3Ainst_%5B0-9%5D%2B", // re:inst_[0-9]+
			Resp: `{
				"data": [
					{
						"create_time":"0001-01-01T00:00:00Z",
						"update_time":"0001-01-01T00:00:00Z",
						"id": "inst_1",
						"hostname": "test_host",
						"ip": "1.1.1.1",
						"port": 12345,
						"version": "0.0.1",
						"belong_service": "test_svc"
					}
				],
				"page_num": 0,
				"page_size": 10,
				"total":1
			}`,
			IsError: false,
		},
	}
	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			s := newTestServer(t)
			defer s.rawCtl.Stop()
			for _, inst := range c.Instances {
				assert.NoError(t, s.instCtl.Add(inst))
			}
			time.Sleep(time.Millisecond * 100)
			resp := testHandler(httptest.NewRequest(http.MethodGet, c.ReqURI, nil), s)
			if c.IsError {
				assert.NotEqual(t, http.StatusOK, resp.Code)
				return
			}
			assert.Equal(t, http.StatusOK, resp.Code)
			assert.JSONEq(t, c.Resp, resp.Body.String())
		})
	}
}

func TestHandleGetInstance(t *testing.T) {
	s := newTestServer(t)
	defer s.rawCtl.Stop()

	assert.NoError(t, s.instCtl.Add(&config.Instance{
		Metadata:      config.Metadata{},
		ID:            "inst_1",
		Hostname:      "test_host",
		IP:            "1.1.1.1",
		Port:          12345,
		Version:       "0.0.1",
		BelongService: "test_svc",
	}))
	assert.NoError(t, s.rawCtl.Add(config.NamespaceSamaritan, config.TypeSamaritanInstance, "foo", nil))

	time.Sleep(time.Millisecond * 10)

	t.Run("not found", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodGet, "/api/instances/inst_foo", nil), s)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodGet, "/api/instances/foo", nil), s)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("OK", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodGet, "/api/instances/inst_1", nil), s)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.JSONEq(t, `{
						"create_time":"0001-01-01T00:00:00Z",
						"update_time":"0001-01-01T00:00:00Z",
						"id": "inst_1",
						"hostname": "test_host",
						"ip": "1.1.1.1",
						"port": 12345,
						"version": "0.0.1",
						"belong_service": "test_svc"
					}`, resp.Body.String())
	})
}

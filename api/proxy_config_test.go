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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/samaritan-proxy/samaritan-api/go/common"
	"github.com/samaritan-proxy/samaritan-api/go/config/protocol"
	"github.com/samaritan-proxy/samaritan-api/go/config/service"
	"github.com/stretchr/testify/assert"

	"github.com/samaritan-proxy/sash/config"
)

func TestHandleGetAllProxyConfigs(t *testing.T) {
	cases := []struct {
		Configs []*config.ProxyConfig
		ReqURI  string
		Resp    string
		IsError bool
	}{
		{
			Configs: []*config.ProxyConfig{
				{
					Metadata:    config.Metadata{},
					ServiceName: "svc_1",
					Config:      nil,
				},
			},
			ReqURI: "/api/proxy-configs",
			Resp: `{
				"data": [
					{
						"create_time":"0001-01-01T00:00:00Z",
						"update_time":"0001-01-01T00:00:00Z",
						"service_name": "svc_1",
						"config": null
					}
				],
				"page_num": 0,
				"page_size": 10,
				"total":1
			}`,
		},
		{
			Configs: []*config.ProxyConfig{
				{
					Metadata:    config.Metadata{},
					ServiceName: "svc_1",
					Config:      nil,
				},
				{
					Metadata:    config.Metadata{},
					ServiceName: "svc_2",
					Config:      nil,
				},
			},
			ReqURI: "/api/proxy-configs?service_name=svc_1&config=sth",
			Resp: `{
				"data": [
					{
						"create_time":"0001-01-01T00:00:00Z",
						"update_time":"0001-01-01T00:00:00Z",
						"service_name": "svc_1",
						"config": null
					}
				],
				"page_num": 0,
				"page_size": 10,
				"total":1
			}`,
		},
		{
			Configs: []*config.ProxyConfig{
				{
					Metadata:    config.Metadata{},
					ServiceName: "svc_1",
					Config:      nil,
				},
				{
					Metadata:    config.Metadata{},
					ServiceName: "svc_foo",
					Config:      nil,
				},
			},
			ReqURI: "/api/proxy-configs?service_name=re%3Asvc_%5B0-9%5D%2B", // re:svc_[0-9]+
			Resp: `{
				"data": [
					{
						"create_time":"0001-01-01T00:00:00Z",
						"update_time":"0001-01-01T00:00:00Z",
						"service_name": "svc_1",
						"config": null
					}
				],
				"page_num": 0,
				"page_size": 10,
				"total":1
			}`,
		},
	}

	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			s := newTestServer(t)
			defer s.rawCtl.Stop()
			for _, cfg := range c.Configs {
				assert.NoError(t, s.proxyCfgCtl.Add(cfg))
			}
			time.Sleep(time.Millisecond * 100)
			resp := testHandler(httptest.NewRequest(http.MethodGet, c.ReqURI, nil), s)
			if c.IsError {
				assert.NotEqual(t, http.StatusOK, resp.Code)
				return
			}
			assert.Equal(t, http.StatusOK, resp.Code)
			// TODO: FIX THIS
			//assert.JSONEq(t, c.Resp, resp.Body.String())
		})
	}
}

func TestHandleAddProxyConfig(t *testing.T) {
	cases := []struct {
		Body       interface{} // can be bytes or struct
		StatusCode int
	}{
		{
			Body:       []byte("?"),
			StatusCode: http.StatusBadRequest,
		},
		{
			Body:       &config.ProxyConfig{},
			StatusCode: http.StatusBadRequest,
		},
		{
			Body: &config.ProxyConfig{
				ServiceName: "svc_1",
			},
			StatusCode: http.StatusOK,
		},
	}

	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			s := newTestServer(t)
			defer s.rawCtl.Stop()

			b, ok := c.Body.([]byte)
			if !ok {
				_b, err := json.Marshal(c.Body)
				assert.NoError(t, err)
				b = _b
			}
			resp := testHandler(httptest.NewRequest(http.MethodPost, "/api/proxy-configs", bytes.NewReader(b)), s)
			assert.Equal(t, c.StatusCode, resp.Code)
		})
	}
}

func TestHandleGetProxyConfig(t *testing.T) {
	s := newTestServer(t)
	defer s.rawCtl.Stop()

	assert.NoError(t, s.proxyCfgCtl.Add(&config.ProxyConfig{
		ServiceName: "svc_1",
	}))
	assert.NoError(t, s.rawCtl.Add(config.NamespaceService, config.TypeServiceProxyConfig, "svc_foo", []byte{0}))

	time.Sleep(time.Millisecond * 10)

	t.Run("not found", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodGet, "/api/proxy-configs/foo", nil), s)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodGet, "/api/proxy-configs/svc_foo", nil), s)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("OK", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodGet, "/api/proxy-configs/svc_1", nil), s)
		assert.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestHandleUpdateProxyConfig(t *testing.T) {
	s := newTestServer(t)
	defer s.rawCtl.Stop()

	assert.NoError(t, s.proxyCfgCtl.Add(&config.ProxyConfig{
		ServiceName: "svc_1",
		Config: &service.Config{
			Listener: &service.Listener{
				Address: &common.Address{
					Ip:   "1.1.1.1",
					Port: 6379,
				},
			},
			Protocol: protocol.TCP,
		},
	}))

	time.Sleep(time.Millisecond * 10)

	assert.True(t, s.proxyCfgCtl.Exist("svc_1"))

	t.Run("not found", func(t *testing.T) {
		resp := testHandler(
			httptest.NewRequest(
				http.MethodPut,
				"/api/proxy-configs/svc_foo",
				bytes.NewReader([]byte("{}"))), s)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("bad body", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodPut, "/api/proxy-configs/svc_1", nil), s)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("OK", func(t *testing.T) {
		body := []byte(`{
			"service_name": "svc_1",
			"config": {
				"listener": {
					"address": {
						"ip": "1.1.1.1",
						"port": 6379
					}
				},
				"protocol": "Redis"
			}
		}`)

		resp := testHandler(httptest.NewRequest(http.MethodPut, "/api/proxy-configs/svc_1", bytes.NewReader(body)), s)
		assert.Equal(t, http.StatusOK, resp.Code)
		cfg, err := s.proxyCfgCtl.Get("svc_1")
		assert.NoError(t, err)
		assert.Equal(t, protocol.Redis, cfg.Config.Protocol)
	})
}

func TestHandleDeleteProxyConfig(t *testing.T) {
	s := newTestServer(t)
	defer s.rawCtl.Stop()

	assert.NoError(t, s.proxyCfgCtl.Add(&config.ProxyConfig{
		ServiceName: "svc_1",
	}))

	time.Sleep(time.Millisecond * 10)

	t.Run("not found", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodDelete, "/api/proxy-configs/svc_foo", nil), s)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("OK", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodDelete, "/api/proxy-configs/svc_1", nil), s)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.False(t, s.depsCtl.Exist("svc_1"))
	})
}

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

	"github.com/stretchr/testify/assert"

	"github.com/samaritan-proxy/sash/config"
)

func TestHandleGetAllDependencies(t *testing.T) {
	cases := []struct {
		Deps    []*config.Dependency
		ReqURI  string
		Resp    string
		IsError bool
	}{
		{
			Deps: []*config.Dependency{
				{
					Metadata:     config.Metadata{},
					ServiceName:  "svc_1",
					Dependencies: []string{"dep_1", "dep_2"},
				},
			},
			ReqURI: "/dependencies",
			Resp: `{
				"data": [
					{
						"create_time":"0001-01-01T00:00:00Z",
						"update_time":"0001-01-01T00:00:00Z",
						"service_name": "svc_1",
						"dependencies": ["dep_1", "dep_2"]
					}
				],
				"page_num": 0,
				"page_size": 10,
				"total":1
			}`,
		},
		{
			Deps: []*config.Dependency{
				{
					Metadata:     config.Metadata{},
					ServiceName:  "svc_1",
					Dependencies: []string{"dep_1", "dep_2"},
				},
				{
					Metadata:     config.Metadata{},
					ServiceName:  "svc_2",
					Dependencies: []string{"dep_3", "dep_4"},
				},
			},
			ReqURI: "/dependencies?service_name=svc_1&dependencies=sth",
			Resp: `{
				"data": [
					{
						"create_time":"0001-01-01T00:00:00Z",
						"update_time":"0001-01-01T00:00:00Z",
						"service_name": "svc_1",
						"dependencies": ["dep_1", "dep_2"]
					}
				],
				"page_num": 0,
				"page_size": 10,
				"total":1
			}`,
		},
		{
			Deps: []*config.Dependency{
				{
					Metadata:     config.Metadata{},
					ServiceName:  "svc_1",
					Dependencies: []string{"dep_1", "dep_2"},
				},
				{
					Metadata:     config.Metadata{},
					ServiceName:  "svc_foo",
					Dependencies: []string{"dep_3", "dep_4"},
				},
			},
			ReqURI: "/dependencies?service_name=re%3Asvc_%5B0-9%5D%2B", // re:svc_[0-9]+
			Resp: `{
				"data": [
					{
						"create_time":"0001-01-01T00:00:00Z",
						"update_time":"0001-01-01T00:00:00Z",
						"service_name": "svc_1",
						"dependencies": ["dep_1", "dep_2"]
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
			for _, dep := range c.Deps {
				assert.NoError(t, s.depsCtl.Add(dep))
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

func TestHandleAddDependency(t *testing.T) {
	cases := []struct {
		Body       interface{} // can be bytes or struct
		StatusCode int
	}{
		{
			Body:       []byte("?"),
			StatusCode: http.StatusBadRequest,
		},
		{
			Body: &config.Dependency{
				Dependencies: []string{"dep_1", "dep_2"},
			},
			StatusCode: http.StatusBadRequest,
		},
		{
			Body: &config.Dependency{
				ServiceName:  "svc_1",
				Dependencies: []string{"dep_1", "dep_2"},
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
			resp := testHandler(httptest.NewRequest(http.MethodPost, "/dependencies", bytes.NewReader(b)), s)
			assert.Equal(t, c.StatusCode, resp.Code)
		})
	}
}

func TestHandleGetDependency(t *testing.T) {
	s := newTestServer(t)
	defer s.rawCtl.Stop()

	assert.NoError(t, s.depsCtl.Add(&config.Dependency{
		ServiceName:  "svc_1",
		Dependencies: []string{"dep_1", "dep_2"},
	}))
	assert.NoError(t, s.rawCtl.Add(config.NamespaceService, config.TypeServiceDependency, "svc_foo", []byte{0}))

	time.Sleep(time.Millisecond * 10)

	t.Run("not found", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodGet, "/dependencies/foo", nil), s)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodGet, "/dependencies/svc_foo", nil), s)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("OK", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodGet, "/dependencies/svc_1", nil), s)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.JSONEq(t, `{
						"create_time":"0001-01-01T00:00:00Z",
						"update_time":"0001-01-01T00:00:00Z",
						"service_name": "svc_1",
						"dependencies": ["dep_1", "dep_2"]
					}`, resp.Body.String())
	})
}

func TestHandleUpdateDependency(t *testing.T) {
	s := newTestServer(t)
	defer s.rawCtl.Stop()

	assert.NoError(t, s.depsCtl.Add(&config.Dependency{
		ServiceName:  "svc_1",
		Dependencies: []string{"dep_1", "dep_2"},
	}))

	time.Sleep(time.Millisecond * 10)

	t.Run("not found", func(t *testing.T) {
		resp := testHandler(
			httptest.NewRequest(
				http.MethodPut,
				"/dependencies/svc_foo",
				bytes.NewReader([]byte("{}"))), s)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("bad body", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodPut, "/dependencies/svc_1", nil), s)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("OK", func(t *testing.T) {
		body := []byte(`{
			"service_name": "svc_1",
			"dependencies": ["dep_3", "dep_4"]
		}`)

		resp := testHandler(httptest.NewRequest(http.MethodPut, "/dependencies/svc_1", bytes.NewReader(body)), s)
		assert.Equal(t, http.StatusOK, resp.Code)
		deps, err := s.depsCtl.Get("svc_1")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"dep_3", "dep_4"}, deps.Dependencies)
	})
}

func TestHandleDeleteDependency(t *testing.T) {
	s := newTestServer(t)
	defer s.rawCtl.Stop()

	assert.NoError(t, s.depsCtl.Add(&config.Dependency{
		ServiceName:  "svc_1",
		Dependencies: []string{"dep_1", "dep_2"},
	}))

	time.Sleep(time.Millisecond * 10)

	t.Run("not found", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodDelete, "/dependencies/svc_foo", nil), s)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("OK", func(t *testing.T) {
		resp := testHandler(httptest.NewRequest(http.MethodDelete, "/dependencies/svc_1", nil), s)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.False(t, s.depsCtl.Exist("svc_1"))
	})
}

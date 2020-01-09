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

package config

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/samaritan-proxy/samaritan-api/go/common"
	"github.com/samaritan-proxy/samaritan-api/go/config/protocol"
	"github.com/samaritan-proxy/samaritan-api/go/config/service"
	"github.com/stretchr/testify/assert"
)

func TestProxyConfig_Verify(t *testing.T) {
	cases := []struct {
		Config  *ProxyConfig
		IsError bool
	}{
		{Config: new(ProxyConfig), IsError: true},
		{Config: &ProxyConfig{ServiceName: "foo"}, IsError: false},
		{Config: &ProxyConfig{ServiceName: "foo", Config: &service.Config{}}, IsError: true},
		{Config: &ProxyConfig{ServiceName: "foo", Config: &service.Config{Protocol: protocol.TCP}}, IsError: true},
	}
	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			assert.Equal(t, c.IsError, c.Config.Verify() != nil)
		})
	}
}

func TestSortProxyConfigs(t *testing.T) {
	configs := ProxyConfigs{{ServiceName: "b"}, {ServiceName: "a"}, {ServiceName: "c"}}
	sort.Sort(configs)
	assert.Equal(t, ProxyConfigs{{ServiceName: "a"}, {ServiceName: "b"}, {ServiceName: "c"}}, configs)
}

func genProxyConfigsController(t *testing.T, mockCtl *gomock.Controller, configs ...*ProxyConfig) (*ProxyConfigsController, func()) {
	store := genMockStore(t, mockCtl, nil, configs, nil)
	ctl := NewController(store, Interval(time.Millisecond))
	assert.NoError(t, ctl.Start())
	time.Sleep(time.Millisecond)
	return ctl.ProxyConfigs(), ctl.Stop
}

func TestProxyConfigsController_GetNamespace(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genProxyConfigsController(t, mockCtl)
	defer cancel()

	assert.Equal(t, NamespaceService, ctl.getNamespace())
}

func TestProxyConfigsController_GetType(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genProxyConfigsController(t, mockCtl)
	defer cancel()

	assert.Equal(t, TypeServiceProxyConfig, ctl.getType())
}

func TestProxyConfigsController_UnmarshalInstance(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genProxyConfigsController(t, mockCtl)
	defer cancel()

	b, _ := (&service.Config{Protocol: protocol.TCP}).MarshalJSON()
	cases := []struct {
		Bytes       []byte
		ProxyConfig *service.Config
		IsError     bool
	}{
		{Bytes: []byte{0}, ProxyConfig: nil, IsError: true},
		{Bytes: b, ProxyConfig: &service.Config{Protocol: protocol.TCP}, IsError: false},
	}
	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			inst, err := ctl.unmarshalSvcCfg(c.Bytes)
			if c.IsError {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, c.ProxyConfig, inst)
		})
	}
}

func TestProxyConfigsController_Get(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genProxyConfigsController(t, mockCtl,
		&ProxyConfig{
			ServiceName: "foo",
		},
		&ProxyConfig{
			ServiceName: "svc",
			Config: &service.Config{
				Protocol: protocol.TCP,
			},
		},
	)
	defer cancel()

	t.Run("not exist", func(t *testing.T) {
		_, err := ctl.Get("bar")
		assert.Equal(t, ErrNotExist, err)
	})

	t.Run("nil config", func(t *testing.T) {
		cfg, err := ctl.Get("foo")
		assert.NoError(t, err)
		assert.Equal(t, "foo", cfg.ServiceName)
	})

	t.Run("ok", func(t *testing.T) {
		cfg, err := ctl.Get("svc")
		assert.NoError(t, err)
		assert.Equal(t, "svc", cfg.ServiceName)
		assert.True(t, (&service.Config{Protocol: protocol.TCP}).Equal(cfg.Config))
	})
}

func TestProxyConfigsController_GetCache(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genProxyConfigsController(t, mockCtl,
		&ProxyConfig{
			ServiceName: "foo",
		},
		&ProxyConfig{
			ServiceName: "svc",
			Config: &service.Config{
				Protocol: protocol.TCP,
			},
		},
	)
	defer cancel()

	t.Run("not exist", func(t *testing.T) {
		_, err := ctl.GetCache("bar")
		assert.Equal(t, ErrNotExist, err)
	})

	t.Run("nil config", func(t *testing.T) {
		cfg, err := ctl.GetCache("foo")
		assert.NoError(t, err)
		assert.Equal(t, "foo", cfg.ServiceName)
	})

	t.Run("ok", func(t *testing.T) {
		cfg, err := ctl.GetCache("svc")
		assert.NoError(t, err)
		assert.Equal(t, "svc", cfg.ServiceName)
		assert.True(t, (&service.Config{Protocol: protocol.TCP}).Equal(cfg.Config))
	})
}

func TestProxyConfigsController_Add(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genProxyConfigsController(t, mockCtl, &ProxyConfig{
		ServiceName: "existSvc",
	})
	defer cancel()

	t.Run("nil", func(t *testing.T) {
		assert.NoError(t, ctl.Add(nil))
	})

	t.Run("bad config", func(t *testing.T) {
		assert.Error(t, ctl.Add(&ProxyConfig{}))
	})

	t.Run("bad config", func(t *testing.T) {
		assert.Equal(t, ErrExist, ctl.Add(&ProxyConfig{
			ServiceName: "existSvc",
		}))
	})

	t.Run("OK", func(t *testing.T) {
		assert.NoError(t, ctl.Add(&ProxyConfig{
			ServiceName: "foo",
		}))
		assert.True(t, ctl.Exist("foo"))
	})
}

func TestProxyConfigsController_Update(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genProxyConfigsController(t, mockCtl, &ProxyConfig{
		ServiceName: "existSvc",
	})
	defer cancel()

	t.Run("nil", func(t *testing.T) {
		assert.NoError(t, ctl.Update(nil))
	})

	t.Run("bad config", func(t *testing.T) {
		assert.Error(t, ctl.Update(&ProxyConfig{}))
	})

	t.Run("not exist", func(t *testing.T) {
		assert.Equal(t, ErrNotExist, ctl.Update(&ProxyConfig{
			ServiceName: "foo",
		}))
	})

	t.Run("OK", func(t *testing.T) {
		assert.NoError(t, ctl.Update(&ProxyConfig{
			ServiceName: "existSvc",
		}))
	})
}

func TestProxyConfigsController_Delete(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genProxyConfigsController(t, mockCtl)
	defer cancel()

	assert.NoError(t, ctl.Add(&ProxyConfig{
		ServiceName: "foo",
	}))
	assert.True(t, ctl.Exist("foo"))
	assert.NoError(t, ctl.Delete("foo"))
	assert.False(t, ctl.Exist("foo"))
}

func TestProxyConfigsController_GetAll(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	t.Run("not exist", func(t *testing.T) {
		ctl, cancel := genProxyConfigsController(t, mockCtl)
		defer cancel()
		_, err := ctl.GetAll()
		assert.Equal(t, ErrNotExist, err)
	})
	t.Run("OK", func(t *testing.T) {
		expectConfigs := ProxyConfigs{
			{
				ServiceName: "svc_1",
			},
			{
				ServiceName: "svc_2",
			},
		}
		ctl, cancel := genProxyConfigsController(t, mockCtl, expectConfigs...)
		defer cancel()
		configs, err := ctl.GetAll()
		assert.NoError(t, err)
		assert.ElementsMatch(t, expectConfigs, configs)
	})
}

func TestProxyConfigsController_RegisterEventHandlerWithAdd(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genProxyConfigsController(t, mockCtl)
	defer cancel()

	assert.NoError(t, ctl.Add(&ProxyConfig{
		ServiceName: "svc",
		Config:      nil,
	}))

	done := make(chan struct{})
	ctl.RegisterEventHandler(func(event *ProxyConfigEvent) {
		assert.Equal(t, &ProxyConfigEvent{
			Type: EventAdd,
			ProxyConfig: &ProxyConfig{
				ServiceName: "svc",
				Config:      nil,
			},
		}, event)
		close(done)
	})

	select {
	case <-time.NewTimer(time.Second).C:
		t.Fatal("timeout")
	case <-done:
	}
}

func TestProxyConfigsController_RegisterEventHandlerWithUpdate(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genProxyConfigsController(t, mockCtl, &ProxyConfig{
		ServiceName: "svc",
		Config:      nil,
	})
	defer cancel()

	assert.NoError(t, ctl.Update(&ProxyConfig{
		ServiceName: "svc",
		Config: &service.Config{
			Listener: &service.Listener{
				Address: &common.Address{
					Ip:   "1.1.1.1",
					Port: 80,
				},
			},
			Protocol: protocol.TCP,
		},
	}))

	done := make(chan struct{})
	ctl.RegisterEventHandler(func(event *ProxyConfigEvent) {
		assert.Equal(t, &ProxyConfigEvent{
			Type: EventUpdate,
			ProxyConfig: &ProxyConfig{
				ServiceName: "svc",
				Config: &service.Config{
					Listener: &service.Listener{
						Address: &common.Address{
							Ip:   "1.1.1.1",
							Port: 80,
						},
					},
					Protocol: protocol.TCP,
				},
			},
		}, event)
		close(done)
	})

	select {
	case <-time.NewTimer(time.Second).C:
		t.Fatal("timeout")
	case <-done:
	}
}

func TestProxyConfigsController_RegisterEventHandlerWithDel(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genProxyConfigsController(t, mockCtl, &ProxyConfig{
		ServiceName: "svc",
		Config:      nil,
	})
	defer cancel()

	assert.NoError(t, ctl.Delete("svc"))

	done := make(chan struct{})
	ctl.RegisterEventHandler(func(event *ProxyConfigEvent) {
		assert.Equal(t, &ProxyConfigEvent{
			Type: EventDelete,
			ProxyConfig: &ProxyConfig{
				ServiceName: "svc",
				Config:      nil,
			},
		}, event)
		close(done)
	})

	select {
	case <-time.NewTimer(time.Second).C:
		t.Fatal("timeout")
	case <-done:
	}
}

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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/samaritan-proxy/samaritan/pb/common"
	"github.com/samaritan-proxy/samaritan/pb/config/service"
	"github.com/stretchr/testify/assert"
)

func assertNotTimeout(t *testing.T, fn func(), timeout time.Duration) {
	timer := time.NewTimer(timeout)
	done := make(chan struct{})
	go func() {
		fn()
		close(done)
	}()
	select {
	case <-done:
	case <-timer.C:
		t.Fatalf("exec timeout")
	}
}

func TestController_fetchAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockStore(ctrl)
	s.EXPECT().GetKeys(NamespaceService, TypeServiceDependence).Return([]string{"key"}, nil)
	s.EXPECT().GetKeys(
		gomock.Not(gomock.Eq(NamespaceService)),
		gomock.Any(),
	).Return(nil, ErrNamespaceNotExist).AnyTimes()
	s.EXPECT().GetKeys(
		NamespaceService,
		gomock.Not(gomock.Eq(TypeServiceDependence)),
	).Return(nil, ErrTypeNotExist).AnyTimes()
	s.EXPECT().Get(NamespaceService, TypeServiceDependence, "key").Return([]byte("value"), nil)

	c := NewController(s, time.Second)
	cfgs, err := c.fetchAll()
	assert.NoError(t, err)
	cfg := newRawConf(NamespaceService, TypeServiceDependence, "key", []byte("value"))
	assert.Equal(t, map[uint32]*rawConf{cfg.Hashcode(): cfg}, cfgs)
}

func TestController_diff(t *testing.T) {
	c := NewController(nil, time.Second)
	cfg1 := newRawConf("ns", "type", "k1", []byte("hello"))
	cfg2 := newRawConf("ns", "type", "k2", []byte("hello"))
	c.storeCache(map[uint32]*rawConf{
		cfg1.Hashcode(): cfg1,
		cfg2.Hashcode(): cfg2,
	})
	newCfg1 := newRawConf("ns", "type", "k1", []byte("hi"))
	newCfg3 := newRawConf("ns", "type", "k3", []byte("hello"))

	var (
		add, update, del bool
	)
	c.cfgAddHdl = func(namespace, typ, key string, value []byte) {
		add = true
	}
	c.cfgUpdateHdl = func(namespace, typ, key string, value []byte) {
		update = true
	}
	c.cfgDelHdl = func(namespace, typ, key string) {
		del = true
	}

	c.diff(map[uint32]*rawConf{
		newCfg1.Hashcode(): newCfg1,
		newCfg3.Hashcode(): newCfg3,
	})

	assert.True(t, add)
	assert.True(t, update)
	assert.True(t, del)
}

func TestController_TrySubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockStore(ctrl)
	c := NewController(s, time.Second)
	assert.NoError(t, c.trySubscribe())

	ss := NewMockSubscribableStore(ctrl)
	ss.EXPECT().Subscribe(gomock.Any()).Return(nil).Times(1)
	c = NewController(ss, time.Second)
	assert.NoError(t, c.trySubscribe("namespace"))
}

func TestController_GetAndExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockStore(ctrl)
	s.EXPECT().Start().Return(nil)
	s.EXPECT().Stop()
	s.EXPECT().GetKeys(NamespaceService, TypeServiceDependence).Return([]string{"key"}, nil)
	s.EXPECT().GetKeys(
		gomock.Not(gomock.Eq(NamespaceService)),
		gomock.Any(),
	).Return(nil, ErrNamespaceNotExist).AnyTimes()
	s.EXPECT().GetKeys(
		NamespaceService,
		gomock.Not(gomock.Eq(TypeServiceDependence)),
	).Return(nil, ErrTypeNotExist).AnyTimes()
	s.EXPECT().Get(NamespaceService, TypeServiceDependence, "key").Return([]byte("value"), nil)
	c := NewController(s, 2*time.Second)
	assert.NoError(t, c.Start())
	defer func() {
		done := make(chan struct{})
		go func() {
			c.Stop()
			close(done)
		}()
		assertNotTimeout(t, func() {
			<-done
		}, time.Second)
	}()
	b, err := c.get(NamespaceService, TypeServiceDependence, "key")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value"), b)

	assert.True(t, c.exist(NamespaceService, TypeServiceDependence, "key"))
}

func TestController_Set(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockStore(ctrl)
	s.EXPECT().Set("namespace", "type", "key", []byte("value")).Return(nil)
	c := NewController(s, time.Second)
	assert.NoError(t, c.set("namespace", "type", "key", []byte("value")))
}

func TestController_Del(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockStore(ctrl)
	s.EXPECT().Del("namespace", "type", "key").Return(nil)
	c := NewController(s, time.Second)
	assert.NoError(t, c.del("namespace", "type", "key"))
}

func TestController_TriggerUpdate(t *testing.T) {
	c := NewController(nil, time.Second)
	assert.Len(t, c.updateCh, 0)
	c.triggerUpdate()
	select {
	case <-c.updateCh:
	default:
		t.Fatal("updateCh is empty")
	}
}

func TestController_RegisterEventHandler(t *testing.T) {
	c := NewController(nil, time.Second)
	assert.Empty(t, c.svcCfgEvtHdls)
	c.RegisterEventHandler(func(event *SvcConfigEvent) {})
	assert.Len(t, c.svcCfgEvtHdls, 1)
}

func TestController_HandleCfgUpdate(t *testing.T) {
	c := NewController(nil, time.Second)
	c.RegisterEventHandler(func(event *SvcConfigEvent) {
		assert.Equal(t, EventUpdate, event.Type)
		assert.Equal(t, "service", event.Config.Service)
	})
	c.handleCfgUpdate(NamespaceService, TypeServiceProxyConfig, "service", nil)
}

func TestController_HandleCfgAdd(t *testing.T) {
	c := NewController(nil, time.Second)
	c.RegisterEventHandler(func(event *SvcConfigEvent) {
		assert.Equal(t, EventAdd, event.Type)
		assert.Equal(t, "service", event.Config.Service)
	})
	c.handleCfgAdd(NamespaceService, TypeServiceProxyConfig, "service", nil)
}

func TestController_HandleCfgDel(t *testing.T) {
	c := NewController(nil, time.Second)
	c.RegisterEventHandler(func(event *SvcConfigEvent) {
		assert.Equal(t, EventDelete, event.Type)
		assert.Equal(t, "service", event.Config.Service)
	})
	c.handleCfgDel(NamespaceService, TypeServiceProxyConfig, "service")
}

func TestController_Trigger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Store", func(t *testing.T) {
		s := NewMockStore(ctrl)
		c := NewController(s, 500*time.Millisecond)
		c.wg.Add(1)
		assert.Empty(t, c.updateCh)
		go func() {
			c.trigger()
		}()
		time.Sleep(600 * time.Millisecond)
		assert.Len(t, c.updateCh, 1)
		close(c.stop)
		assertNotTimeout(t, func() {
			c.wg.Wait()
		}, time.Millisecond)
	})

	t.Run("SubscribableStore", func(t *testing.T) {
		ch := make(chan struct{}, 1)
		s := NewMockSubscribableStore(ctrl)
		s.EXPECT().Event().Return(ch)

		c := NewController(s, time.Second)
		c.wg.Add(1)
		assert.Empty(t, c.updateCh)
		go func() {
			c.trigger()
		}()
		ch <- struct{}{}
		time.Sleep(time.Millisecond)
		assert.Len(t, c.updateCh, 1)
		close(c.stop)
		assertNotTimeout(t, func() {
			c.wg.Wait()
		}, time.Millisecond)
	})
}

func TestController_GetSvcConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockStore(ctrl)
	s.EXPECT().Start().Return(nil)
	s.EXPECT().Stop()
	s.EXPECT().GetKeys(NamespaceService, TypeServiceProxyConfig).Return([]string{"svc"}, nil)
	s.EXPECT().GetKeys(
		gomock.Not(gomock.Eq(NamespaceService)),
		gomock.Any(),
	).Return(nil, ErrNamespaceNotExist).AnyTimes()
	s.EXPECT().GetKeys(
		NamespaceService,
		gomock.Not(gomock.Eq(TypeServiceProxyConfig)),
	).Return(nil, ErrTypeNotExist).AnyTimes()
	cfg := &service.Config{
		Listener: &service.Listener{
			Address: &common.Address{
				Ip:   "0.0.0.0",
				Port: 54321,
			},
		},
	}
	b, err := cfg.Marshal()
	assert.NoError(t, err)
	s.EXPECT().Get(NamespaceService, TypeServiceProxyConfig, "svc").Return(b, nil)

	c := NewController(s, time.Second)
	assert.NoError(t, c.Start())
	defer assertNotTimeout(t, c.Stop, time.Second)

	svcCfg, err := c.GetSvcConfig("svc")
	assert.NoError(t, err)
	assert.Equal(t, "svc", svcCfg.Service)
	assert.True(t, svcCfg.Config.Equal(cfg))
}

func TestController_GetSvcDependence(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockStore(ctrl)
	s.EXPECT().Start().Return(nil)
	s.EXPECT().Stop()
	s.EXPECT().GetKeys(NamespaceService, TypeServiceDependence).Return([]string{"svc"}, nil)
	s.EXPECT().GetKeys(
		gomock.Not(gomock.Eq(NamespaceService)),
		gomock.Any(),
	).Return(nil, ErrNamespaceNotExist).AnyTimes()
	s.EXPECT().GetKeys(
		NamespaceService,
		gomock.Not(gomock.Eq(TypeServiceDependence)),
	).Return(nil, ErrTypeNotExist).AnyTimes()
	s.EXPECT().Get(NamespaceService, TypeServiceDependence, "svc").Return([]byte("svc1,svc2,svc3"), nil)

	c := NewController(s, time.Second)
	assert.NoError(t, c.Start())
	defer assertNotTimeout(t, c.Stop, time.Second)

	deps, err := c.GetSvcDependence("svc")
	assert.NoError(t, err)
	assert.Equal(t, "svc", deps.Service)
	assert.Equal(t, []string{"svc1", "svc2", "svc3"}, deps.Dependencies)
}

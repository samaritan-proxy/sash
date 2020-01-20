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
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func genMockStore(t *testing.T, mockCtl *gomock.Controller, dependencies Dependencies, configs ProxyConfigs, instances Instances) *MockStore {
	lock := new(sync.RWMutex)
	cache := NewCache()
	for _, item := range dependencies {
		b, err := json.Marshal(item.Dependencies)
		assert.NoError(t, err)
		cache.Set(NamespaceService, TypeServiceDependency, item.ServiceName, b)
	}
	for _, item := range configs {
		var b []byte
		if item.Config != nil {
			_b, err := item.Config.MarshalJSON()
			assert.NoError(t, err)
			b = _b
		}
		cache.Set(NamespaceService, TypeServiceProxyConfig, item.ServiceName, b)
	}
	for _, item := range instances {
		b, err := json.Marshal(item)
		assert.NoError(t, err)
		cache.Set(NamespaceSamaritan, TypeSamaritanInstance, item.ID, b)
	}

	store := NewMockStore(mockCtl)
	store.EXPECT().Start().Return(nil).AnyTimes()
	store.EXPECT().Stop().AnyTimes()
	store.EXPECT().GetKeys(gomock.Any(), gomock.Any()).DoAndReturn(func(ns, typ string) ([]string, error) {
		lock.RLock()
		defer lock.RUnlock()
		return cache.Keys(ns, typ)
	}).AnyTimes()
	store.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ns, typ, key string) ([]byte, error) {
		lock.RLock()
		defer lock.RUnlock()
		return cache.Get(ns, typ, key)
	}).AnyTimes()
	store.EXPECT().Exist(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ns, typ, key string) bool {
		lock.RLock()
		defer lock.RUnlock()
		_, err := cache.Get(ns, typ, key)
		return err == nil
	}).AnyTimes()
	store.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ns, typ, key string, value []byte) error {
		lock.Lock()
		defer lock.Unlock()
		_, err := cache.Get(ns, typ, key)
		if err == nil {
			return ErrExist
		}
		cache.Set(ns, typ, key, value)
		return nil
	}).AnyTimes()
	store.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ns, typ, key string, value []byte) error {
		lock.Lock()
		defer lock.Unlock()
		_, err := cache.Get(ns, typ, key)
		if err != nil {
			return err
		}
		cache.Set(ns, typ, key, value)
		return nil
	}).AnyTimes()
	store.EXPECT().Del(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ns, typ, key string) error {
		lock.Lock()
		defer lock.Unlock()
		return cache.Del(ns, typ, key)
	}).AnyTimes()
	return store
}

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

func TestController_FetchAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockStore(ctrl)
	s.EXPECT().GetKeys(NamespaceService, TypeServiceDependency).Return([]string{"key"}, nil)
	s.EXPECT().GetKeys(
		gomock.Not(gomock.Eq(NamespaceService)),
		gomock.Any(),
	).Return(nil, ErrNotExist).AnyTimes()
	s.EXPECT().GetKeys(
		NamespaceService,
		gomock.Not(gomock.Eq(TypeServiceDependency)),
	).Return(nil, ErrNotExist).AnyTimes()
	s.EXPECT().Get(NamespaceService, TypeServiceDependency, "key").Return([]byte("value"), nil)

	c := NewController(s)
	all, err := c.fetchAll()
	assert.NoError(t, err)
	v, err := all.Get(NamespaceService, TypeServiceDependency, "key")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value"), v)
}

func TestController_FetchAllWithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// mock backoff max retries
	oldMaxRetries := defaultBackoffMaxRetries
	defaultBackoffMaxRetries = 2
	defer func() { defaultBackoffMaxRetries = oldMaxRetries }()

	t.Run("GetKeys", func(t *testing.T) {
		s := NewMockStore(ctrl)
		s.EXPECT().GetKeys(gomock.Any(), gomock.Any()).Return(nil, errors.New("err")).AnyTimes()
		c := NewController(s)
		_, err := c.fetchAll()
		assert.Error(t, err)
	})

	t.Run("Get", func(t *testing.T) {
		s := NewMockStore(ctrl)
		s.EXPECT().GetKeys(gomock.Any(), gomock.Any()).Return([]string{"key"}, nil)
		s.EXPECT().Get(NamespaceService, TypeServiceProxyConfig, "key").Return(nil, errors.New("err")).AnyTimes()
		c := NewController(s)
		_, err := c.fetchAll()
		assert.Error(t, err)
	})

}

func TestController_Diff(t *testing.T) {
	c := NewController(nil)
	cache := NewCache()
	cache.Set("ns", "type", "k", []byte("value"))
	cache.Set("ns", "type", "k1", []byte("hello"))
	cache.Set("ns", "type", "k2", []byte("hello"))
	c.storeCache(cache)

	var add, update, del int
	c.RegisterEventHandler(func(event *Event) {
		switch event.Type {
		case EventAdd:
			add += 1
		case EventUpdate:
			update += 1
		case EventDelete:
			del += 1
		}
	})

	cache = NewCache()
	cache.Set("ns", "type", "k", []byte("value"))
	cache.Set("ns", "type", "k1", []byte("hi"))
	cache.Set("ns", "type", "k3", []byte("hello"))
	c.diffCache(cache)

	assert.Equal(t, 1, add)
	assert.Equal(t, 1, update)
	assert.Equal(t, 1, del)
}

func TestController_TrySubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockStore(ctrl)
	c := NewController(s)
	assert.NoError(t, c.trySubscribe())

	ss := NewMockSubscribableStore(ctrl)
	ss.EXPECT().Subscribe("namespace").Return(nil).Times(1)
	c = NewController(ss)
	assert.NoError(t, c.trySubscribe("namespace"))

	ss.EXPECT().Subscribe("bad_path").Return(errors.New("test")).Times(1)
	assert.Error(t, c.trySubscribe("bad_path", "foo"))
}

func TestController_GetAndExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockStore(ctrl)
	s.EXPECT().Start().Return(nil)
	s.EXPECT().Stop()
	s.EXPECT().GetKeys(NamespaceService, TypeServiceDependency).Return([]string{"key"}, nil).AnyTimes()
	s.EXPECT().GetKeys(
		gomock.Not(gomock.Eq(NamespaceService)),
		gomock.Any(),
	).Return(nil, ErrNotExist).AnyTimes()
	s.EXPECT().GetKeys(
		NamespaceService,
		gomock.Not(gomock.Eq(TypeServiceDependency)),
	).Return(nil, ErrNotExist).AnyTimes()
	s.EXPECT().Get(NamespaceService, TypeServiceDependency, "key").Return([]byte("value"), nil).AnyTimes()
	c := NewController(s)
	assert.NoError(t, c.Start())
	// wait Controller init
	time.Sleep(time.Millisecond)
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
	b, err := c.Get(NamespaceService, TypeServiceDependency, "key")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value"), b)

	assert.True(t, c.Exist(NamespaceService, TypeServiceDependency, "key"))
}

func TestController_Add(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockStore(ctrl)
	s.EXPECT().Start().Return(nil)
	s.EXPECT().Stop()
	s.EXPECT().GetKeys(
		gomock.Any(),
		gomock.Any(),
	).Return(nil, ErrNotExist).AnyTimes()
	s.EXPECT().Add(NamespaceService, TypeServiceDependency, "key", []byte("value")).Return(nil)
	c := NewController(s)
	assert.NoError(t, c.Start())
	// wait Controller init
	time.Sleep(time.Millisecond)
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
	assert.NoError(t, c.Add(NamespaceService, TypeServiceDependency, "key", []byte("value")))
}

func TestController_Del(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockStore(ctrl)
	s.EXPECT().Start().Return(nil)
	s.EXPECT().Stop()
	s.EXPECT().GetKeys(NamespaceService, TypeServiceDependency).Return([]string{"key"}, nil)
	s.EXPECT().GetKeys(
		gomock.Not(gomock.Eq(NamespaceService)),
		gomock.Any(),
	).Return(nil, ErrNotExist).AnyTimes()
	s.EXPECT().GetKeys(
		NamespaceService,
		gomock.Not(gomock.Eq(TypeServiceDependency)),
	).Return(nil, ErrNotExist).AnyTimes()
	s.EXPECT().Get(NamespaceService, TypeServiceDependency, "key").Return([]byte("value"), nil)
	s.EXPECT().Del(NamespaceService, TypeServiceDependency, "key").Return(nil)
	c := NewController(s)
	assert.NoError(t, c.Start())
	// wait Controller init
	time.Sleep(time.Millisecond)
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
	assert.NoError(t, c.Del(NamespaceService, TypeServiceDependency, "key"))
}

func TestController_TriggerUpdate(t *testing.T) {
	c := NewController(nil)
	assert.Len(t, c.updateCh, 0)
	c.triggerUpdate()
	select {
	case <-c.updateCh:
	default:
		t.Fatal("updateCh is empty")
	}
}

func TestController_Trigger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Store", func(t *testing.T) {
		s := NewMockStore(ctrl)
		c := NewController(s, SyncInterval(500*time.Millisecond))
		c.wg.Add(1)
		assert.Empty(t, c.updateCh)
		go func() {
			c.triggerLoop()
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

		c := NewController(s)
		c.wg.Add(1)
		assert.Empty(t, c.updateCh)
		go func() {
			c.triggerLoop()
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

func TestController_KeysCached(t *testing.T) {
	c := NewController(nil)
	cache := NewCache()
	cache.Set("ns", "type", "k", []byte("value"))
	cache.Set("ns", "type", "k1", []byte("hello"))
	c.storeCache(cache)

	t.Run("bad namespace", func(t *testing.T) {
		_, err := c.KeysCached("foo", "type")
		assert.Equal(t, ErrNotExist, err)
	})

	t.Run("bad type", func(t *testing.T) {
		_, err := c.KeysCached("ns", "foo")
		assert.Equal(t, ErrNotExist, err)
	})

	t.Run("correct", func(t *testing.T) {
		keys, err := c.KeysCached("ns", "type")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"k", "k1"}, keys)
	})
}

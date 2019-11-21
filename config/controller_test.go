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
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
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

func TestController_FetchAll(t *testing.T) {
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

	c := NewController(s)
	all, err := c.fetchAll()
	assert.NoError(t, err)
	v, err := all.Get(NamespaceService, TypeServiceDependence, "key")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value"), v)
}

func TestController_FetchAllWithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
	c.diff(cache)

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
	b, err := c.Get(NamespaceService, TypeServiceDependence, "key")
	assert.NoError(t, err)
	assert.Equal(t, []byte("value"), b)

	assert.True(t, c.Exist(NamespaceService, TypeServiceDependence, "key"))
}

func TestController_GetWithError(t *testing.T) {
	c := NewController(nil)
	cache := NewCache()
	cache.Set("ns", "type", "key", []byte("value"))
	c.storeCache(cache)
	t.Run("bad ns", func(t *testing.T) {
		_, err := c.Get("foo", "type", "key")
		assert.Equal(t, ErrNamespaceNotExist, err)
	})
	t.Run("bad type", func(t *testing.T) {
		_, err := c.Get("ns", "foo", "key")
		assert.Equal(t, ErrTypeNotExist, err)
	})
	t.Run("bad key", func(t *testing.T) {
		_, err := c.Get("ns", "type", "foo")
		assert.Equal(t, ErrKeyNotExist, err)
	})
}

func TestController_Set(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewMockStore(ctrl)
	s.EXPECT().Start().Return(nil)
	s.EXPECT().Stop()
	s.EXPECT().GetKeys(
		gomock.Any(),
		gomock.Any(),
	).Return(nil, ErrNamespaceNotExist).AnyTimes()
	s.EXPECT().Set(NamespaceService, TypeServiceDependence, "key", []byte("value")).Return(nil)
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
	assert.NoError(t, c.Set(NamespaceService, TypeServiceDependence, "key", []byte("value")))
	assert.True(t, c.Exist(NamespaceService, TypeServiceDependence, "key"))
}

func TestController_Del(t *testing.T) {
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
	s.EXPECT().Del(NamespaceService, TypeServiceDependence, "key").Return(nil)
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
	assert.NoError(t, c.Del(NamespaceService, TypeServiceDependence, "key"))
	assert.False(t, c.Exist(NamespaceService, TypeServiceDependence, "key"))
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

func TestController_RegisterEventHandler(t *testing.T) {
	c := NewController(nil)
	assert.Empty(t, c.evtHdls)
	c.RegisterEventHandler(func(event *Event) {})
	assert.Len(t, c.evtHdls, 1)
}

func TestController_Trigger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Store", func(t *testing.T) {
		s := NewMockStore(ctrl)
		c := NewController(s, Interval(500*time.Millisecond))
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

func TestController_Keys(t *testing.T) {
	c := NewController(nil)
	cache := NewCache()
	cache.Set("ns", "type", "k", []byte("value"))
	cache.Set("ns", "type", "k1", []byte("hello"))
	c.storeCache(cache)

	t.Run("bad namespace", func(t *testing.T) {
		_, err := c.Keys("foo", "type")
		assert.Equal(t, ErrNamespaceNotExist, err)
	})

	t.Run("bad type", func(t *testing.T) {
		_, err := c.Keys("ns", "foo")
		assert.Equal(t, ErrTypeNotExist, err)
	})

	t.Run("correct", func(t *testing.T) {
		keys, err := c.Keys("ns", "type")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"k", "k1"}, keys)
	})
}

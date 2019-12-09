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

package registry

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/samaritan-proxy/sash/model"
	"github.com/samaritan-proxy/sash/registry/memory"
)

func TestSetSyncFreqOption(t *testing.T) {
	opts := defaultCacheOptions()

	freq := time.Second * 30
	o := SyncFreq(freq)
	o(opts)
	assert.Equal(t, freq, opts.syncFreq)
}

func TestSetSyncJitterOption(t *testing.T) {
	opts := defaultCacheOptions()

	jitter := 0.5
	o := SyncJitter(jitter)
	o(opts)
	assert.Equal(t, jitter, opts.syncJitter)
}

func TestCacheSyncFail(t *testing.T) {
	t.Run("get services", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		r := NewMockServiceRegistry(ctrl)
		r.EXPECT().
			List().
			Return(nil, errors.New("internal error"))

		c := newCache(r)
		err := c.Sync(context.TODO())
		assert.Error(t, err)
	})

	t.Run("retry timeout", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// decrease the max retry interval
		oldMaxInterval := defaultBackoffMaxInterval
		defer func() { defaultBackoffMaxInterval = oldMaxInterval }()
		defaultBackoffMaxInterval = 150 * time.Millisecond

		r := NewMockServiceRegistry(ctrl)
		r.EXPECT().List().Return([]string{"foo"}, nil)
		r.EXPECT().
			Get("foo").
			Return(nil, errors.New("internal error")).
			MaxTimes(defaultBackoffMaxRetries + 1)

		c := newCache(r)
		err := c.Sync(context.TODO())
		assert.Error(t, err)
	})

	t.Run("cancel when retrying", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var retries = new(int)
		r := NewMockServiceRegistry(ctrl)
		r.EXPECT().List().Return([]string{"foo"}, nil)
		r.EXPECT().
			Get("foo").
			DoAndReturn(func(string) (*model.Service, error) {
				*retries = *retries + 1
				return nil, errors.New("internal error")
			}).
			MaxTimes(defaultBackoffMaxRetries + 1)

		c := newCache(r)
		ctx, cancel := context.WithCancel(context.TODO())
		time.AfterFunc(time.Second, cancel)
		err := c.Sync(ctx)
		assert.Error(t, err)
		assert.True(t, *retries < defaultBackoffMaxRetries+1)
	})
}

func TestCacheSyncSuccess(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		r := NewMockServiceRegistry(ctrl)
		r.EXPECT().
			List().
			Return([]string{"foo"}, nil)
		r.EXPECT().
			Get("foo").
			Return(model.NewService("foo"), nil)

		c := newCache(r)
		err := c.Sync(context.TODO())
		assert.NoError(t, err)
		assert.Len(t, c.services, 1)
		assert.NotNil(t, c.services["foo"])
	})

	t.Run("auto retry", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		r := NewMockServiceRegistry(ctrl)
		r.EXPECT().
			List().
			Return([]string{"foo", "bar", "zoo"}, nil)
		maxRetries := defaultBackoffMaxRetries / 2
		retries := make(map[string]int)
		f := func(name string) (*model.Service, error) {
			if retries[name] == maxRetries {
				return model.NewService(name), nil
			}
			retries[name]++
			return nil, errors.New("server is busy")
		}
		r.EXPECT().Get("foo").DoAndReturn(f).AnyTimes()
		r.EXPECT().Get("bar").DoAndReturn(f).AnyTimes()
		r.EXPECT().Get("zoo").DoAndReturn(f).AnyTimes()

		// decrease the max retry interval
		oldMaxInterval := defaultBackoffMaxInterval
		defer func() { defaultBackoffMaxInterval = oldMaxInterval }()
		defaultBackoffMaxInterval = 150 * time.Millisecond

		c := newCache(r)
		err := c.Sync(context.TODO())
		assert.NoError(t, err)
		assert.Len(t, c.services, 3)
		assert.NotNil(t, c.services["foo"])
		assert.NotNil(t, c.services["bar"])
		assert.NotNil(t, c.services["zoo"])
	})
}

func findServiceEvent(all []*ServiceEvent, target *ServiceEvent) bool {
	for _, event := range all {
		if !reflect.DeepEqual(event, target) {
			continue
		}
		return true
	}
	return false
}

func findInstanceEvent(all []*InstanceEvent, target *InstanceEvent) bool {
	for _, event := range all {
		if !reflect.DeepEqual(event, target) {
			continue
		}
		return true
	}
	return false
}

func TestCacheOmitServiceEvents(t *testing.T) {
	svc1 := model.NewService("svc1")
	svc2 := model.NewService(
		"svc2",
		model.NewServiceInstance("127.0.0.1", 8888),
		model.NewServiceInstance("127.0.0.1", 8889),
	)
	r := memory.NewRegistry(svc1, svc2)

	c := newCache(r)
	err := c.Sync(context.TODO())
	assert.NoError(t, err)

	// register handlers
	svcEvts := make([]*ServiceEvent, 0, 32)
	svcHandler := func(event *ServiceEvent) {
		svcEvts = append(svcEvts, event)
	}
	c.RegisterServiceEventHandler(svcHandler)

	// simulate service changes
	// delete svc1
	r.Deregister("svc1")
	// add svc3
	svc3 := model.NewService(
		"svc3",
		model.NewServiceInstance("127.0.0.1", 9999),
	)
	r.Register(svc3)

	err = c.Sync(context.TODO())
	assert.NoError(t, err)
	// assert events
	assert.Equal(t, 2, len(svcEvts))
	assert.True(t, findServiceEvent(
		svcEvts,
		&ServiceEvent{
			Type:    EventDelete,
			Service: svc1,
		},
	))
	assert.True(t, findServiceEvent(
		svcEvts,
		&ServiceEvent{
			Type:    EventAdd,
			Service: svc3,
		},
	))
}

func TestCacheOmitInstanceEvents(t *testing.T) {
	inst1 := model.NewServiceInstance("127.0.0.1", 8888)
	inst2 := model.NewServiceInstance("127.0.0.1", 8889)
	svcName := "svc1"
	svc := model.NewService(svcName, inst1, inst2)
	r := memory.NewRegistry(svc)

	c := newCache(r)
	err := c.Sync(context.TODO())
	assert.NoError(t, err)

	// register handlers
	instEvts := make([]*InstanceEvent, 0, 32)
	instHandler := func(event *InstanceEvent) {
		instEvts = append(instEvts, event)
	}
	c.RegisterInstanceEventHandler(instHandler)

	// simulate instances changes
	// delete inst1
	r.DeleteInstance("svc1", inst1)
	// update inst2
	inst2.State = model.StateUnhealthy
	// add inst3
	inst3 := model.NewServiceInstance("127.0.0.1", 9999)
	r.AddInstance("svc1", inst3)

	err = c.Sync(context.TODO())
	assert.NoError(t, err)

	// assert events
	assert.Equal(t, 3, len(instEvts))
	assert.True(t, findInstanceEvent(
		instEvts,
		&InstanceEvent{
			Type:        EventDelete,
			ServiceName: svcName,
			Instances:   []*model.ServiceInstance{inst1},
		},
	))
	assert.True(t, findInstanceEvent(
		instEvts,
		&InstanceEvent{
			Type:        EventUpdate,
			ServiceName: svcName,
			Instances:   []*model.ServiceInstance{inst2},
		},
	))
	assert.True(t, findInstanceEvent(
		instEvts,
		&InstanceEvent{
			Type:        EventAdd,
			ServiceName: svcName,
			Instances:   []*model.ServiceInstance{inst3},
		},
	))
}

func TestCachePeriodicSync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := NewMockServiceRegistry(ctrl)
	// fist is failed, the subsequent are successful.
	r.EXPECT().List().Return(nil, errors.New("server is busy"))
	r.EXPECT().List().Return([]string{"foo"}, nil).AnyTimes()
	r.EXPECT().Get("foo").Return(
		model.NewService("foo"), nil,
	).AnyTimes()

	c := newCache(r, SyncFreq(time.Second))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		c.Run(ctx)
		close(done)
	}()
	time.AfterFunc(time.Second*2, cancel)
	<-done
	// assert
	names, err := c.List()
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo"}, names)
	svc, err := c.Get("foo")
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

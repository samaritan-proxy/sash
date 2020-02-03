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

//import (
//	"fmt"
//	"sort"
//	"testing"
//	"time"
//
//	"github.com/golang/mock/gomock"
//	"github.com/stretchr/testify/assert"
//)
//
//func TestInstance_Verify(t *testing.T) {
//	cases := []struct {
//		Instance *Instance
//		IsError  bool
//	}{
//		{Instance: new(Instance), IsError: true},
//		{Instance: &Instance{ID: "foo"}, IsError: false},
//	}
//	for idx, c := range cases {
//		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
//			assert.Equal(t, c.IsError, c.Instance.Verify() != nil)
//		})
//	}
//}
//
//func TestSortInstances(t *testing.T) {
//	instances := Instances{{ID: "b"}, {ID: "a"}, {ID: "c"}}
//	sort.Sort(instances)
//	assert.Equal(t, Instances{{ID: "a"}, {ID: "b"}, {ID: "c"}}, instances)
//}
//
//func genInstancesController(t *testing.T, mockCtl *gomock.Controller, insts ...*Instance) (*InstancesController, func()) {
//	store := genMockStore(t, mockCtl, nil, nil, insts)
//	ctl := NewController(store, SyncInterval(time.Millisecond))
//	assert.NoError(t, ctl.Start())
//	time.Sleep(time.Millisecond)
//	return ctl.Instances(), ctl.Stop
//}
//
//func TestInstancesController_GetNamespace(t *testing.T) {
//	mockCtl := gomock.NewController(t)
//	defer mockCtl.Finish()
//
//	ctl, cancel := genInstancesController(t, mockCtl)
//	defer cancel()
//
//	assert.Equal(t, NamespaceSamaritan, ctl.getNamespace())
//}
//
//func TestInstancesController_GetType(t *testing.T) {
//	mockCtl := gomock.NewController(t)
//	defer mockCtl.Finish()
//
//	ctl, cancel := genInstancesController(t, mockCtl)
//	defer cancel()
//
//	assert.Equal(t, TypeSamaritanInstance, ctl.getType())
//}
//
//func TestInstancesController_UnmarshalInstance(t *testing.T) {
//	mockCtl := gomock.NewController(t)
//	defer mockCtl.Finish()
//
//	ctl, cancel := genInstancesController(t, mockCtl)
//	defer cancel()
//
//	cases := []struct {
//		Bytes     []byte
//		Instances *Instance
//		IsError   bool
//	}{
//		{Bytes: []byte{0}, Instances: nil, IsError: true},
//		{Bytes: []byte(`{"id": "foo"}`), Instances: &Instance{ID: "foo"}, IsError: false},
//	}
//	for idx, c := range cases {
//		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
//			inst, err := ctl.unmarshalInstance(c.Bytes)
//			if c.IsError {
//				assert.Error(t, err)
//				return
//			}
//			assert.Equal(t, c.Instances, inst)
//		})
//	}
//}
//
//func TestInstancesController_Get(t *testing.T) {
//	mockCtl := gomock.NewController(t)
//	defer mockCtl.Finish()
//
//	ctl, cancel := genInstancesController(t, mockCtl, &Instance{
//		ID: "foo",
//	})
//	defer cancel()
//
//	t.Run("not exist", func(t *testing.T) {
//		_, err := ctl.Get("bar")
//		assert.Equal(t, ErrNotExist, err)
//	})
//
//	t.Run("OK", func(t *testing.T) {
//		inst, err := ctl.Get("foo")
//		assert.NoError(t, err)
//		assert.Equal(t, "foo", inst.ID)
//	})
//}
//
//func TestInstancesController_GetCache(t *testing.T) {
//	mockCtl := gomock.NewController(t)
//	defer mockCtl.Finish()
//
//	ctl, cancel := genInstancesController(t, mockCtl, &Instance{
//		ID: "foo",
//	})
//	defer cancel()
//
//	t.Run("not exist", func(t *testing.T) {
//		_, err := ctl.GetCache("bar")
//		assert.Equal(t, ErrNotExist, err)
//	})
//
//	t.Run("OK", func(t *testing.T) {
//		inst, err := ctl.GetCache("foo")
//		assert.NoError(t, err)
//		assert.Equal(t, "foo", inst.ID)
//	})
//}
//
//func TestInstancesController_Add(t *testing.T) {
//	mockCtl := gomock.NewController(t)
//	defer mockCtl.Finish()
//
//	ctl, cancel := genInstancesController(t, mockCtl, &Instance{
//		ID: "existInst",
//	})
//	defer cancel()
//
//	t.Run("nil", func(t *testing.T) {
//		assert.NoError(t, ctl.Add(nil))
//	})
//
//	t.Run("bad instance", func(t *testing.T) {
//		assert.Error(t, ctl.Add(&Instance{}))
//	})
//
//	t.Run("exist", func(t *testing.T) {
//		assert.Equal(t, ErrExist, ctl.Add(&Instance{
//			ID: "existInst",
//		}))
//	})
//
//	t.Run("OK", func(t *testing.T) {
//		assert.NoError(t, ctl.Add(&Instance{
//			ID: "foo",
//		}))
//		assert.True(t, ctl.Exist("foo"))
//	})
//}
//
//func TestInstancesController_Update(t *testing.T) {
//	mockCtl := gomock.NewController(t)
//	defer mockCtl.Finish()
//
//	ctl, cancel := genInstancesController(t, mockCtl, &Instance{
//		ID: "existInst",
//	})
//	defer cancel()
//
//	t.Run("nil", func(t *testing.T) {
//		assert.NoError(t, ctl.Update(nil))
//	})
//
//	t.Run("bad instance", func(t *testing.T) {
//		assert.Error(t, ctl.Update(&Instance{}))
//	})
//
//	t.Run("not exist", func(t *testing.T) {
//		assert.Equal(t, ErrNotExist, ctl.Update(&Instance{
//			ID: "foo",
//		}))
//	})
//
//	t.Run("OK", func(t *testing.T) {
//		assert.NoError(t, ctl.Update(&Instance{
//			ID: "existInst",
//		}))
//	})
//}
//
//func TestInstancesController_Delete(t *testing.T) {
//	mockCtl := gomock.NewController(t)
//	defer mockCtl.Finish()
//
//	ctl, cancel := genInstancesController(t, mockCtl)
//	defer cancel()
//
//	assert.NoError(t, ctl.Add(&Instance{
//		ID: "foo",
//	}))
//	assert.True(t, ctl.Exist("foo"))
//	assert.NoError(t, ctl.Delete("foo"))
//	assert.False(t, ctl.Exist("foo"))
//}
//
//func TestInstancesController_GetAll(t *testing.T) {
//	mockCtl := gomock.NewController(t)
//	defer mockCtl.Finish()
//
//	t.Run("not exist", func(t *testing.T) {
//		ctl, cancel := genInstancesController(t, mockCtl)
//		defer cancel()
//		_, err := ctl.GetAll()
//		assert.Equal(t, ErrNotExist, err)
//	})
//	t.Run("OK", func(t *testing.T) {
//		expectInstances := Instances{
//			{
//				ID: "instance_1",
//			},
//			{
//				ID: "instance_2",
//			},
//		}
//		ctl, cancel := genInstancesController(t, mockCtl, expectInstances...)
//		defer cancel()
//		instances, err := ctl.GetAll()
//		assert.NoError(t, err)
//		assert.ElementsMatch(t, expectInstances, instances)
//	})
//}

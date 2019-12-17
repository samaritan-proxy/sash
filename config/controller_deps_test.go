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
	"github.com/stretchr/testify/assert"
)

func TestDependency_Verify(t *testing.T) {
	cases := []struct {
		Dependency *Dependency
		IsError    bool
	}{
		{
			Dependency: &Dependency{
				ServiceName:  "",
				Dependencies: nil,
			},
			IsError: true,
		},
		{
			Dependency: &Dependency{
				ServiceName:  "svc",
				Dependencies: nil,
			},
			IsError: false,
		},
	}

	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			assert.Equal(t, c.IsError, c.Dependency.Verify() != nil)
		})
	}
}

func TestSortDependencies(t *testing.T) {
	deps := Dependencies{
		{ServiceName: "b"}, {ServiceName: "a"}, {ServiceName: "c"},
	}
	sort.Sort(deps)
	assert.Equal(t, Dependencies{
		{ServiceName: "a"}, {ServiceName: "b"}, {ServiceName: "c"},
	}, deps)
}

func genDependenciesController(t *testing.T, mockCtl *gomock.Controller, deps ...*Dependency) (*DependenciesController, func()) {
	store := genMockStore(t, mockCtl, deps, nil, nil)
	ctl := NewController(store, Interval(time.Millisecond))
	assert.NoError(t, ctl.Start())
	time.Sleep(time.Millisecond)
	return ctl.Dependencies(), ctl.Stop
}

func TestDependenciesController_GetNamespace(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	c, cancel := genDependenciesController(t, mockCtl)
	defer cancel()
	assert.Equal(t, NamespaceService, c.getNamespace())
}

func TestDependenciesController_GetType(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	c, cancel := genDependenciesController(t, mockCtl)
	defer cancel()
	assert.Equal(t, TypeServiceDependency, c.getType())
}

func TestDependenciesController_UnmarshalDependency(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	ctl, cancel := genDependenciesController(t, mockCtl)
	defer cancel()

	cases := []struct {
		Bytes   []byte
		Deps    []string
		IsError bool
	}{
		{Bytes: []byte{0}, Deps: nil, IsError: true},
		{Bytes: []byte(`["dep_1","dep_2"]`), Deps: []string{"dep_1", "dep_2"}, IsError: false},
	}
	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			deps, err := ctl.unmarshalDependency(c.Bytes)
			if c.IsError {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, c.Deps, deps)
		})
	}
}

func TestDependenciesController_Get(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	expectDep := &Dependency{
		ServiceName:  "svc",
		Dependencies: []string{"dep_1", "dep_2"},
	}
	ctl, cancel := genDependenciesController(t, mockCtl, expectDep)
	defer cancel()

	t.Run("not exist", func(t *testing.T) {
		_, err := ctl.Get("foo")
		assert.Equal(t, ErrNotExist, err)
	})

	t.Run("OK", func(t *testing.T) {
		dep, err := ctl.Get("svc")
		assert.NoError(t, err)
		assert.Equal(t, expectDep, dep)
	})
}

func TestDependenciesController_GetCache(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genDependenciesController(t, mockCtl)
	defer cancel()

	expectDep := &Dependency{
		ServiceName:  "svc",
		Dependencies: []string{"dep_1", "dep_2"},
	}
	assert.NoError(t, ctl.Set(expectDep))

	time.Sleep(time.Millisecond)

	t.Run("not exist", func(t *testing.T) {
		_, err := ctl.GetCache("foo")
		assert.Equal(t, ErrNotExist, err)
	})

	t.Run("OK", func(t *testing.T) {
		dep, err := ctl.GetCache("svc")
		assert.NoError(t, err)
		assert.Equal(t, expectDep, dep)
	})
}

func TestDependenciesController_Set(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genDependenciesController(t, mockCtl)
	defer cancel()

	t.Run("nil", func(t *testing.T) {
		assert.NoError(t, ctl.Set(nil))
	})
	t.Run("bad dep", func(t *testing.T) {
		assert.Error(t, ctl.Set(&Dependency{}))
	})
	t.Run("ok", func(t *testing.T) {
		assert.NoError(t, ctl.Set(&Dependency{
			ServiceName:  "svc",
			Dependencies: []string{"dep"},
		}))
		assert.True(t, ctl.Exist("svc"))
	})
}

func TestDependenciesController_Delete(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genDependenciesController(t, mockCtl)
	defer cancel()

	assert.NoError(t, ctl.Set(&Dependency{
		ServiceName:  "svc",
		Dependencies: []string{"dep"},
	}))

	assert.NoError(t, ctl.Delete("svc"))
	assert.False(t, ctl.Exist("svc"))
}

func TestDependenciesController_GetAll(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	t.Run("not exist", func(t *testing.T) {
		ctl, cancel := genDependenciesController(t, mockCtl)
		defer cancel()
		_, err := ctl.GetAll()
		assert.Equal(t, ErrNotExist, err)
	})
	t.Run("OK", func(t *testing.T) {
		expectDeps := Dependencies{
			{
				ServiceName:  "svc_1",
				Dependencies: []string{"dep_1"},
			},
			{
				ServiceName:  "svc_2",
				Dependencies: []string{"dep_2"},
			},
		}
		ctl, cancel := genDependenciesController(t, mockCtl, expectDeps...)
		defer cancel()
		deps, err := ctl.GetAll()
		assert.NoError(t, err)
		assert.ElementsMatch(t, expectDeps, deps)
	})
}

func TestDependenciesController_RegisterEventHandlerWithAdd(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genDependenciesController(t, mockCtl)
	defer cancel()

	done := make(chan struct{})
	ctl.RegisterEventHandler(func(event *DependencyEvent) {
		assert.Equal(t, "svc", event.ServiceName)
		assert.Equal(t, []string{"dep_1", "dep_2"}, event.Add)
		close(done)
	})

	assert.NoError(t, ctl.Set(&Dependency{
		ServiceName:  "svc",
		Dependencies: []string{"dep_1", "dep_2"},
	}))

	select {
	case <-time.NewTimer(time.Second).C:
		t.Fatal("timeout")
	case <-done:
	}
}

func TestDependenciesController_RegisterEventHandlerWithUpdate(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genDependenciesController(t, mockCtl)
	defer cancel()

	assert.NoError(t, ctl.Set(&Dependency{
		ServiceName:  "svc",
		Dependencies: []string{"dep_1", "dep_2"},
	}))
	_, err := ctl.Get("svc")
	assert.NoError(t, err)

	// wait first cache init.
	time.Sleep(time.Millisecond * 100)

	done := make(chan struct{})
	ctl.RegisterEventHandler(func(event *DependencyEvent) {
		assert.Equal(t, &DependencyEvent{
			ServiceName: "svc",
			Add:         []string{"dep_3"},
			Del:         []string{"dep_1"},
		}, event)
		close(done)
	})

	assert.NoError(t, ctl.Set(&Dependency{
		ServiceName:  "svc",
		Dependencies: []string{"dep_2", "dep_3"},
	}))

	select {
	case <-time.NewTimer(time.Second).C:
		t.Fatal("timeout")
	case <-done:
	}
}

func TestDependenciesController_RegisterEventHandlerWithDel(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctl, cancel := genDependenciesController(t, mockCtl, &Dependency{
		ServiceName:  "svc",
		Dependencies: []string{"dep_1", "dep_2"},
	})
	defer cancel()

	_, err := ctl.Get("svc")
	assert.NoError(t, err)

	done := make(chan struct{})
	ctl.RegisterEventHandler(func(event *DependencyEvent) {
		assert.Equal(t, &DependencyEvent{
			ServiceName: "svc",
			Del:         []string{"dep_1", "dep_2"},
		}, event)
		close(done)
	})

	assert.NoError(t, ctl.Delete("svc"))

	select {
	case <-time.NewTimer(time.Second).C:
		t.Fatal("timeout")
	case <-done:
	}
}

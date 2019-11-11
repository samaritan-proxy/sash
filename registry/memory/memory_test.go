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

package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/samaritan-proxy/sash/model"
)

func TestList(t *testing.T) {
	name := "foo"
	service := model.NewService(name)
	r := NewRegistry(service)
	names, _ := r.List()
	assert.Equal(t, []string{"foo"}, names)
}

func TestGet(t *testing.T) {
	name := "foo"
	r := NewRegistry(model.NewService(name))
	service, _ := r.Get(name)
	assert.NotNil(t, service)
}

func TestRegister(t *testing.T) {
	r := NewRegistry()
	names, _ := r.List()
	assert.Len(t, names, 0)

	service := model.NewService("foo")
	r.Register(service)
	names, _ = r.List()
	assert.Len(t, names, 1)
}

func TestDeregister(t *testing.T) {
	name := "foo"
	service := model.NewService(name)
	r := NewRegistry(service)
	names, _ := r.List()
	assert.Len(t, names, 1)

	r.Deregister(name)
	names, _ = r.List()
	assert.Len(t, names, 0)
}

func TestAddInstance(t *testing.T) {
	name := "foo"
	r := NewRegistry(model.NewService(name))
	service, _ := r.Get(name)
	assert.Len(t, service.Instances, 0)

	inst := model.NewServiceInstance("1.1.1.1")
	r.AddInstance(name, inst)
	service, _ = r.Get(name)
	assert.Len(t, service.Instances, 1)
}

func TestDeleteInstance(t *testing.T) {
	inst := model.NewServiceInstance("1.1.1.1")
	name := "foo"
	r := NewRegistry(
		model.NewService(name, inst),
	)
	service, _ := r.Get(name)
	assert.Len(t, service.Instances, 1)

	r.DeleteInstance(name, inst)
	service, _ = r.Get(name)
	assert.Len(t, service.Instances, 0)
}

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
	"context"

	"github.com/samaritan-proxy/sash/model"
)

var _ model.ServiceRegistry = new(Registry)

// Register is an implementation of model.ServiceRegistry.
type Registry struct {
	services map[string]*model.Service
}

// NewRegistry creates a memory service registry.
// NOTE: It is designed for unit tests, all methods aren't goroutine-safe.
func NewRegistry(services ...*model.Service) *Registry {
	m := make(map[string]*model.Service)
	for _, service := range services {
		m[service.Name] = service
	}
	return &Registry{
		services: m,
	}
}

// Run runs the registry.
func (r *Registry) Run(ctx context.Context) {}

// List returns all registered service names.
func (r *Registry) List() ([]string, error) {
	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	return names, nil
}

// Get gets the service info with the given name.
func (r *Registry) Get(name string) (*model.Service, error) {
	service, ok := r.services[name]
	if !ok {
		return nil, nil
	}
	return service.DeepCopy(), nil
}

// Deregister deregisters a service.
func (r *Registry) Deregister(name string) bool {
	_, ok := r.services[name]
	if !ok {
		return false
	}
	delete(r.services, name)
	return true
}

// Register registers a service.
func (r *Registry) Register(service *model.Service) {
	r.services[service.Name] = service
}

// AddInstance adds some instances to the specificed service.
func (r *Registry) AddInstance(name string, instances ...*model.ServiceInstance) {
	service, ok := r.services[name]
	if !ok {
		return
	}

	for _, instance := range instances {
		addr := instance.Addr()
		service.Instances[addr] = instance
	}
}

// DeleteInstance deletes some instances from the specificed service.
func (r *Registry) DeleteInstance(name string, instances ...*model.ServiceInstance) {
	service, ok := r.services[name]
	if !ok {
		return
	}

	for _, instance := range instances {
		delete(service.Instances, instance.Addr())
	}
}

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

package model

import "context"

// Service represents a service.
type Service struct {
	Name      string
	Instances map[string]*ServiceInstance
}

// NewService creates a service.
func NewService(name string, insts ...*ServiceInstance) *Service {
	service := &Service{
		Name:      name,
		Instances: make(map[string]*ServiceInstance),
	}
	for _, inst := range insts {
		service.Instances[inst.Addr] = inst
	}
	return service
}

// ServiceInstanceState indicates the state of service instance.
type ServiceInstanceState uint8

// The following shows the available states of service instance.
const (
	StateHealthy ServiceInstanceState = iota
	StateUnhealthy
)

// ServiceInstance represents an instance of service.
type ServiceInstance struct {
	Addr  string               `json:"addr"`
	State ServiceInstanceState `json:"state"`
	Meta  map[string]string    `json:"meta"`
}

// NewServerInstance creates a plain service instance.
func NewServiceInstance(addr string) *ServiceInstance {
	return &ServiceInstance{
		Addr:  addr,
		State: StateHealthy,
		Meta:  make(map[string]string),
	}
}

// ServiceRegistry represents a service registry.
type ServiceRegistry interface {
	// Run runs the registry until a stop signal received.
	Run(ctx context.Context)

	// List returns all registered service names.
	List() ([]string, error)
	// Get gets the service info with the given name.
	Get(name string) (*Service, error)
}

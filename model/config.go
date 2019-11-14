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

import (
	"github.com/samaritan-proxy/samaritan/pb/config/service"
)

// ServiceConfig represents a config of service.
type ServiceConfig struct {
	ServiceName string
	Config      *service.Config
}

// NewServiceConfig return a new ServiceConfig.
func NewServiceConfig(name string, config *service.Config) *ServiceConfig {
	return &ServiceConfig{
		ServiceName: name,
		Config:      config,
	}
}

// ServiceConfig contain all dependencies of a services.
type ServiceDependence struct {
	Service      string
	Dependencies []string
}

// NewServiceDependence return a new ServiceDependence.
func NewServiceDependence(name string, dependencies []string) *ServiceDependence {
	return &ServiceDependence{
		Service:      name,
		Dependencies: dependencies,
	}
}

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

import "github.com/samaritan-proxy/sash/model"

// EventType indicates the type of event.
type EventType uint8

// The following shows the available event types.
const (
	EventAdd EventType = iota + 1
	EventUpdate
	EventDelete
)

// ServiceEvent represents a service event.
type ServiceEvent struct {
	Type    EventType
	Service *model.Service
}

// InstanceEvent represents a instance event.
type InstanceEvent struct {
	Type        EventType
	ServiceName string
	Instances   []*model.ServiceInstance
}

type (
	// ServiceEventHandler is used to handle the service event.
	ServiceEventHandler func(event *ServiceEvent)
	// InstanceEventHandler is used to handel isntance event.
	InstanceEventHandler func(event *InstanceEvent)
)

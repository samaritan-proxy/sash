package registry

import "github.com/samaritan-proxy/sash/model"

type EventType uint8

const (
	EventAdd EventType = iota + 1
	EventUpdate
	EventDelete
)

type ServiceEvent struct {
	Type    EventType
	Service *model.Service
}

type InstanceEvent struct {
	Type        EventType
	ServiceName string
	Instances   *model.ServiceInstance
}

type (
	ServiceEventHandler  func(event *ServiceEvent)
	InstanceEventHandler func(event *InstanceEvent)
)

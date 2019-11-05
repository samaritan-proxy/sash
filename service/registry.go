package service

// Instance represents an instance of service.
type Instance struct {
	Addr  string
	State uint8
	Meta  map[string]string
}

// EventType indicates the type of event.
type RegistryEventType uint8

// The following describes the type of registry event.
const (
	RegistryEventServiceDelete = iota + 1
	RegistryEventInstanceAdd
	RegistryEventInstanceDelete
)

// Event represents an event of service.
type RegistryEvent struct {
	Type      RegistryEventType
	Service   string
	Instances []*Instance
}

// Registry represents a service registry.
type Registry interface {
	// List returns all registerd services.
	List() ([]string, error)
	// Get gets all instances of the specified service.
	Get(service string) ([]*Instance, error)

	// GetAndSub gets all instances of the specified service, and subscribes
	// the subsequent changes.
	GetAndSub(service string) ([]*Instance, error)
	// Event returns a event channel which will publish all changes
	// of the subscribed services.
	Event() <-chan RegistryEvent
}

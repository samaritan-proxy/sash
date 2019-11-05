package registry

// Instance represents an instance of service.
type Instance struct {
	Addr  string
	State uint8
	Meta  map[string]string
}

// EventType indicates the type of event.
type EventType uint8

const (
	// EventServiceDelete indicates an event of service delete.
	EventServiceDelete = iota + 1
	// EventServiceAdd indicates an event of service's instance add.
	EventInstanceAdd
	// EventInstanceDelete indicates an event of service's instance delete.
	EventInstanceDelete
)

// Event represents an event of service.
type Event struct {
	Type      EventType
	Service   string
	Instances []*Instance
}

// Registry represents a service registry.
type Registry interface {
	// Event returns a event channel which will publish all changes
	// of the subscribed services.
	Event() <-chan Event

	// Get gets all instances of the specified service.
	Get(service string) ([]*Instance, error)
	// GetAndSub gets all instances of the specified service, and subscribes
	// the subsequent changes.
	GetAndSub(service string) ([]*Instance, error)
	// Unsubscribe unsubscribes changes of the specified service.
	Unsubscribe(service string) error
}

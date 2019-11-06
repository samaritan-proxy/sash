package model

// Service represents a service.
type Service struct {
	Name      string
	Instances map[string]*ServiceInstance
}

// ServiceInstance represents an instance of service.
type ServiceInstance struct {
	Addr  string
	State uint8
	Meta  map[string]string
}

// ServiceRegistry represents a service registry.
type ServiceRegistry interface {
	// Run starts the registry unitl a stop signal received.
	Run(stop <-chan struct{})

	// List returns all registerd service names.
	List() ([]string, error)
	// Get gets the service info with the given name.
	Get(name string) (*Service, error)
}

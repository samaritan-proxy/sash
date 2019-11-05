package memory

import "github.com/samaritan-proxy/sash/service/registry"

var _ registry.Registry = new(memRegistry)

type memRegistry struct{}

func (r *memRegistry) Event() <-chan registry.Event {
	return nil
}

func (r *memRegistry) Get(service string) ([]*registry.Instance, error) {
	return nil, nil
}

func (r *memRegistry) GetAndSub(service string) ([]*registry.Instance, error) {
	return nil, nil
}

func (r *memRegistry) Unsubscribe(service string) error {
	return nil
}

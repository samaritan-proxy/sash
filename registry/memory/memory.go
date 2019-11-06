package memory

import (
	"github.com/samaritan-proxy/sash/model"
)

var _ model.ServiceRegistry = new(Registry)

type Registry struct {
	services map[string]*model.Service
}

func NewRegistry(services []*model.Service) *Registry {
	m := make(map[string]*model.Service, len(services))
	for _, service := range services {
		m[service.Name] = service
	}
	return &Registry{
		services: m,
	}
}

func (r *Registry) Run(stop <-chan struct{}) {}

func (r *Registry) List() ([]string, error) {
	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	return names, nil
}

func (r *Registry) Get(name string) (*model.Service, error) {
	return r.services[name], nil
}

func (r *Registry) DeleteService(name string) bool {
	_, ok := r.services[name]
	if !ok {
		return false
	}
	delete(r.services, name)
	return true
}

func (r *Registry) AddService(service *model.Service) {
	r.services[service.Name] = service
}

func (r *Registry) AddInstance(name string, instances ...*model.ServiceInstance) {
	r.addOrUpdateInstance(name, instances...)
}

func (r *Registry) UpdateInstance(name string, instances ...*model.ServiceInstance) {
	r.addOrUpdateInstance(name, instances...)
}

func (r *Registry) addOrUpdateInstance(name string, instances ...*model.ServiceInstance) {
	service, ok := r.services[name]
	if !ok {
		return
	}

	for _, instance := range instances {
		addr := instance.Addr
		service.Instances[addr] = instance
	}
}

func (r *Registry) DeleteInstance(name string, instances ...*model.ServiceInstance) {
	service, ok := r.services[name]
	if !ok {
		return
	}

	for _, instance := range instances {
		delete(service.Instances, instance.Addr)
	}
}

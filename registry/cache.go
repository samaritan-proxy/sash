package registry

import (
	"context"
	"time"

	"github.com/samaritan-proxy/sash/model"
)

type cacheOptions struct {
	syncFreq   time.Duration
	syncJitter float64
}

func newDefaultCacheOptions() *cacheOptions {
	return &cacheOptions{}
}

type CacheOption func(o *cacheOptions)

type Cache struct {
	model.ServiceRegistry

	options     *cacheOptions
	svcEvtHdls  []ServiceEventHandler
	instEvtHdls []InstanceEventHandler

	services map[string]*model.Service
}

func NewCache(registry model.ServiceRegistry, opts ...CacheOption) *Cache {
	// init options
	o := newDefaultCacheOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &Cache{
		ServiceRegistry: registry,
		options:         o,
	}
}

func (c *Cache) RegisterServiceEventHandler(handler ServiceEventHandler) {
	c.svcEvtHdls = append(c.svcEvtHdls, handler)
}

func (c *Cache) RegisterInstanceEventHandler(handler InstanceEventHandler) {
	c.instEvtHdls = append(c.instEvtHdls, handler)
}

func (c *Cache) Run(ctx context.Context) {
}

func (c *Cache) loadFromRegistry(ctx context.Context) (map[string]*model.Service, error) {
	names, err := c.List()
	if err != nil {
		return nil, err
	}

	services := make(map[string]*model.Service, len(names))
	for _, name := range names {
		service, err := c.Get(name)
		if err == nil {
			services[name] = service
			continue
		}

		// retry
	}
	return services, nil
}

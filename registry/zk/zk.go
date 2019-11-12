package zk

import (
	"context"
	"encoding/json"
	"path"

	"github.com/samaritan-proxy/sash/internal/zk"
	"github.com/samaritan-proxy/sash/model"
)

// InstanceUnmarshaler unmarshals raw byte array which is stored
// in the underlying registry into a service instance. JSONInstanceUnmarshaler
// is the default implementation, could customize according to demands.
type InstanceUnmarshaler interface {
	Unmarshal(data []byte) (*model.ServiceInstance, error)
}

// JSONInstanceUnmarshaler implements the InstanceUnmarshaler interface.
type JSONInstanceUnmarshaler struct{}

// Unmarshal unmarshals raw data into a service instance.
func (u *JSONInstanceUnmarshaler) Unmarshal(data []byte) (*model.ServiceInstance, error) {
	inst := &model.ServiceInstance{}
	err := json.Unmarshal(data, inst)
	return inst, err
}

type discoveryClientOption func(c *DiscoveryClient)

// WithInstanceUnmarshaler returns a option specifying a customized InstanceUnmarshaler.
func WithInstanceUnmarshaler(u InstanceUnmarshaler) discoveryClientOption {
	return func(c *DiscoveryClient) {
		c.instUnmarshaler = u
	}
}

// DiscoveryClient is a client for service discovery based on zookeeper.
type DiscoveryClient struct {
	connCfg         *zk.ConnConfig
	conn            *zk.Conn
	basePath        string
	instUnmarshaler InstanceUnmarshaler
}

// NewDiscoveryClient creates a discovery client with given parameters.
func NewDiscoveryClient(connCfg *zk.ConnConfig, basePath string, options ...discoveryClientOption) (*DiscoveryClient, error) {
	conn, err := zk.CreateConn(connCfg, false)
	if err != nil {
		return nil, err
	}
	c := NewDiscoveryClientWithConn(conn, basePath, options...)
	c.connCfg = connCfg
	return c, nil
}

// NewDiscoveryClientWithConn creates a discovery client with given zk conn and other parameters.
func NewDiscoveryClientWithConn(conn *zk.Conn, basePath string, options ...discoveryClientOption) *DiscoveryClient {
	c := &DiscoveryClient{
		conn:            conn,
		basePath:        basePath,
		instUnmarshaler: new(JSONInstanceUnmarshaler),
	}
	for _, option := range options {
		option(c)
	}
	return c
}

// Run starts the client unitl the context is canceld or deadline exceed.
func (c *DiscoveryClient) Run(ctx context.Context) {
	<-ctx.Done()
	if c.connCfg != nil {
		c.conn.Close()
	}
}

// List lists all registered service names.
func (c *DiscoveryClient) List() ([]string, error) {
	names, _, err := c.conn.Children(c.basePath)
	return names, err
}

// Get retrieves a service by name.
func (c *DiscoveryClient) Get(name string) (*model.Service, error) {
	servicePath := path.Join(c.basePath, name)
	instNames, _, err := c.conn.Children(servicePath)
	if err != nil {
		return nil, err
	}

	insts := make([]*model.ServiceInstance, 0, len(instNames))
	for _, instName := range instNames {
		instPath := path.Join(servicePath, instName)
		data, _, err := c.conn.Get(instPath)
		if err != nil {
			return nil, err
		}

		inst, err := c.instUnmarshaler.Unmarshal(data)
		if err != nil {
			return nil, err
		}
		insts = append(insts, inst)
	}
	svc := model.NewService(name, insts...)
	return svc, nil
}

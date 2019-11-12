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

package zk

import (
	"context"
	"encoding/json"
	"errors"
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
	conn            zk.Conn
	basePath        string
	instUnmarshaler InstanceUnmarshaler
}

// NewDiscoveryClient creates a discovery client with given parameters.
func NewDiscoveryClient(connCfg *zk.ConnConfig, basePath string, options ...discoveryClientOption) (*DiscoveryClient, error) {
	conn, err := zk.CreateConn(connCfg, false)
	if err != nil {
		return nil, err
	}

	c, err := NewDiscoveryClientWithConn(conn, basePath, options...)
	if err != nil {
		conn.Close()
		return nil, err
	}
	c.connCfg = connCfg
	return c, nil
}

// NewDiscoveryClientWithConn creates a discovery client with given zk conn and other parameters.
func NewDiscoveryClientWithConn(conn zk.Conn, basePath string, options ...discoveryClientOption) (*DiscoveryClient, error) {
	if basePath == "" {
		return nil, errors.New("empty base path")
	}
	c := &DiscoveryClient{
		conn:            conn,
		basePath:        basePath,
		instUnmarshaler: new(JSONInstanceUnmarshaler),
	}
	for _, option := range options {
		option(c)
	}
	return c, nil
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
	svcPath := path.Join(c.basePath, name)
	instNames, _, err := c.conn.Children(svcPath)
	if err != nil {
		return nil, err
	}

	insts := make([]*model.ServiceInstance, 0, len(instNames))
	for _, instName := range instNames {
		instPath := path.Join(svcPath, instName)
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

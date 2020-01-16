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
	"errors"
	"path"

	zkpkg "github.com/mesosphere/go-zookeeper/zk"

	"github.com/samaritan-proxy/sash/config"
	"github.com/samaritan-proxy/sash/internal/zk"
)

type ConnConfig = zk.ConnConfig

type Store struct {
	connCfg *zk.ConnConfig
	conn    zk.Conn

	basePath string
}

func New(connCfg *ConnConfig) (*Store, error) {
	conn, err := zk.CreateConn(connCfg, true)
	if err != nil {
		return nil, err
	}
	c, err := NewWithConn(conn, connCfg.BasePath)
	if err != nil {
		conn.Close()
		return nil, err
	}
	c.connCfg = connCfg
	return c, nil
}

func NewWithConn(conn zk.Conn, basePath string) (*Store, error) {
	if basePath == "" {
		return nil, errors.New("empty base path")
	}
	c := &Store{
		conn:     conn,
		basePath: basePath,
	}
	return c, nil
}

func (s *Store) Get(namespace, typ, key string) ([]byte, error) {
	zPath := path.Join(s.basePath, namespace, typ, key)
	b, _, err := s.conn.Get(zPath)
	switch err {
	case nil:
		return b, nil
	case zkpkg.ErrNoNode:
		return nil, config.ErrNotExist
	default:
		return nil, err
	}
}

func (s *Store) Add(namespace, typ, key string, value []byte) error {
	// TODO: return error when key exist
	zPath := path.Join(s.basePath, namespace, typ, key)
	return s.conn.CreateRecursively(zPath, value)
}

func (s *Store) Update(namespace, typ, key string, value []byte) error {
	// TODO: return error when key not exist
	zPath := path.Join(s.basePath, namespace, typ, key)
	return s.conn.CreateRecursively(zPath, value)
}

func (s *Store) Del(namespace, typ, key string) error {
	zPath := path.Join(s.basePath, namespace, typ, key)
	switch err := s.conn.DeleteWithChildren(zPath); err {
	case zkpkg.ErrNoNode:
		return config.ErrNotExist
	default:
		return err
	}
}

func (s *Store) Exist(namespace, typ, key string) bool {
	zPath := path.Join(s.basePath, namespace, typ, key)
	ok, _, _ := s.conn.Exists(zPath)
	return ok
}

func (s *Store) GetKeys(namespace, typ string) ([]string, error) {
	zPath := path.Join(s.basePath, namespace, typ)
	nodes, _, err := s.conn.Children(zPath)
	switch err {
	case nil:
		return nodes, err
	case zkpkg.ErrNoNode:
		return nil, nil
	default:
		return nil, err
	}
}

func (s *Store) Start() error { return nil }

func (s *Store) Stop() {
	// zk conn is created inside
	if s.connCfg != nil {
		s.conn.Close()
	}
}

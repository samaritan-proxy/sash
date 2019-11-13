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
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"path"
	"time"

	"github.com/mesosphere/go-zookeeper/zk"
	"go.uber.org/atomic"

	"github.com/samaritan-proxy/sash/logger"
)

var (
	errConnectTimeout = errors.New("zk connect timeout")
)

func init() {
	// To make sure the seed of rand.globalRand is not overrode by anything.
	// This is done for the process of shuffling servers in package zk.
	// See also: https://github.com/samuel/go-zookeeper/pull/146
	rand.Seed(time.Now().UnixNano())
	// The above line exists in an init function of package zk.
	// But global seed is still 1(default), maybe the seed is set to 1 after that somewhere else?
	// I can't find where the seed is set to 1, but this works.
}

// ConnConfig contains all zk connection configurations.
type ConnConfig struct {
	Hosts          []string
	User           string
	Pwd            string
	ConnectTimeout time.Duration
	SessionTimeout time.Duration
}

// auth returns a string of "User:Pwd".
func (cfg *ConnConfig) auth() string {
	if cfg.User != "" || cfg.Pwd != "" {
		return fmt.Sprintf("%s:%s", cfg.User, cfg.Pwd)
	}
	return ""
}

// Conn is a zk connection wrapper.
type Conn interface {
	Get(path string) ([]byte, *zk.Stat, error)
	Children(path string) ([]string, *zk.Stat, error)
	Exists(path string) (bool, *zk.Stat, error)
	Close()
}

//go:generate mockgen -source ./conn.go -destination mock_conn.go -package zk

// CreateConn creates zk connection and connects it
func CreateConn(cfg *ConnConfig, waitConnected bool) (Conn, error) {
	return createConn(cfg, waitConnected)
}

type conn struct {
	*zk.Conn
	update    <-chan zk.Event
	cfg       *ConnConfig
	authAdded *atomic.Bool
}

func createConn(cfg *ConnConfig, waitConnected bool) (*conn, error) {
	if cfg.SessionTimeout <= 0 {
		cfg.SessionTimeout = time.Second
	}
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = time.Second
	}

	var err error
	c := &conn{
		cfg:       cfg,
		authAdded: atomic.NewBool(false),
	}
	c.Conn, c.update, err = zk.Connect(cfg.Hosts, cfg.SessionTimeout,
		func(c *zk.Conn) {
			c.SetLogger(logger.Get())
		},
		zk.WithDialer(zk.Dialer(func(network, address string, timeout time.Duration) (net.Conn, error) {
			// a dialer only to override the timeout which is set to 1s inside go-zookeeper
			return net.DialTimeout(network, address, cfg.ConnectTimeout)
		})))
	if err != nil {
		return nil, err
	}

	err = c.addAuth()
	if err != nil {
		return nil, err
	}

	if waitConnected {
		err = c.waitConnected()
	}
	return c, err
}

// Authed returns if the zk connection is authorized
func (c *conn) Authed() bool {
	return c.authAdded.Load()
}

var addAuth = func(conn *zk.Conn, scheme string, auth []byte) error {
	return conn.AddAuth(scheme, auth)
}

func (c *conn) addAuth() error {
	if c.cfg.auth() != "" && !c.Authed() {
		logger.Debug("Authenticating with: ", c.cfg.auth())
		err := addAuth(c.Conn, "digest", []byte(c.cfg.auth()))
		if err == nil {
			c.authAdded.Store(true)
		}
		return err
	}
	return nil
}

func (c *conn) waitConnected() error {
	attempt := 0

	for {
		e := <-c.update
		switch e.State {
		case zk.StateConnected, zk.StateHasSession:
			logger.Info("Zookeeper connection established")
			return nil
		case zk.StateConnecting:
			attempt++
		}

		if attempt >= 2 {
			break
		}
	}
	return errConnectTimeout
}

// CreateRecursively creates path with given data and its parents if necessary
func (c *conn) CreateRecursively(p string, data string) error {
	var err error

	if err = c.createParentRecursively(p); err != nil {
		return err
	}

	isEqual, err := c.nodeEqual(p, data)
	if err != nil {
		_, err = c.Create(p, []byte(data), 0, c.getACL())
		return err
	}
	if !isEqual {
		_, err = c.Set(p, []byte(data), -1)
	}
	return err
}

// createParentRecursively creates given path's parents if necessary.
func (c *conn) createParentRecursively(p string) error {
	var parent = path.Dir(p)
	if parent == p {
		return nil
	}

	existed, _, err := c.Exists(parent)
	if err != nil {
		return err
	}
	if existed {
		return nil
	}

	if err = c.createParentRecursively(parent); err != nil {
		return err
	}

	_, err = c.Create(parent, []byte(""), 0, c.getACL())
	return err
}

func (c *conn) nodeEqual(p string, data string) (bool, error) {
	existData, _, err := c.Get(p)
	if err != nil {
		return false, err
	}
	return bytes.Equal([]byte(data), existData), nil
}

func (c *conn) getACL() []zk.ACL {
	var acl = zk.WorldACL(zk.PermAll)
	if c.cfg.auth() != "" {
		acl = zk.DigestACL(zk.PermAll, c.cfg.User, c.cfg.Pwd)
	}
	return acl
}

// DeleteWithChildren delete path and its children if necessary
func (c *conn) DeleteWithChildren(pathcur string) error {
	deleteFn := func(pathcur string, data []byte) error {
		return c.Delete(pathcur, int32(-1))
	}
	err := c.walk(pathcur, deleteFn)
	return err
}

func (c *conn) walk(pathcur string, fn func(pathcur string, data []byte) error) error {
	pathIsExist, _, err := c.Exists(pathcur)
	if err != nil {
		return err
	}
	if !pathIsExist {
		return zk.ErrNoNode
	}
	children, _, err := c.Children(pathcur)
	if err != nil {
		return err
	}
	for _, child := range children {
		childpath := path.Join(pathcur, child)
		if err := c.walk(childpath, fn); err != nil {
			return err
		}
	}
	data, _, err := c.Get(pathcur)
	if err != nil {
		return err
	}
	return fn(pathcur, data)
}

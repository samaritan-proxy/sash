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

package config

import (
	"bytes"
	"fmt"
	"hash/fnv"
)

// RawConf represents a raw configuration.
type RawConf struct {
	Namespace string
	Type      string
	Key       string
	Value     []byte
}

// NewRawConf return a new RawConf.
func NewRawConf(namespace, typ, key string, value []byte) *RawConf {
	return &RawConf{
		Namespace: namespace,
		Type:      typ,
		Key:       key,
		Value:     value,
	}
}

// Hashcode return the hashcode of this RawConf
func (c *RawConf) Hashcode() uint32 {
	if c == nil {
		return 0
	}

	hash := fnv.New32()
	_, err := fmt.Fprintf(hash, "%s|%s|%s", c.Namespace, c.Type, c.Key)
	if err != nil {
		return 0
	}
	return hash.Sum32()
}

// Equal return true if input RawConf is equal with current RawConf.
func (c *RawConf) Equal(that *RawConf) bool {
	if that == nil {
		return c == nil
	}
	if that.Namespace != c.Namespace {
		return false
	}
	if that.Type != c.Type {
		return false
	}
	if that.Key != c.Key {
		return false
	}
	return bytes.Equal(that.Value, c.Value)
}

type typ struct {
	name    string
	configs map[string][]byte
}

func newType(name string) *typ {
	return &typ{
		name:    name,
		configs: make(map[string][]byte),
	}
}

func (t *typ) Exist(key string) bool {
	_, ok := t.configs[key]
	return ok
}

func (t *typ) Get(key string) ([]byte, error) {
	if !t.Exist(key) {
		return nil, ErrKeyNotExist
	}
	return t.configs[key], nil
}

func (t *typ) Set(key string, value []byte) {
	t.configs[key] = value
}

func (t *typ) Del(key string) error {
	if !t.Exist(key) {
		return ErrKeyNotExist
	}
	delete(t.configs, key)
	return nil
}

func (t *typ) Keys() []string {
	keys := make([]string, 0, len(t.configs))
	for k := range t.configs {
		keys = append(keys, k)
	}
	return keys
}

func (t *typ) IsEmpty() bool {
	return len(t.configs) == 0
}

type namespace struct {
	name  string
	types map[string]*typ
}

func NewNamespace(name string) *namespace {
	return &namespace{
		name:  name,
		types: make(map[string]*typ),
	}
}

func (n *namespace) Exist(typ string) bool {
	_, ok := n.types[typ]
	return ok
}

func (n *namespace) Get(typ, key string) ([]byte, error) {
	if !n.Exist(typ) {
		return nil, ErrTypeNotExist
	}
	return n.types[typ].Get(key)
}

func (n *namespace) Set(typ, key string, value []byte) {
	if !n.Exist(typ) {
		n.types[typ] = newType(typ)
	}
	n.types[typ].Set(key, value)
}

func (n *namespace) Del(typ, key string) error {
	if !n.Exist(typ) {
		return ErrTypeNotExist
	}
	typs := n.types[typ]
	if err := typs.Del(key); err != nil {
		return err
	}
	if typs.IsEmpty() {
		delete(n.types, typ)
	}
	return nil
}

func (n *namespace) Keys(typ string) ([]string, error) {
	if !n.Exist(typ) {
		return nil, ErrTypeNotExist
	}
	return n.types[typ].Keys(), nil
}

func (n *namespace) IsEmpty() bool {
	return len(n.types) == 0
}

type Cache struct {
	namespaces map[string]*namespace
	all        map[uint32]*RawConf
}

func NewCache() *Cache {
	return &Cache{
		namespaces: make(map[string]*namespace),
		all:        make(map[uint32]*RawConf),
	}
}

func (c *Cache) Exist(ns string) bool {
	_, ok := c.namespaces[ns]
	return ok
}

func (c *Cache) Get(ns, typ, key string) ([]byte, error) {
	if !c.Exist(ns) {
		return nil, ErrNamespaceNotExist
	}
	return c.namespaces[ns].Get(typ, key)
}

func (c *Cache) Set(ns, typ, key string, value []byte) {
	if !c.Exist(ns) {
		c.namespaces[ns] = NewNamespace(ns)
	}
	c.namespaces[ns].Set(typ, key, value)
	cfg := NewRawConf(ns, typ, key, value)
	c.all[cfg.Hashcode()] = cfg
}

func (c *Cache) Del(ns, typ, key string) error {
	if !c.Exist(ns) {
		return ErrNamespaceNotExist
	}
	namespace := c.namespaces[ns]
	if err := namespace.Del(typ, key); err != nil {
		return err
	}
	if namespace.IsEmpty() {
		delete(c.namespaces, ns)
	}
	hashcode := NewRawConf(ns, typ, key, nil).Hashcode()
	delete(c.all, hashcode)
	return nil
}

func (c *Cache) Diff(that *Cache) (add, update, del []*RawConf) {
	allKeys := make(map[uint32]struct{})
	for k := range c.all {
		allKeys[k] = struct{}{}
	}
	for k := range that.all {
		allKeys[k] = struct{}{}
	}

	for k := range allKeys {
		var (
			newConf, newConfExist = that.all[k]
			oldConf, oldConfExist = c.all[k]
		)

		switch {
		case newConfExist && oldConfExist:
			if oldConf.Equal(newConf) {
				continue
			}
			update = append(update, newConf)
		case newConfExist && !oldConfExist: //Add
			add = append(add, newConf)
		case !newConfExist && oldConfExist: // Remove
			del = append(del, oldConf)
		}
	}
	return
}

func (c *Cache) Keys(ns, typ string) ([]string, error) {
	if !c.Exist(ns) {
		return nil, ErrNamespaceNotExist
	}
	return c.namespaces[ns].Keys(typ)
}

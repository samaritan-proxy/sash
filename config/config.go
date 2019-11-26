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

func (c *RawConf) Copy() *RawConf {
	if c == nil {
		return nil
	}
	value := make([]byte, len(c.Value))
	copy(value, c.Value)
	return &RawConf{
		Namespace: c.Namespace,
		Type:      c.Type,
		Key:       c.Key,
		Value:     value,
	}
}

type typeCache struct {
	name    string
	configs map[string][]byte
}

func newTypeCache(name string) *typeCache {
	return &typeCache{
		name:    name,
		configs: make(map[string][]byte),
	}
}

func (t *typeCache) Exist(key string) bool {
	_, ok := t.configs[key]
	return ok
}

func (t *typeCache) Get(key string) ([]byte, error) {
	if !t.Exist(key) {
		return nil, ErrNotExist
	}
	return t.configs[key], nil
}

func (t *typeCache) Set(key string, value []byte) {
	t.configs[key] = value
}

func (t *typeCache) Del(key string) error {
	if !t.Exist(key) {
		return ErrNotExist
	}
	delete(t.configs, key)
	return nil
}

func (t *typeCache) Keys() []string {
	keys := make([]string, 0, len(t.configs))
	for k := range t.configs {
		keys = append(keys, k)
	}
	return keys
}

func (t *typeCache) IsEmpty() bool {
	return len(t.configs) == 0
}

func (t *typeCache) Copy() *typeCache {
	if t == nil {
		return nil
	}
	configs := make(map[string][]byte, len(t.configs))
	for k, v := range t.configs {
		value := make([]byte, len(v))
		copy(value, v)
		configs[k] = value
	}
	return &typeCache{
		name:    t.name,
		configs: configs,
	}
}

type nsCache struct {
	name  string
	types map[string]*typeCache
}

func newNsCache(name string) *nsCache {
	return &nsCache{
		name:  name,
		types: make(map[string]*typeCache),
	}
}

func (n *nsCache) Exist(typ string) bool {
	_, ok := n.types[typ]
	return ok
}

func (n *nsCache) Get(typ, key string) ([]byte, error) {
	if !n.Exist(typ) {
		return nil, ErrNotExist
	}
	return n.types[typ].Get(key)
}

func (n *nsCache) Set(typ, key string, value []byte) {
	if !n.Exist(typ) {
		n.types[typ] = newTypeCache(typ)
	}
	n.types[typ].Set(key, value)
}

func (n *nsCache) Del(typ, key string) error {
	if !n.Exist(typ) {
		return ErrNotExist
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

func (n *nsCache) Keys(typ string) ([]string, error) {
	if !n.Exist(typ) {
		return nil, ErrNotExist
	}
	return n.types[typ].Keys(), nil
}

func (n *nsCache) IsEmpty() bool {
	return len(n.types) == 0
}

func (n *nsCache) Copy() *nsCache {
	if n == nil {
		return nil
	}
	types := make(map[string]*typeCache, len(n.types))
	for k, v := range n.types {
		types[k] = v.Copy()
	}
	return &nsCache{
		name:  n.name,
		types: types,
	}
}

type Cache struct {
	namespaces map[string]*nsCache
	all        map[uint32]*RawConf
}

func NewCache() *Cache {
	return &Cache{
		namespaces: make(map[string]*nsCache),
		all:        make(map[uint32]*RawConf),
	}
}

func (c *Cache) Exist(ns string) bool {
	_, ok := c.namespaces[ns]
	return ok
}

func (c *Cache) Get(ns, typ, key string) ([]byte, error) {
	if !c.Exist(ns) {
		return nil, ErrNotExist
	}
	return c.namespaces[ns].Get(typ, key)
}

func (c *Cache) Set(ns, typ, key string, value []byte) {
	if !c.Exist(ns) {
		c.namespaces[ns] = newNsCache(ns)
	}
	c.namespaces[ns].Set(typ, key, value)
	cfg := NewRawConf(ns, typ, key, value)
	c.all[cfg.Hashcode()] = cfg
}

func (c *Cache) Del(ns, typ, key string) error {
	if !c.Exist(ns) {
		return ErrNotExist
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
		return nil, ErrNotExist
	}
	return c.namespaces[ns].Keys(typ)
}

func (c *Cache) Copy() *Cache {
	if c == nil {
		return nil
	}
	namespaces := make(map[string]*nsCache, len(c.namespaces))
	for k, v := range c.namespaces {
		namespaces[k] = v.Copy()
	}
	all := make(map[uint32]*RawConf)
	for k, v := range c.all {
		all[k] = v.Copy()
	}
	return &Cache{
		namespaces: namespaces,
		all:        all,
	}
}

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

package memory

import (
	"bytes"
	"sync"

	"github.com/samaritan-proxy/sash/config"
)

// MemStore is a in memory implement of Store.
type MemStore struct {
	sync.RWMutex
	evtCh       chan struct{}
	configs     map[string]map[string]map[string][]byte // namespace, type, key, value
	subscribeNS map[string]struct{}
}

func NewMemStore() *MemStore {
	return &MemStore{
		evtCh:       make(chan struct{}, 64),
		configs:     make(map[string]map[string]map[string][]byte),
		subscribeNS: make(map[string]struct{}),
	}
}

func (s *MemStore) Get(namespace, typ, key string) ([]byte, error) {
	s.RLock()
	defer s.RUnlock()

	types, ok := s.configs[namespace]
	if !ok || types == nil {
		return nil, config.ErrNamespaceNotExist
	}

	keys, ok := types[typ]
	if !ok || keys == nil {
		return nil, config.ErrTypeNotExist
	}

	b, ok := keys[key]
	if !ok {
		return nil, config.ErrKeyNotExist
	}

	return b, nil
}

func (s *MemStore) Set(namespace, typ, key string, value []byte) error {
	s.Lock()
	defer s.Unlock()

	types, ok := s.configs[namespace]
	if !ok || types == nil {
		types = make(map[string]map[string][]byte)
		s.configs[namespace] = types
	}

	keys, ok := types[typ]
	if !ok || keys == nil {
		keys = make(map[string][]byte)
		s.configs[namespace][typ] = keys
	}

	update := false
	oldValue, ok := keys[key]
	if ok {
		update = !bytes.Equal(oldValue, value)
	}
	keys[key] = value

	if _, ok := s.subscribeNS[namespace]; ok && update {
		s.evtCh <- struct{}{}
	}
	return nil
}

func (s *MemStore) Del(namespace, typ, key string) error {
	s.Lock()
	defer s.Unlock()

	types, ok := s.configs[namespace]
	if !ok || types == nil {
		return config.ErrNamespaceNotExist
	}

	keys, ok := types[typ]
	if !ok || keys == nil {
		return config.ErrTypeNotExist
	}

	_, ok = keys[key]
	if !ok {
		return config.ErrKeyNotExist
	}

	delete(keys, key)
	if len(keys) == 0 {
		delete(types, typ)
	}
	if len(types) == 0 {
		delete(s.configs, namespace)
	}

	return nil
}

func (s *MemStore) Exist(namespace, typ, key string) bool {
	s.RLock()
	defer s.RUnlock()

	types, ok := s.configs[namespace]
	if !ok || types == nil {
		return false
	}

	keys, ok := types[typ]
	if !ok || keys == nil {
		return false
	}

	_, ok = keys[key]

	return ok
}

func (s *MemStore) GetKeys(namespace, typ string) ([]string, error) {
	s.RLock()
	defer s.RUnlock()

	if _, ok := s.configs[namespace]; !ok {
		return nil, config.ErrNamespaceNotExist
	}

	if _, ok := s.configs[namespace][typ]; !ok {
		return nil, config.ErrTypeNotExist
	}

	keys := make([]string, 0, len(s.configs[namespace][typ]))
	for key := range s.configs[namespace][typ] {
		keys = append(keys, key)
	}

	return keys, nil
}

func (s *MemStore) Subscribe(namespace string) error {
	s.Lock()
	defer s.Unlock()
	s.subscribeNS[namespace] = struct{}{}
	return nil
}

func (s *MemStore) UnSubscribe(namespace string) error {
	s.Lock()
	defer s.Unlock()
	delete(s.subscribeNS, namespace)
	return nil
}

func (s *MemStore) Event() <-chan struct{} {
	return s.evtCh
}

func (s *MemStore) Start() error { return nil }

func (s *MemStore) Stop() {}

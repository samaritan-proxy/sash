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
	configs     *config.Cache
	subscribeNS map[string]struct{}
}

// NewMemStore return a new MemStore.
func NewMemStore() *MemStore {
	return &MemStore{
		evtCh:       make(chan struct{}, 64),
		configs:     config.NewCache(),
		subscribeNS: make(map[string]struct{}),
	}
}

func (s *MemStore) Get(namespace, typ, key string) ([]byte, error) {
	s.RLock()
	defer s.RUnlock()

	return s.configs.Get(namespace, typ, key)
}

func (s *MemStore) Add(namespace, typ, key string, value []byte) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.configs.Get(namespace, typ, key)
	switch err {
	case config.ErrNotExist:
	case nil:
		return config.ErrExist
	default:
		return err
	}

	s.configs.Set(namespace, typ, key, value)
	if _, ok := s.subscribeNS[namespace]; ok {
		s.evtCh <- struct{}{}
	}
	return nil
}

func (s *MemStore) Update(namespace, typ, key string, value []byte) error {
	s.Lock()
	defer s.Unlock()

	update := false
	oldValue, err := s.configs.Get(namespace, typ, key)
	if err != nil {
		return err
	}
	if !bytes.Equal(oldValue, value) {
		update = true
	}

	s.configs.Set(namespace, typ, key, value)

	if _, ok := s.subscribeNS[namespace]; ok && update {
		s.evtCh <- struct{}{}
	}
	return nil
}

func (s *MemStore) Del(namespace, typ, key string) error {
	s.Lock()
	defer s.Unlock()

	if err := s.configs.Del(namespace, typ, key); err != nil {
		return err
	}
	if _, ok := s.subscribeNS[namespace]; ok {
		s.evtCh <- struct{}{}
	}
	return nil
}

func (s *MemStore) Exist(namespace, typ, key string) bool {
	s.RLock()
	defer s.RUnlock()

	_, err := s.configs.Get(namespace, typ, key)
	return err == nil
}

func (s *MemStore) GetKeys(namespace, typ string) ([]string, error) {
	s.RLock()
	defer s.RUnlock()

	return s.configs.Keys(namespace, typ)
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

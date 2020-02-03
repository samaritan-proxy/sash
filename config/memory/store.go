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
	"strings"
	"sync"
	"time"

	"github.com/samaritan-proxy/sash/config"
)

// Store is a in memory implement of Store.
type Store struct {
	sync.RWMutex
	evtCh       chan struct{}
	configs     *config.Cache
	metaData    map[string]*config.Metadata
	subscribeNS map[string]struct{}
}

// NewStore return a new Store.
func NewStore() *Store {
	return &Store{
		evtCh:       make(chan struct{}, 64),
		configs:     config.NewCache(),
		metaData:    make(map[string]*config.Metadata),
		subscribeNS: make(map[string]struct{}),
	}
}

func (*Store) mergeKey(s ...string) string {
	return strings.Join(s, "|")
}

func (s *Store) Get(namespace, typ, key string) ([]byte, *config.Metadata, error) {
	s.RLock()
	defer s.RUnlock()

	b, err := s.configs.Get(namespace, typ, key)
	if err != nil {
		return nil, nil, err
	}
	return b, s.metaData[s.mergeKey(namespace, typ, key)], nil
}

func (s *Store) Add(namespace, typ, key string, value []byte) error {
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
	s.metaData[s.mergeKey(namespace, typ, key)] = &config.Metadata{
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	if _, ok := s.subscribeNS[namespace]; ok {
		s.evtCh <- struct{}{}
	}
	return nil
}

func (s *Store) Update(namespace, typ, key string, value []byte) error {
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
	s.metaData[s.mergeKey(namespace, typ, key)].UpdateTime = time.Now()

	if _, ok := s.subscribeNS[namespace]; ok && update {
		s.evtCh <- struct{}{}
	}
	return nil
}

func (s *Store) Del(namespace, typ, key string) error {
	s.Lock()
	defer s.Unlock()

	if err := s.configs.Del(namespace, typ, key); err != nil {
		return err
	}
	delete(s.metaData, s.mergeKey(namespace, typ, key))
	if _, ok := s.subscribeNS[namespace]; ok {
		s.evtCh <- struct{}{}
	}
	return nil
}

func (s *Store) Exist(namespace, typ, key string) bool {
	s.RLock()
	defer s.RUnlock()

	_, err := s.configs.Get(namespace, typ, key)
	return err == nil
}

func (s *Store) GetKeys(namespace, typ string) ([]string, error) {
	s.RLock()
	defer s.RUnlock()

	return s.configs.Keys(namespace, typ)
}

func (s *Store) Subscribe(namespace string) error {
	s.Lock()
	defer s.Unlock()
	s.subscribeNS[namespace] = struct{}{}
	return nil
}

func (s *Store) UnSubscribe(namespace string) error {
	s.Lock()
	defer s.Unlock()
	delete(s.subscribeNS, namespace)
	return nil
}

func (s *Store) Event() <-chan struct{} {
	return s.evtCh
}

func (s *Store) Start() error { return nil }

func (s *Store) Stop() {}

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

// TODO: use $REPO_URI
//go:generate mockgen -package $GOPACKAGE -destination store_mock_test.go github.com/samaritan-proxy/sash/$GOPACKAGE Store,SubscribableStore

import (
	"errors"
)

var (
	ErrNamespaceNotExist = errors.New("namespace not exist")
	ErrTypeNotExist      = errors.New("type not exist")
	ErrKeyNotExist       = errors.New("key not exist")
)

// The store is a kv store.
type Store interface {
	Get(namespace, typ, key string) ([]byte, error)
	Set(namespace, typ, key string, value []byte) error
	Del(namespace, typ, key string) error
	Exist(namespace, typ, key string) bool

	GetKeys(namespace, typ string) ([]string, error)

	Start() error
	Stop()
}

// SubscribableStore allows you to subscribe to the namespace you are interested in,
// and will send a notification when there is some change in the namespace.
type SubscribableStore interface {
	Store
	Subscribe(namespace string) error
	UnSubscribe(namespace string) error
	Event() <-chan struct{}
}

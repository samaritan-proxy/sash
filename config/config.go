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

type rawConf struct {
	Namespace string
	Type      string
	Key       string
	Value     []byte
}

func newRawConf(namespace, typ, key string, value []byte) *rawConf {
	return &rawConf{
		Namespace: namespace,
		Type:      typ,
		Key:       key,
		Value:     value,
	}
}

func (n *rawConf) Hashcode() uint32 {
	if n == nil {
		return 0
	}

	hash := fnv.New32()
	_, err := fmt.Fprintf(hash, "%s|%s|%s", n.Namespace, n.Type, n.Key)
	if err != nil {
		return 0
	}
	return hash.Sum32()
}

func (n *rawConf) Equal(that *rawConf) bool {
	if that == nil {
		return n == nil
	}
	if that.Namespace != n.Namespace {
		return false
	}
	if that.Type != n.Type {
		return false
	}
	if that.Key != n.Key {
		return false
	}
	return bytes.Equal(that.Value, n.Value)
}

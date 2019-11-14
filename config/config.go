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

type RawConf struct {
	Namespace string
	Type      string
	Key       string
	Value     []byte
}

func NewRawConf(namespace, typ, key string, value []byte) *RawConf {
	return &RawConf{
		Namespace: namespace,
		Type:      typ,
		Key:       key,
		Value:     value,
	}
}

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

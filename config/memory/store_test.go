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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/samaritan-proxy/sash/config"
)

func TestGetAndAdd(t *testing.T) {
	s := NewMemStore()
	assert.NoError(t, s.Start())
	defer s.Stop()

	assert.NoError(t, s.Add("a", "b", "c", []byte("hello")))

	b, err := s.Get("a", "b", "c")
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello"), b)

	assert.True(t, s.Exist("a", "b", "c"))
	assert.False(t, s.Exist("a", "d", "c"))

	t.Run("bad key", func(t *testing.T) {
		_, err := s.Get("a", "b", "foo")
		assert.Equal(t, config.ErrNotExist, err)
	})

	t.Run("bad type", func(t *testing.T) {
		_, err = s.Get("a", "foo", "c")
		assert.Equal(t, config.ErrNotExist, err)
	})

	t.Run("bad namespace", func(t *testing.T) {
		_, err = s.Get("foo", "b", "c")
		assert.Equal(t, config.ErrNotExist, err)
	})
}

func TestDel(t *testing.T) {
	s := NewMemStore()
	assert.NoError(t, s.Start())
	defer s.Stop()

	assert.Error(t, s.Del("a", "b", "c"))
	assert.NoError(t, s.Add("a", "b", "c", []byte("hello")))
	assert.NoError(t, s.Del("a", "b", "c"))
	assert.False(t, s.Exist("a", "b", "c"))
}

func TestGetKeys(t *testing.T) {
	s := NewMemStore()
	assert.NoError(t, s.Start())
	defer s.Stop()

	keys := []string{"a", "b", "c"}
	sort.Strings(keys)
	for _, key := range keys {
		assert.NoError(t, s.Add("ns", "type", key, nil))
	}
	_, err := s.GetKeys("foo", "foo")
	assert.Equal(t, config.ErrNotExist, err)
	_, err = s.GetKeys("ns", "foo")
	assert.Equal(t, config.ErrNotExist, err)
	ks, err := s.GetKeys("ns", "type")
	assert.NoError(t, err)
	sort.Strings(ks)
	assert.Equal(t, keys, ks)
}

func TestSubscribeAndUnSubscribe(t *testing.T) {
	s := NewMemStore()
	assert.NoError(t, s.Start())
	defer s.Stop()

	assert.NoError(t, s.Add("ns1", "b", "c", []byte("hello")))
	assert.NoError(t, s.Add("ns2", "b", "c", []byte("hello")))
	assert.NoError(t, s.Subscribe("ns1"))
	assert.NoError(t, s.Update("ns1", "b", "c", []byte("hi")))
	assert.NoError(t, s.Update("ns2", "b", "c", []byte("hi")))
	assert.Len(t, s.Event(), 1)
	<-s.Event()
	assert.Len(t, s.Event(), 0)

	assert.NoError(t, s.UnSubscribe("ns1"))
	assert.NoError(t, s.Update("ns1", "b", "c", []byte("hello")))
	assert.NoError(t, s.Update("ns2", "b", "c", []byte("hello")))
	assert.Len(t, s.Event(), 0)
}

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
	"fmt"
	"hash/fnv"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func fnv32(str string) uint32 {
	hash := fnv.New32()
	_, err := hash.Write([]byte(str))
	if err != nil {
		log.Fatal(err)
	}
	return hash.Sum32()
}

func TestRawConf_Hashcode(t *testing.T) {
	cases := []struct {
		Conf     *RawConf
		Hashcode uint32
	}{
		{nil, 0},
		{NewRawConf("a", "b", "c", nil), fnv32("a|b|c")},
	}
	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			assert.Equal(t, c.Hashcode, c.Conf.Hashcode())
		})
	}
}

func TestRawConf_Equal(t *testing.T) {
	cases := []struct {
		A, B   *RawConf
		Expect bool
	}{
		{
			A:      nil,
			B:      nil,
			Expect: true,
		},
		{
			A:      NewRawConf("a", "b", "c", []byte("hello")),
			B:      NewRawConf("a", "b", "c", []byte("hello")),
			Expect: true,
		},
		{
			A:      NewRawConf("a", "b", "c", []byte("hello")),
			B:      NewRawConf("foo", "b", "c", []byte("hello")),
			Expect: false,
		},
		{
			A:      NewRawConf("a", "b", "c", []byte("hello")),
			B:      NewRawConf("a", "foo", "c", []byte("hello")),
			Expect: false,
		},
		{
			A:      NewRawConf("a", "b", "c", []byte("hello")),
			B:      NewRawConf("a", "b", "foo", []byte("hello")),
			Expect: false,
		},
		{
			A:      NewRawConf("a", "b", "c", []byte("hello")),
			B:      NewRawConf("a", "b", "c", []byte("hi")),
			Expect: false,
		},
	}
	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			assert.Equal(t, c.Expect, c.A.Equal(c.B))
		})
	}
}

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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalSvcDependence(t *testing.T) {
	cases := []struct {
		Input  []byte
		Output []string
	}{
		{
			Input:  []byte(""),
			Output: []string{},
		},
		{
			Input:  []byte("1"),
			Output: []string{"1"},
		},
		{
			Input:  []byte("1,2"),
			Output: []string{"1", "2"},
		},
		{
			Input:  []byte("abc,def,gij"),
			Output: []string{"abc", "def", "gij"},
		},
		{
			Input:  []byte("abc,def,gij,"),
			Output: []string{"abc", "def", "gij"},
		},
	}
	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			deps, err := unmarshalSvcDependence(c.Input)
			assert.NoError(t, err)
			assert.Equal(t, c.Output, deps)
		})
	}
}

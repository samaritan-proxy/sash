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

package discovery

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffSlice(t *testing.T) {
	cases := []struct {
		A, B     []string
		Add, Del []string
	}{
		{
			A:   []string{},
			B:   []string{},
			Add: []string{},
			Del: []string{},
		},
		{
			A:   nil,
			B:   nil,
			Add: []string{},
			Del: []string{},
		},
		{
			A:   []string{"A", "B"},
			B:   []string{"B", "C"},
			Add: []string{"C"},
			Del: []string{"A"},
		},
	}
	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			add, del := diffSlice(c.A, c.B)
			assert.ElementsMatch(t, c.Add, add)
			assert.ElementsMatch(t, c.Del, del)
		})
	}
}

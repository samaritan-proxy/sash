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

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteMsg(t *testing.T) {
	resp := httptest.NewRecorder()
	writeMsg(resp, http.StatusAccepted, "foo")
	assert.Equal(t, http.StatusAccepted, resp.Code)
	assert.Equal(t, "foo", resp.Body.String())
}

func TestWriteJSON(t *testing.T) {
	cases := []struct {
		Obj      interface{}
		HasError bool
	}{
		{
			Obj: nil,
		},
		{
			Obj: map[string]string{
				"foo": "bar",
			},
		},
		{
			// json.Marshal can not marshal a map which the type of key is interface{}
			Obj: map[interface{}]string{
				1: "foo",
			},
			HasError: true,
		},
	}
	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			resp := httptest.NewRecorder()
			writeJSON(resp, c.Obj)
			if c.HasError {
				assert.Equal(t, http.StatusInternalServerError, resp.Code)
				return
			}
			b, err := json.Marshal(c.Obj)
			assert.NoError(t, err)

			assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
			assert.Equal(t, b, resp.Body.Bytes())
		})
	}
}

type _strings []string

func (s _strings) Len() int { return len(s) }

func (s _strings) Swap(a, b int) { s[a], s[b] = s[b], s[a] }

func (s _strings) Less(a, b int) bool { return s[a] < s[b] }

type badItem struct{}

func (*badItem) Len() int { return 0 }

func (*badItem) Swap(a, b int) {}

func (*badItem) Less(a, b int) bool { return false }

func TestWritePagedResp(t *testing.T) {
	cases := []struct {
		PageNum  int
		PageSize int
		Item     interface{}
		HasError bool
		Response string
	}{
		{
			PageNum:  0,
			PageSize: 0,
			Item:     nil,
			HasError: false,
			Response: `{"page_num": 0, "page_size": 0, "total": 0, "data": null}`,
		},
		{
			PageNum:  0,
			PageSize: 3,
			Item:     _strings{"b", "a", "c", "d"},
			HasError: false,
			Response: `{"page_num": 0, "page_size": 3, "total": 4, "data": ["a", "b", "c"]}`,
		},
		{
			PageNum:  0,
			PageSize: 3,
			Item:     []string{"b", "a", "c", "d"},
			HasError: false,
			Response: `{"page_num": 0, "page_size": 3, "total": 4, "data": ["b", "a", "c"]}`,
		},
		{
			PageNum:  1,
			PageSize: 2,
			Item:     _strings{"b", "a", "c", "d"},
			HasError: false,
			Response: `{"page_num": 1, "page_size": 2, "total": 4, "data": ["c", "d"]}`,
		},
		{
			PageNum:  2,
			PageSize: 2,
			Item:     _strings{"b", "a", "c", "d"},
			HasError: false,
			Response: `{"page_num": 2, "page_size": 2, "total": 4, "data": null}`,
		},
		{
			PageNum:  2,
			PageSize: 2,
			Item:     _strings{"b", "a", "c", "d", "e"},
			HasError: false,
			Response: `{"page_num": 2, "page_size": 2, "total": 5, "data": ["e"]}`,
		},
		{
			PageNum:  -1,
			PageSize: -2,
			Item:     _strings{"b", "a", "c", "d", "e"},
			HasError: false,
			Response: `{"page_num": 0, "page_size": 10, "total": 5, "data": ["a", "b", "c", "d", "e"]}`,
		},
		{
			Item:     &badItem{},
			HasError: true,
		},
	}
	for idx, c := range cases {
		t.Run(fmt.Sprintf("case %d", idx+1), func(t *testing.T) {
			resp := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("http://1.1.1.1/mock?page_num=%d&page_size=%d", c.PageNum, c.PageSize),
				nil,
			)
			writePagedResp(resp, req, c.Item)
			if c.HasError {
				assert.NotEqual(t, http.StatusOK, resp.Code)
				return
			}
			assert.JSONEq(t, c.Response, resp.Body.String())
		})
	}
}

type TestStruct struct {
	FieldString  string  `json:"field_string"`
	FieldInt32   int32   `json:"field_int_32"`
	FieldUint32  uint32  `json:"field_uint_32"`
	FieldFloat64 float64 `json:"field_float_64"`
	FieldBool    bool    `json:"field_bool"`

	NoTagField       string
	NotBaseTypeField []string `json:"not_base_type_field"`
}

func TestFilterByRequestParams(t *testing.T) {
	cases := []struct {
		Description string
		Items       interface{}
		RequestURI  string
		Expect      []interface{}
		IsError     bool
	}{
		{
			Description: "nil item",
			Items:       nil,
			RequestURI:  "/",
			Expect:      nil,
			IsError:     false,
		},
		{
			Description: "empty slice",
			Items:       []TestStruct{},
			RequestURI:  "/",
			Expect:      []interface{}{},
			IsError:     false,
		},
		{
			Description: "empty array",
			Items:       [0]TestStruct{},
			RequestURI:  "/",
			Expect:      []interface{}{},
			IsError:     false,
		},
		{
			Description: "not a slice or array",
			Items:       "foo",
			RequestURI:  "/",
			Expect:      nil,
			IsError:     true,
		},
		{
			Description: "string",
			Items: []interface{}{
				&TestStruct{
					FieldString: "foo",
				},
				TestStruct{
					FieldString: "foo",
				},
				&TestStruct{
					FieldString: "bar",
				},
				"this item will be skip",
				nil,
			},
			RequestURI: "/?field_string=foo&page=2",
			Expect: []interface{}{
				&TestStruct{
					FieldString: "foo",
				},
				TestStruct{
					FieldString: "foo",
				},
			},
			IsError: false,
		},
		{
			Description: "regexp",
			Items: []interface{}{
				&TestStruct{
					FieldString: "foo",
				},
				&TestStruct{
					FieldString: "123",
				},
			},
			RequestURI: "/?field_string=re%3A%5Ba-z%5D%2B", // re:[a-z]+
			Expect: []interface{}{
				&TestStruct{
					FieldString: "foo",
				},
			},
			IsError: false,
		},
		{
			Description: "bad regexp",
			Items: []interface{}{
				&TestStruct{
					FieldString: "foo",
				},
				&TestStruct{
					FieldString: "123",
				},
			},
			RequestURI: "/?field_string=re%3A%5D%5B", // re:][
			Expect:     nil,
			IsError:    true,
		},
		{
			Description: "int",
			Items: []*TestStruct{
				{
					FieldInt32: 11,
				},
				{
					FieldInt32: 11,
				},
				{
					FieldInt32: 17,
				},
			},
			RequestURI: "/?field_int_32=11",
			Expect: []interface{}{
				&TestStruct{
					FieldInt32: 11,
				},
				&TestStruct{
					FieldInt32: 11,
				},
			},
			IsError: false,
		},
		{
			Description: "array",
			Items: [2]TestStruct{
				{FieldString: "foo"},
				{FieldString: "bar"},
			},
			RequestURI: "/?field_string=foo",
			Expect: []interface{}{
				TestStruct{
					FieldString: "foo",
				},
			},
			IsError: false,
		},
		{
			Description: "multi params",
			Items: []*TestStruct{
				{
					FieldString:  "foo",
					FieldInt32:   11,
					FieldFloat64: 5.0,
				},
				{
					FieldString:  "foo",
					FieldInt32:   12,
					FieldFloat64: 5.0,
				},
				{
					FieldString:  "bar",
					FieldInt32:   11,
					FieldFloat64: 5.0,
				},
			},
			RequestURI: "/?field_string=foo&field_int_32=11",
			Expect: []interface{}{
				&TestStruct{
					FieldString:  "foo",
					FieldInt32:   11,
					FieldFloat64: 5.0,
				},
			},
			IsError: false,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, c.RequestURI, nil)
			result, err := filterItemsByRequestParams(req, c.Items)
			if c.IsError {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, c.Expect, result)
		})
	}
}

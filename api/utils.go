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
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	defaultPageNum  = 0
	defaultPageSize = 10

	contentType     = "Content-Type"
	contentTypeJSON = "application/json"
)

func writeMsg(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(msg))
}

func writeError(w http.ResponseWriter, err error) {

}

func writeJSON(w http.ResponseWriter, obj interface{}) {
	w.Header().Set(contentType, contentTypeJSON)
	b, err := json.Marshal(obj)
	if err != nil {
		writeMsg(w, http.StatusInternalServerError, err.Error())
		return
	}
	_, _ = w.Write(b)
}

func parsePageRequest(r *http.Request) *PageRequest {
	pageNum, err := strconv.Atoi(r.URL.Query().Get(paramPageNum))
	if err != nil || pageNum < 0 {
		pageNum = defaultPageNum
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get(paramPageSize))
	if err != nil || pageSize < 0 {
		pageSize = defaultPageSize
	}
	return &PageRequest{
		PageNum:  pageNum,
		PageSize: pageSize,
	}
}

func writePagedResp(w http.ResponseWriter, r *http.Request, items interface{}) {
	if items == nil {
		writeJSON(w, &PagedResponse{})
		return
	}

	value := reflect.ValueOf(items)
	switch kind := value.Kind(); kind {
	case reflect.Array, reflect.Slice:
	default:
		writeMsg(w, http.StatusInternalServerError, fmt.Sprintf("writePagedResp: unexpected type: %s", kind))
		return
	}

	if sorter, ok := items.(sort.Interface); ok {
		sort.Sort(sorter)
	}

	var (
		pageReq = parsePageRequest(r)
		start   = pageReq.PageNum * pageReq.PageSize
		end     = start + pageReq.PageSize
		all     = value.Len()
		result  interface{}
	)
	if end > all {
		end = all
	}

	if start < all {
		result = value.Slice(start, end).Interface()
	}

	writeJSON(w, &PagedResponse{
		PageNum:  pageReq.PageNum,
		PageSize: pageReq.PageSize,
		Total:    all,
		Data:     result,
	})
}

func isEqual(base, that string) (bool, error) {
	if strings.HasPrefix(base, "re:") {
		goto REGEXP
	} else {
		goto BASE
	}
BASE:
	return base == that, nil
REGEXP:
	re, err := regexp.Compile(base[3:])
	if err != nil {
		return false, err
	}
	return re.MatchString(that), nil
}

// filterItemsByRequestParams will filter elements in items by request parameters.
//
// The items must be a slice or an array, and the type of element in this items must be a
// struct or a pointer which point to a struct.
func filterItemsByRequestParams(r *http.Request, items interface{}) ([]interface{}, error) {
	if items == nil {
		return nil, nil
	}

	value := reflect.ValueOf(items)
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
	default:
		return nil, errors.New("items must be a slice of array")
	}
	if value.Len() == 0 {
		return []interface{}{}, nil
	}

	res := make([]interface{}, 0, 4)
LoopElement:
	for i := 0; i < value.Len(); i++ {
		element := value.Index(i)
		// parse real value of element by call .Elem() repeatedly,
		// stop when kind of value is a struct or not a pointer or a interface.
	Loop:
		for {
			switch kind := element.Kind(); kind {
			case reflect.Struct:
				break Loop
			case reflect.Ptr, reflect.Interface:
				element = element.Elem()
				continue
			default:
				continue LoopElement
			}
		}

		// add all fields and value to a map which has json tag and
		// type is basic type like string, int, float and bool.
		elementType := element.Type()
		fields := make(map[string]string, element.NumField())
		for j := 0; j < element.NumField(); j++ {
			field := elementType.Field(j)
			fieldValue := element.Field(j)
			tag, ok := field.Tag.Lookup("json")
			if !ok {
				continue
			}
			switch fieldValue.Kind() {
			case reflect.Bool:
				fields[tag] = strconv.FormatBool(fieldValue.Bool())
			case reflect.String:
				fields[tag] = fieldValue.String()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				fields[tag] = strconv.FormatInt(fieldValue.Int(), 10)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				fields[tag] = strconv.FormatUint(fieldValue.Uint(), 10)
			case reflect.Float32, reflect.Float64:
				fields[tag] = strconv.FormatFloat(fieldValue.Float(), 'g', -1, 64)
			default:
				continue
			}
		}

		for key := range r.URL.Query() {
			fv, ok := fields[key]
			if !ok {
				continue
			}
			targetValue := r.URL.Query().Get(key)
			ok, err := isEqual(targetValue, fv)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue LoopElement
			}
		}

		// keep original type
		res = append(res, value.Index(i).Interface())
	}

	return res, nil
}

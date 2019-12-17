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
	"net/http"

	"github.com/samaritan-proxy/sash/config"
)

func (s *Server) handleGetAllDependencies(w http.ResponseWriter, r *http.Request) {
	deps, err := s.depsCtl.GetAll()
	if err != nil {
		writeMsg(w, http.StatusInternalServerError, err.Error())
		return
	}
	result, err := filterItemsByRequestParams(r, deps)
	if err != nil {
		writeMsg(w, http.StatusBadRequest, err.Error())
		return
	}
	writePagedResp(w, r, result)
}

func (s *Server) handleAddDependency(w http.ResponseWriter, r *http.Request) {
	var (
		dep = new(config.Dependency)
		err = json.NewDecoder(r.Body).Decode(&dep)
	)
	if err != nil {
		goto BadRequest
	}
	if err = dep.Verify(); err != nil {
		goto BadRequest
	}
	if err = s.depsCtl.Set(dep); err != nil {
		goto InternalError
	}
	writeMsg(w, http.StatusOK, "OK")
	return

BadRequest:
	writeMsg(w, http.StatusBadRequest, err.Error())
InternalError:
	writeMsg(w, http.StatusInternalServerError, err.Error())
}

func (s *Server) handleGetDependency(w http.ResponseWriter, r *http.Request) {
	service, ok := s.getServiceAndAssertExist(w, r, s.depsCtl.Exist)
	if !ok {
		return
	}
	dep, err := s.depsCtl.Get(service)
	if err != nil {
		writeMsg(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, dep)
}

func (s *Server) handleUpdateDependency(w http.ResponseWriter, r *http.Request) {
	_, ok := s.getServiceAndAssertExist(w, r, s.depsCtl.Exist)
	if !ok {
		return
	}
	s.handleAddDependency(w, r)
}

func (s *Server) handleDeleteDependency(w http.ResponseWriter, r *http.Request) {
	service, ok := s.getServiceAndAssertExist(w, r, s.depsCtl.Exist)
	if !ok {
		return
	}
	if err := s.depsCtl.Delete(service); err != nil {
		writeMsg(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeMsg(w, http.StatusOK, "OK")
}

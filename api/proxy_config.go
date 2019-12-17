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

func (s *Server) handleGetAllProxyConfigs(w http.ResponseWriter, r *http.Request) {
	cfgs, err := s.proxyCfgCtl.GetAll()
	if err != nil {
		writeMsg(w, http.StatusInternalServerError, err.Error())
		return
	}
	result, err := filterItemsByRequestParams(r, cfgs)
	if err != nil {
		writeMsg(w, http.StatusBadRequest, err.Error())
		return
	}
	writePagedResp(w, r, result)
}

func (s *Server) handleAddProxyConfig(w http.ResponseWriter, r *http.Request) {
	var (
		cfg = new(config.ProxyConfig)
		err = json.NewDecoder(r.Body).Decode(&cfg)
	)
	if err != nil {
		goto BadRequest
	}
	if err = cfg.Verify(); err != nil {
		goto BadRequest
	}
	if err = s.proxyCfgCtl.Set(cfg); err != nil {
		goto InternalError
	}
	writeMsg(w, http.StatusOK, "OK")
	return

BadRequest:
	writeMsg(w, http.StatusBadRequest, err.Error())
InternalError:
	writeMsg(w, http.StatusInternalServerError, err.Error())
}

func (s *Server) handleGetProxyConfig(w http.ResponseWriter, r *http.Request) {
	service, ok := s.getServiceAndAssertExist(w, r, s.proxyCfgCtl.Exist)
	if !ok {
		return
	}
	cfg, err := s.proxyCfgCtl.Get(service)
	if err != nil {
		writeMsg(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, cfg)
}

func (s *Server) handleUpdateProxyConfig(w http.ResponseWriter, r *http.Request) {
	_, ok := s.getServiceAndAssertExist(w, r, s.proxyCfgCtl.Exist)
	if !ok {
		return
	}
	s.handleAddProxyConfig(w, r)
}

func (s *Server) handleDeleteProxyConfig(w http.ResponseWriter, r *http.Request) {
	service, ok := s.getServiceAndAssertExist(w, r, s.proxyCfgCtl.Exist)
	if !ok {
		return
	}
	if err := s.proxyCfgCtl.Delete(service); err != nil {
		writeMsg(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeMsg(w, http.StatusOK, "OK")
}

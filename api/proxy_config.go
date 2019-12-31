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

	"github.com/gorilla/mux"

	"github.com/samaritan-proxy/sash/config"
)

func (s *Server) handleGetAllProxyConfigs(w http.ResponseWriter, r *http.Request) {
	cfgs, err := s.proxyCfgCtl.GetAllCache()
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
	switch err = s.proxyCfgCtl.Add(cfg); err {
	case nil:
		writeMsg(w, http.StatusOK, "OK")
		return
	case config.ErrExist:
		goto BadRequest
	default:
		goto InternalError
	}
BadRequest:
	writeMsg(w, http.StatusBadRequest, err.Error())
InternalError:
	writeMsg(w, http.StatusInternalServerError, err.Error())
}

func (s *Server) handleGetProxyConfig(w http.ResponseWriter, r *http.Request) {
	service := mux.Vars(r)[paramService]
	cfg, err := s.proxyCfgCtl.Get(service)
	switch err {
	case nil:
		writeJSON(w, cfg)
	case config.ErrNotExist:
		writeMsg(w, http.StatusNotFound, fmt.Sprintf("service[%s] not found", service))
	default:
		writeMsg(w, http.StatusInternalServerError, err.Error())
	}
}

func (s *Server) handleUpdateProxyConfig(w http.ResponseWriter, r *http.Request) {
	var (
		service = mux.Vars(r)[paramService]
		cfg     = new(config.ProxyConfig)
		err     = json.NewDecoder(r.Body).Decode(&cfg)
	)
	if err != nil {
		writeMsg(w, http.StatusBadRequest, err.Error())
		return
	}
	cfg.ServiceName = service
	switch err = s.proxyCfgCtl.Update(cfg); err {
	case nil:
		writeMsg(w, http.StatusOK, "OK")
	case config.ErrNotExist:
		writeMsg(w, http.StatusNotFound, fmt.Sprintf("service[%s] not found", service))
	default:
		writeMsg(w, http.StatusInternalServerError, err.Error())
	}
}

func (s *Server) handleDeleteProxyConfig(w http.ResponseWriter, r *http.Request) {
	service := mux.Vars(r)[paramService]
	switch err := s.proxyCfgCtl.Delete(service); err {
	case nil:
		writeMsg(w, http.StatusOK, "OK")
	case config.ErrNotExist:
		writeMsg(w, http.StatusNotFound, fmt.Sprintf("service[%s] not found", service))
	default:
		writeMsg(w, http.StatusInternalServerError, err.Error())
	}
}

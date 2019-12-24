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
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	routeDependencies = "/dependencies"
	routeInstances    = "/instances"
	routeProxyConfigs = "/proxy-configs"
	routePing         = "/ping"

	paramPageNum  = "page_num"
	paramPageSize = "page_size"
	paramService  = "service"
	paramInstance = "instance"
)

func (s *Server) genProxyConfigsRouter(r *mux.Router) {
	r.HandleFunc("", s.handleGetAllProxyConfigs).Methods(http.MethodGet)
	r.HandleFunc("", s.handleAddProxyConfig).Methods(http.MethodPost)
	r.HandleFunc(fmt.Sprintf("/{%s}", paramService), s.handleGetProxyConfig).Methods(http.MethodGet)
	r.HandleFunc(fmt.Sprintf("/{%s}", paramService), s.handleUpdateProxyConfig).Methods(http.MethodPut)
	r.HandleFunc(fmt.Sprintf("/{%s}", paramService), s.handleDeleteProxyConfig).Methods(http.MethodDelete)
}

func (s *Server) genDependenciesRouter(r *mux.Router) {
	r.HandleFunc("", s.handleGetAllDependencies).Methods(http.MethodGet)
	r.HandleFunc("", s.handleAddDependency).Methods(http.MethodPost)
	r.HandleFunc(fmt.Sprintf("/{%s}", paramService), s.handleGetDependency).Methods(http.MethodGet)
	r.HandleFunc(fmt.Sprintf("/{%s}", paramService), s.handleUpdateDependency).Methods(http.MethodPut)
	r.HandleFunc(fmt.Sprintf("/{%s}", paramService), s.handleDeleteDependency).Methods(http.MethodDelete)
}

func (s *Server) genInstancesRouter(r *mux.Router) {
	r.HandleFunc("", s.handleGetAllInstances).Methods(http.MethodGet)
	r.HandleFunc(fmt.Sprintf("/{%s}", paramInstance), s.handleGetInstance).Methods(http.MethodGet)
}

func (s *Server) handlePing(w http.ResponseWriter, _ *http.Request) {
	writeMsg(w, http.StatusOK, "PONG")
}

func handleSubRoute(baseRouter *mux.Router, path string, fn func(router *mux.Router)) {
	fn(baseRouter.PathPrefix(path).Subrouter())
}

func (s *Server) genRouter() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc(routePing, s.handlePing)
	handleSubRoute(router, routeDependencies, s.genDependenciesRouter)
	handleSubRoute(router, routeInstances, s.genInstancesRouter)
	handleSubRoute(router, routeProxyConfigs, s.genProxyConfigsRouter)
	return router
}

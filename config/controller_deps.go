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
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
)

// Dependency is a wrapper of dependency info of service.
type Dependency struct {
	Metadata
	ServiceName  string   `json:"service_name"`
	Dependencies []string `json:"dependencies"`
}

// Verify this dependency.
func (d *Dependency) Verify() error {
	if len(d.ServiceName) == 0 {
		return fmt.Errorf("serivce_name is null")
	}
	return nil
}

// Dependencies is a slice of Dependency, impl the sort.Interface.
type Dependencies []*Dependency

func (d Dependencies) Len() int { return len(d) }

func (d Dependencies) Swap(i, j int) { d[i], d[j] = d[j], d[i] }

func (d Dependencies) Less(i, j int) bool { return d[i].ServiceName < d[j].ServiceName }

type DependenciesController struct {
	sync.RWMutex
	ctl          *Controller
	dependencies map[string]*Dependency
	handlers     atomic.Value //[]DependencyEventHandler
}

func (c *Controller) Dependencies() *DependenciesController {
	depCtl := &DependenciesController{
		ctl:          c,
		dependencies: make(map[string]*Dependency),
		handlers:     atomic.Value{},
	}
	c.RegisterEventHandler(depCtl.handleRawEvent)
	return depCtl
}

func (*DependenciesController) getNamespace() string { return NamespaceService }

func (*DependenciesController) getType() string { return TypeServiceDependency }

func (*DependenciesController) unmarshalDependency(b []byte) ([]string, error) {
	deps := make([]string, 0, 4)
	err := json.Unmarshal(b, &deps)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

func (*DependenciesController) marshallDependency(deps []string) ([]byte, error) {
	return json.Marshal(deps)
}

func (c *DependenciesController) loadHandlers() []DependencyEventHandler {
	hdls, ok := c.handlers.Load().([]DependencyEventHandler)
	if !ok {
		return nil
	}
	return hdls
}

func (c *DependenciesController) get(svc string, from func(svc string) ([]byte, error)) (*Dependency, error) {
	b, err := from(svc)
	if err != nil {
		return nil, err
	}
	var rawDeps []string
	if b != nil {
		_deps, err := c.unmarshalDependency(b)
		if err != nil {
			return nil, err
		}
		rawDeps = _deps
	}
	return &Dependency{
		ServiceName:  svc,
		Dependencies: rawDeps,
	}, nil
}

func (c *DependenciesController) Get(svc string) (*Dependency, error) {
	return c.get(svc, func(svc string) (bytes []byte, err error) {
		return c.ctl.Get(c.getNamespace(), c.getType(), svc)
	})
}

func (c *DependenciesController) GetCache(svc string) (*Dependency, error) {
	c.RLock()
	defer c.RUnlock()
	dep, ok := c.dependencies[svc]
	if !ok {
		return nil, ErrNotExist
	}
	return dep, nil
}

func (c *DependenciesController) Add(dependency *Dependency) error {
	if dependency == nil {
		return nil
	}
	if err := dependency.Verify(); err != nil {
		return err
	}
	b, err := c.marshallDependency(dependency.Dependencies)
	if err != nil {
		return err
	}
	return c.ctl.Add(c.getNamespace(), c.getType(), dependency.ServiceName, b)
}

func (c *DependenciesController) Update(dependency *Dependency) error {
	if dependency == nil {
		return nil
	}
	if err := dependency.Verify(); err != nil {
		return err
	}
	b, err := c.marshallDependency(dependency.Dependencies)
	if err != nil {
		return err
	}
	return c.ctl.Update(c.getNamespace(), c.getType(), dependency.ServiceName, b)
}

func (c *DependenciesController) Exist(svc string) bool {
	return c.ctl.Exist(c.getNamespace(), c.getType(), svc)
}

func (c *DependenciesController) Delete(svc string) error {
	return c.ctl.Del(c.getNamespace(), c.getType(), svc)
}

func (c *DependenciesController) getAll(getKeysFn func(string, string) ([]string, error), getFn func(string) (*Dependency, error)) (Dependencies, error) {
	svcs, err := getKeysFn(c.getNamespace(), c.getType())
	if err != nil {
		return nil, err
	}
	deps := Dependencies{}
	for _, svc := range svcs {
		dep, err := getFn(svc)
		if err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}
	return deps, nil
}

func (c *DependenciesController) GetAll() (Dependencies, error) {
	return c.getAll(c.ctl.Keys, c.Get)
}

func (c *DependenciesController) GetAllCache() (Dependencies, error) {
	return c.getAll(c.ctl.KeysCached, c.GetCache)
}

func (c *DependenciesController) handleRawEvent(event *Event) {
	if event.Config.Namespace != c.getNamespace() || event.Config.Type != c.getType() {
		return
	}

	svcName := event.Config.Key
	rawDeps, err := c.unmarshalDependency(event.Config.Value)
	if err != nil {
		return
	}

	var depEvt *DependencyEvent
	switch event.Type {
	case EventAdd:
		depEvt = &DependencyEvent{
			ServiceName: svcName,
			Add:         rawDeps,
		}
		defer func() {
			c.Lock()
			defer c.Unlock()
			c.dependencies[svcName] = &Dependency{
				ServiceName:  svcName,
				Dependencies: rawDeps,
			}
		}()
	case EventDelete:
		depEvt = &DependencyEvent{
			ServiceName: svcName,
			Del:         rawDeps,
		}
		defer func() {
			c.Lock()
			defer c.Unlock()
			delete(c.dependencies, svcName)
		}()
	case EventUpdate:
		var before, after, incr, decr []string
		after = rawDeps
		if dep, ok := c.dependencies[svcName]; ok && dep != nil {
			before = dep.Dependencies
		}
		incr, decr = diffSlice(before, after)
		depEvt = &DependencyEvent{
			ServiceName: svcName,
			Add:         incr,
			Del:         decr,
		}
		defer func() {
			c.Lock()
			defer c.Unlock()
			c.dependencies[svcName] = &Dependency{
				ServiceName:  "svcName",
				Dependencies: rawDeps,
			}
		}()
	}
	for _, hdl := range c.loadHandlers() {
		hdl(depEvt)
	}
}

func (c *DependenciesController) RegisterEventHandler(handler DependencyEventHandler) {
	c.Lock()
	defer c.Unlock()
	oldHdls := c.loadHandlers()
	newHdls := make([]DependencyEventHandler, 0, len(oldHdls)+1)
	copy(newHdls, newHdls)
	newHdls = append(newHdls, handler)
	c.handlers.Store(newHdls)
}

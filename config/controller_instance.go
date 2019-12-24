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
)

type Instance struct {
	Metadata
	ID            string `json:"id"`
	Hostname      string `json:"hostname"`
	IP            string `json:"ip"`
	Port          int    `json:"port"`
	Version       string `json:"version"`
	BelongService string `json:"belong_service"`
}

func (i *Instance) Verify() error {
	if len(i.ID) == 0 {
		return fmt.Errorf("id is null")
	}
	return nil
}

type Instances []*Instance

func (i Instances) Len() int { return len(i) }

func (i Instances) Swap(a, b int) { i[a], i[b] = i[b], i[a] }

func (i Instances) Less(a, b int) bool { return i[a].ID < i[b].ID }

type InstancesController struct {
	ctl *Controller
}

func (c *Controller) Instances() *InstancesController {
	return &InstancesController{ctl: c}
}

func (*InstancesController) getNamespace() string { return NamespaceSamaritan }

func (*InstancesController) getType() string { return TypeSamaritanInstance }

func (*InstancesController) unmarshalInstance(b []byte) (*Instance, error) {
	inst := new(Instance)
	if err := json.Unmarshal(b, inst); err != nil {
		return nil, err
	}
	return inst, nil
}

func (*InstancesController) marshalInstance(i *Instance) ([]byte, error) {
	return json.Marshal(i)
}

func (c *InstancesController) Get(id string) (*Instance, error) {
	b, err := c.ctl.Get(c.getNamespace(), c.getType(), id)
	if err != nil {
		return nil, err
	}
	return c.unmarshalInstance(b)
}

func (c *InstancesController) GetCache(id string) (*Instance, error) {
	b, err := c.ctl.GetCache(c.getNamespace(), c.getType(), id)
	if err != nil {
		return nil, err
	}
	return c.unmarshalInstance(b)
}

func (c *InstancesController) Add(inst *Instance) error {
	if inst == nil {
		return nil
	}
	if err := inst.Verify(); err != nil {
		return err
	}
	b, err := c.marshalInstance(inst)
	if err != nil {
		return err
	}
	return c.ctl.Add(c.getNamespace(), c.getType(), inst.ID, b)
}

func (c *InstancesController) Update(inst *Instance) error {
	if inst == nil {
		return nil
	}
	if err := inst.Verify(); err != nil {
		return err
	}
	b, err := c.marshalInstance(inst)
	if err != nil {
		return err
	}
	return c.ctl.Update(c.getNamespace(), c.getType(), inst.ID, b)
}

func (c *InstancesController) Exist(id string) bool {
	return c.ctl.Exist(c.getNamespace(), c.getType(), id)
}

func (c *InstancesController) Delete(id string) error {
	return c.ctl.Del(c.getNamespace(), c.getType(), id)
}

func (c *InstancesController) GetAll() (Instances, error) {
	ids, err := c.ctl.Keys(c.getNamespace(), c.getType())
	if err != nil {
		return nil, err
	}
	insts := Instances{}
	for _, id := range ids {
		inst, err := c.Get(id)
		if err != nil {
			return nil, err
		}
		insts = append(insts, inst)
	}
	return insts, nil
}

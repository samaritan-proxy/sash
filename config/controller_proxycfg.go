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

	"github.com/samaritan-proxy/samaritan-api/go/config/service"
)

// ProxyConfig is a wrapper of service.Config.
type ProxyConfig struct {
	Metadata
	ServiceName string          `json:"service_name"`
	Config      *service.Config `json:"config"`
}

// Verify this ProxyConfig.
func (c *ProxyConfig) Verify() error {
	if len(c.ServiceName) == 0 {
		return fmt.Errorf("serivce_name is null")
	}
	if err := c.Config.Validate(); err != nil {
		return err
	}
	return nil
}

// ProxyConfigs is a slice of ProxyConfig, impl the sort.Interface.
type ProxyConfigs []*ProxyConfig

func (c ProxyConfigs) Len() int { return len(c) }

func (c ProxyConfigs) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (c ProxyConfigs) Less(i, j int) bool { return c[i].ServiceName < c[j].ServiceName }

type ProxyConfigsController struct {
	ctl *Controller
}

func newProxyConfigController(ctl *Controller) *ProxyConfigsController {
	return &ProxyConfigsController{ctl: ctl}
}

func (*ProxyConfigsController) getNamespace() string { return NamespaceService }

func (*ProxyConfigsController) getType() string { return TypeServiceProxyConfig }

func (*ProxyConfigsController) unmarshalSvcCfg(b []byte) (*service.Config, error) {
	if b == nil {
		return nil, nil
	}
	cfg := new(service.Config)
	if err := cfg.UnmarshalJSON(b); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (*ProxyConfigsController) marshallSvcCfg(cfg *service.Config) ([]byte, error) {
	if cfg == nil {
		return nil, nil
	}
	return cfg.MarshalJSON()
}

func (c *ProxyConfigsController) get(svc string, from func(svc string) ([]byte, error)) (*ProxyConfig, error) {
	b, err := from(svc)
	if err != nil {
		return nil, err
	}
	var cfg *service.Config
	if b != nil {
		_cfg, err := c.unmarshalSvcCfg(b)
		if err != nil {
			return nil, err
		}
		cfg = _cfg
	}
	return &ProxyConfig{
		ServiceName: svc,
		Config:      cfg,
	}, nil
}

func (c *ProxyConfigsController) Get(svc string) (*ProxyConfig, error) {
	return c.get(svc, func(svc string) (bytes []byte, err error) {
		return c.ctl.Get(c.getNamespace(), c.getType(), svc)
	})
}

func (c *ProxyConfigsController) GetCache(svc string) (*ProxyConfig, error) {
	return c.get(svc, func(svc string) (bytes []byte, err error) {
		return c.ctl.GetCache(c.getNamespace(), c.getType(), svc)
	})
}

func (c *ProxyConfigsController) Add(cfg *ProxyConfig) error {
	if cfg == nil {
		return nil
	}
	if err := cfg.Verify(); err != nil {
		return err
	}
	b, err := c.marshallSvcCfg(cfg.Config)
	if err != nil {
		return err
	}
	return c.ctl.Add(c.getNamespace(), c.getType(), cfg.ServiceName, b)
}

func (c *ProxyConfigsController) Update(cfg *ProxyConfig) error {
	if cfg == nil {
		return nil
	}
	if err := cfg.Verify(); err != nil {
		return err
	}
	b, err := c.marshallSvcCfg(cfg.Config)
	if err != nil {
		return err
	}
	return c.ctl.Update(c.getNamespace(), c.getType(), cfg.ServiceName, b)
}

func (c *ProxyConfigsController) Exist(svc string) bool {
	return c.ctl.Exist(c.getNamespace(), c.getType(), svc)
}

func (c *ProxyConfigsController) Delete(svc string) error {
	return c.ctl.Del(c.getNamespace(), c.getType(), svc)
}

func (c *ProxyConfigsController) getAll(getKeysFn func(string, string) ([]string, error), getFn func(string) (*ProxyConfig, error)) (ProxyConfigs, error) {
	svcs, err := getKeysFn(c.getNamespace(), c.getType())
	if err != nil {
		return nil, err
	}
	cfgs := ProxyConfigs{}
	for _, svc := range svcs {
		cfg, err := getFn(svc)
		if err != nil {
			return nil, err
		}
		cfgs = append(cfgs, cfg)
	}
	return cfgs, nil
}

func (c *ProxyConfigsController) GetAll() (ProxyConfigs, error) {
	return c.getAll(c.ctl.Keys, c.Get)
}

func (c *ProxyConfigsController) GetAllCache() (ProxyConfigs, error) {
	return c.getAll(c.ctl.KeysCached, c.GetCache)
}

func (c *ProxyConfigsController) RegisterEventHandler(handler ProxyConfigEventHandler) {
	c.ctl.RegisterEventHandler(func(event *Event) {
		if event.Config.Namespace != c.getNamespace() || event.Config.Type != c.getType() {
			return
		}
		rawConf, err := c.unmarshalSvcCfg(event.Config.Value)
		if err != nil {
			return
		}
		handler(&ProxyConfigEvent{
			Type: event.Type,
			ProxyConfig: &ProxyConfig{
				ServiceName: event.Config.Key,
				Config:      rawConf,
			},
		})
	})
}

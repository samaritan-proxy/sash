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

package main

import (
	"time"

	"github.com/samaritan-proxy/sash/internal/zk"
)

type RawMessage struct {
	unmarshal func(interface{}) error
}

func (msg *RawMessage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	msg.unmarshal = unmarshal
	return nil
}

func (msg *RawMessage) Unmarshal(v interface{}) error {
	return msg.unmarshal(v)
}

type ConfigStore struct {
	Type     string        `yaml:"type"`
	Spec     interface{}   `yaml:"spec"`
	SyncFreq time.Duration `yaml:"sync_freq"`
}

func (c *ConfigStore) UnmarshalYAML(unmarshal func(interface{}) error) error {
	s := struct {
		Type     string        `yaml:"type"`
		Spec     *RawMessage   `yaml:"spec"`
		SyncFreq time.Duration `yaml:"sync_freq"`
	}{}
	if err := unmarshal(&s); err != nil {
		return err
	}

	c.Type = s.Type
	c.SyncFreq = s.SyncFreq

	// TODO: To improve the maintainability, we should unmarshal spec to
	// the actual structure when intializes the corresponding config store.
	switch s.Type {
	case "memory":
		c.Spec = nil
	case "zk":
		conf := new(zk.ConnConfig)
		if err := s.Spec.Unmarshal(conf); err != nil {
			return err
		}
		c.Spec = conf
	default:
		c.Spec = nil
	}
	return nil
}

type Registry struct {
	Type       string        `yaml:"type"`
	Spec       interface{}   `yaml:"spec"`
	SyncFreq   time.Duration `yaml:"sync_freq"`
	SyncJitter float64       `yaml:"sync_jitter"`
}

func (r *Registry) UnmarshalYAML(unmarshal func(interface{}) error) error {
	s := struct {
		Type       string        `yaml:"type"`
		Spec       *RawMessage   `yaml:"spec"`
		SyncFreq   time.Duration `yaml:"sync_freq"`
		SyncJitter float64       `yaml:"sync_jitter"`
	}{}
	if err := unmarshal(&s); err != nil {
		return err
	}

	r.Type = s.Type
	r.SyncFreq = s.SyncFreq
	r.SyncJitter = s.SyncJitter

	// TODO: To improve the maintainability, we should unmarshal spec to
	// the actual structure when intializes the corresponding service registry.
	switch s.Type {
	case "memory":
		r.Spec = nil
	case "zk":
		conf := new(zk.ConnConfig)
		if err := s.Spec.Unmarshal(conf); err != nil {
			return err
		}
		r.Spec = conf
	default:
		r.Spec = nil
	}
	return nil
}

type API struct {
	Bind string `yaml:"bind"`
}

type Discovery struct {
	Bind string `yaml:"bind"`
}

type Bootstrap struct {
	LogLevel    string      `yaml:"log_level"`
	API         API         `yaml:"api"`
	Discovery   Discovery   `yaml:"discovery"`
	ConfigStore ConfigStore `yaml:"config_store"`
	Registry    Registry    `yaml:"service_registry"`
}

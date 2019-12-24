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
)

type Config struct {
	Endpoint string        `yaml:"endpoint"`
	Interval time.Duration `yaml:"interval"`
}

type Registry struct {
	Endpoint   string        `yaml:"endpoint"`
	SyncFreq   time.Duration `yaml:"sync_freq"`
	SyncJitter float64       `yaml:"sync_jitter"`
}

type API struct {
	Bind string `yaml:"bind"`
}

type XdsRPC struct {
	Bind string `yaml:"bind"`
}

type Configs struct {
	Config   Config   `yaml:"config"`
	Registry Registry `yaml:"registry"`
	API      API      `yaml:"api"`
	XdsRPC   XdsRPC   `yaml:"xds_rpc"`
	LogLevel string   `yaml:"log_level"`
}

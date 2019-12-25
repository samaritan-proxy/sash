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
	"flag"
	"os"
	"time"

	"github.com/go-yaml/yaml"

	"github.com/samaritan-proxy/sash/logger"
)

var (
	bootstrap = Bootstrap{
		ConfigStore: ConfigStore{
			SyncFreq: time.Second * 5,
		},
		Registry: Registry{
			SyncFreq:   time.Second * 5,
			SyncJitter: 0.1,
		},
		API:       API{Bind: ":8882"},
		Discovery: Discovery{Bind: ":9090"},
		LogLevel:  "info",
	}

	configFile string
)

func parseFlags() {
	flag.StringVar(&configFile, "c", "./config.yaml", "config file")
	flag.Parse()

	f, err := os.Open(configFile)
	if err != nil {
		logger.Fatal(err)
	}

	if err = yaml.NewDecoder(f).Decode(&bootstrap); err != nil {
		logger.Fatal(err)
	}
}

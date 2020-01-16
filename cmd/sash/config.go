// Copyright 2020 Samaritan Authors
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
	"errors"
	"fmt"
	"log"

	"github.com/samaritan-proxy/sash/config"
	"github.com/samaritan-proxy/sash/config/zk"
)

func initConfigController(b *Bootstrap) *config.Controller {
	var (
		store config.Store
		err   error
	)

	switch typ := b.ConfigStore.Type; typ {
	case "memory":
		err = errors.New("memory config should only be used in tests")
	case "zk":
		store, err = zk.New(b.ConfigStore.Spec.(*zk.ConnConfig))
	default:
		err = fmt.Errorf("unsupported config store '%s'", typ)
	}

	if err != nil {
		log.Fatal(err)
	}

	ctl := config.NewController(
		store,
		config.SyncInterval(b.ConfigStore.SyncFreq),
	)
	return ctl
}

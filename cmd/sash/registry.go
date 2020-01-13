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

	"github.com/samaritan-proxy/sash/model"
	"github.com/samaritan-proxy/sash/registry"
	"github.com/samaritan-proxy/sash/registry/zk"
)

func initRegistry(b *Bootstrap) registry.Cache {
	var (
		reg model.ServiceRegistry
		err error
	)

	switch typ := b.Registry.Type; typ {
	case "memory":
		err = errors.New("memory registry should only be used in tests.")
	case "zk":
		reg, err = zk.NewDiscoveryClient(b.Registry.Spec.(*zk.ConnConfig))
	default:
		err = fmt.Errorf("unsupported service registry '%s'", typ)
	}

	if err != nil {
		log.Fatal(err)
	}

	options := []registry.CacheOption{
		registry.SyncFreq(b.Registry.SyncFreq),
		registry.SyncJitter(b.Registry.SyncJitter),
	}
	cache := registry.NewCache(reg, options...)
	return cache
}

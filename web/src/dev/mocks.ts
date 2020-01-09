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

import {Dependency} from '../models/dependency'
import {Instance} from "../models/instance";
import {ProxyConfig, ServiceConfig} from "../models/proxy-config";

export const mockDependencies: Dependency[] = [
    {
        service_name: "service1",
        create_time: "2019-12-3 14:00:00",
        update_time: "2019-12-3 14:00:00",
        dependencies: ["service3", "service4", "service5", "service6"],
    },
    {
        service_name: "service2",
        create_time: "2019-12-3 14:00:00",
        update_time: "2019-12-3 14:00:00",
        dependencies: ["service3", "service4"],
    }
];

export const mockInstances: Instance[] = [
    {
        id: "10.10.10.1_12345",
        hostname: "dbdd9df15e7f",
        ip: "10.10.10.1",
        port: 12345,
        version: "1.0",
        belong_service: "service_1",
        create_time: "2019-12-3 14:00:00",
        update_time: "2019-12-3 14:00:00",
    },
    {
        id: "10.10.10.2_12345",
        hostname: "92f604e8caad",
        ip: "10.10.10.2",
        port: 12345,
        version: "1.0",
        belong_service: "service_1",
        create_time: "2019-12-3 14:00:00",
        update_time: "2019-12-3 14:00:00",
    },
];

export const mockProxyConfigs: ProxyConfig[] = [
    {
        service_name: "service_1",
        config: {} as ServiceConfig,
        create_time: "2019-12-3 14:00:00",
        update_time: "2019-12-3 14:00:00",
    },
    {
        service_name: "service_2",
        config: {} as ServiceConfig,
        create_time: "2019-12-3 14:00:00",
        update_time: "2019-12-3 14:00:00",
    }
];

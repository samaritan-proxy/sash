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

import {SashResponse} from "../models/base";
import {Dependency} from "../models/dependency";
import {ajaxDelete, ajaxGet, ajaxPost, ajaxPut} from './ajax'
import {Instance} from "../models/instance";
import {ProxyConfig} from "../models/proxy-config";

const APIPrefix = "";
const dependenciesRoute = "dependencies";
const instanceRoute = "instances";
const proxyConfigRoute = "proxy-configs";

// note page start from 0
export function GetDependencies(page: number, service: string): Promise<SashResponse<Dependency[]>> {
    return ajaxGet(`${APIPrefix}/${dependenciesRoute}?service_name=${escape(service)}&page_num=${page}`)
}

export function GetDependency(service: string): Promise<Dependency> {
    return ajaxGet(`${APIPrefix}/${dependenciesRoute}/${service}`)
}

export function PutDependency(dependency: Dependency) {
    return ajaxPut(
        `${APIPrefix}/${dependenciesRoute}/${escape(dependency.service_name)}`,
        {dependencies: dependency.dependencies}
    )
}

export function PostDependency(dependency: Dependency) {
    return ajaxPost(`${APIPrefix}/${dependenciesRoute}`, dependency)
}

export function DeleteDependency(service: string) {
    return ajaxDelete(`${APIPrefix}/${dependenciesRoute}/${escape(service)}`)
}

export function GetInstances(page: number, id: string, belong_service: string): Promise<SashResponse<Instance[]>> {
    return ajaxGet(`${APIPrefix}/${instanceRoute}?id=${escape(id)}&page_num=${page}&belong_service=${belong_service}`)
}

export function GetInstance(id: string): Promise<Instance> {
    return ajaxGet(`${APIPrefix}/${instanceRoute}/${id}`)
}

export function GetProxyConfigs(page: number, service: string): Promise<SashResponse<ProxyConfig[]>> {
    return ajaxGet(`${APIPrefix}/${proxyConfigRoute}?service_name=${escape(service)}&page_num=${page}`)
}

export function GetProxyConfig(service: string): Promise<ProxyConfig> {
    return ajaxGet(`${APIPrefix}/${proxyConfigRoute}/${service}`)
}

export function PostProxyConfig(cfg: ProxyConfig) {
    return ajaxPost(`${APIPrefix}/${proxyConfigRoute}`, cfg)
}

export function PutProxyConfig(cfg: ProxyConfig) {
    return ajaxPut(`${APIPrefix}/${proxyConfigRoute}/${cfg.service_name}`, cfg)
}

export function DeleteProxyConfig(service: string) {
    return ajaxDelete(`${APIPrefix}/${proxyConfigRoute}/${service}`)
}

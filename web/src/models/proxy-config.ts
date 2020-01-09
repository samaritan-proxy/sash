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

import {Time} from "./base";

export interface ProxyConfig extends Time {
    service_name: string
    config: ServiceConfig
}

export enum Protocol {
    TCP = "TCP",
    MySQL = "MySQL",
    Redis = "Redis"
}

export enum LbPolicy {
    ROUND_ROBIN = "ROUND_ROBIN",
    LEAST_CONNECTION = "LEAST_CONNECTION",
    RANDOM = "RANDOM",
    CLUSTER_PROVIDED = "CLUSTER_PROVIDED"
}

export type Duration = string

export interface Address {
    ip: string
    port: number
}

export interface Listener {
    address: Address
    connectionLimit: number
}

export interface TCPChecker {
}

export interface Action {
    // base64 encoded
    send: string
    // base64 encoded
    expect: string
}

export interface ATCPChecker {
    action: TCPChecker[]
}

export interface MySQLChecker {
    username: string
}

export interface RedisChecker {
    password: string
}

export interface HealthCheck {
    interval: Duration,
    timeout: Duration,
    fallThreshold: number,
    riseThreshold: number,
    tcpChecker?: TCPChecker,
    atcpChecker?: ATCPChecker,
    mysqlChecker?: MySQLChecker,
    redisChecker?: RedisChecker
}

export interface ServiceConfig {
    listener: Listener
    healthCheck: HealthCheck
    connectTimeout: Duration,
    idleTimeout: Duration
    lbPolicy: LbPolicy
    protocol: Protocol
    tcpOption?: object
    redisOption?: object
    mysqlOption?: object
}

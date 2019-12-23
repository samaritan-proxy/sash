# Sash APIs

## Introduction
   
### Status Code

When the `Status Code` is `200`, the `Content-Type` is `application/json` and the body is a json object.
When the `Status Code` is `40X` or `50X`, the `Content-Type` is `text/plain`, and the body is an error message.

### Models

#### Instance

| name           | type   | description             |
| -------------- | ------ | ----------------------- |
| create_time    | string | create time             |
| update_time    | string | update time             |
| id             | string | instance ID             |
| hostname       | string | instance hostname       |
| ip             | string | instance IP             |
| port           | int    | instance port           |
| version        | string | instance version        |
| belong_service | string | instance belong service |

#### Dependency

| name         | type     | description                  |
| ------------ | -------- | ---------------------------- |
| create_time  | string   | create time                  |
| update_time  | string   | update time                  |
| service_name | string   | service name                 |
| dependencies | []string | dependencies of this service |

#### ProxyConfig 

| name         | type   | description                                                           |
| ------------ | ------ | --------------------------------------------------------------------- |
| create_time  | string | create time                                                           |
| update_time  | string | update time                                                           |
| service_name | string | service name                                                          |
| config       | object | [Reference](https://samaritan-proxy.github.io/docs/proto-ref/#config) |

## `GET` /ping

### Response
 
- body: PONG

## `GET` /instances

### Description

Get all instances.

### Parameters

#### Query Parameters

| name           | type   | require | default | description                        |
| -------------- | ------ | ------- | ------- | ---------------------------------- |
| page_num       | int    | false   | 0       | page number                        |
| page_size      | int    | false   | 0       | page size                          |
| id             | string | false   |         | filter instances by id             |
| hostname       | string | false   |         | filter instances by hostname       |
| ip             | string | false   |         | filter instances by ip             |
| port           | string | false   |         | filter instances by port           |
| version        | string | false   |         | filter instances by version        |
| belong_service | string | false   |         | filter instances by belong service |

### Response

- body:

    | name      | type       | description                     |
    | --------- | ---------- | ------------------------------- |
    | page_num  | int        | current page number             |
    | page_size | int        | current page size               |
    | total     | int        | total items count               |
    | data      | []Instance | [Instance Reference](#Instance) |

### Example

#### Request

`curl http://sash/instances?id=inst_1`

#### Response

```json5
{
  "data": [
    {
      "create_time":"0001-01-01T00:00:00Z",
      "update_time":"0001-01-01T00:00:00Z",
      "id": "inst_1",
      "hostname": "test_host",
      "ip": "1.1.1.1",
      "port": 12345,
      "version": "0.0.1",
      "belong_service": "test_svc"
    }
  ],
  "page_num": 0,
  "page_size": 10,
  "total":1
}
```

## `GET` /instances/:instance

### Description

Get an instance by instance id.

### Response

- header:
    - Content-Type: application/json

- body: [Instance Reference](#Instance)

### Example

#### Request

`curl http://sash/instance/inst_1`

#### Response

```json5
{
  "create_time":"0001-01-01T00:00:00Z",
  "update_time":"0001-01-01T00:00:00Z",
  "id": "inst_1",
  "hostname": "test_host",
  "ip": "1.1.1.1",
  "port": 12345,
  "version": "0.0.1",
  "belong_service": "test_svc"
}
```

## `GET` /dependencies

### Description

Get all dependencies

### Parameters

#### Query Parameters

| name         | type   | require | default | description                      |
| ------------ | ------ | ------- | ------- | -------------------------------- |
| page_num     | int    | false   | 0       | page number                      |
| page_size    | int    | false   | 0       | page size                        |
| service_name | string | false   |         | filter instances by service name |

### Response:

- header:
    - Content-Type: application/json

- body:

    | name      | type         | description                         |
    | --------- | ------------ | ----------------------------------- |
    | page_num  | int          | current page number                 |
    | page_size | int          | current page size                   |
    | total     | int          | total items count                   |
    | data      | []Dependency | [Dependency Reference](#Dependency) |

### Example

#### Request

`curl http://sash/dependencies/inst_1`

#### Response

```json5
{
  "data": [
    {
      "create_time":"0001-01-01T00:00:00Z",
      "update_time":"0001-01-01T00:00:00Z",
      "service_name": "svc_1",
      "dependencies": ["dep_1", "dep_2"]
    }
  ],
  "page_num": 0,
  "page_size": 10,
  "total":1
}
```

## `POST` /dependencies

### Description

Add a dependency.

### Parameters

#### Header

- Content-Type: application/json

#### Body

| name         | type   | require | default | description                  |
| ------------ | ------ | ------- | ------- | ---------------------------- |
| service_name | string | true    |         | service name                 |
| dependencies | array  | true    |         | dependencies of this service |

### Response

- body: OK

### Example

#### Request

`curl -X POST -H 'Content-Type: application/json' -d '{"service_name": "svc_1", "dependencies": ["dep_1", "dep_2"]}' http://sash/dependencies`

#### Response

`OK`

## `GET` /dependencies/:service

### Description

Get a dependency by service name.

### Response

- header:
    - Content-Type: application/json
    
- body: [Dependency Reference](#Dependency)

### Example

#### Request

`curl http://sash/dependencies/svc_1`

#### Response

```json5
{
    "create_time":"0001-01-01T00:00:00Z",
    "update_time":"0001-01-01T00:00:00Z",
    "service_name": "svc_1",
    "dependencies": ["dep_1", "dep_2"]
}
```

## `PUT` /dependencies/:service

### Description

Update a dependency by service name.

### Parameters

#### Header

- Content-Type: application/json

#### Body

| name         | type  | require | default | description                  |
| ------------ | ----- | ------- | ------- | ---------------------------- |
| dependencies | array | true    |         | dependencies of this service |

### Response

- body: OK

### Example

#### Request

`curl -X PUT -H 'Content-Type: application/json' -d '{"dependencies": ["dep_1", "dep_2"]}' http://sash/dependencies/svc_1`

#### Response

`OK`

## `DELETE` /dependencies/:service

### Description

Delete a dependency by service name.

### Response

- body: OK

### Example

#### Request

`curl -X DELETE http://sash/dependencies/svc_1`

#### Response

`OK`

## `GET` /proxy-configs

### Description

Get all proxy configs.

### Parameters

#### Query Parameters

| name         | position | type   | require | default | description                      |
| ------------ | -------- | ------ | ------- | ------- | -------------------------------- |
| page_num     | query    | int    | false   | 0       | page number                      |
| page_size    | query    | int    | false   | 0       | page size                        |
| service_name | query    | string | false   |         | filter instances by service name |

### Response

- header:
    - Content-Type: application/json

- body:

    | name      | type          | description                           |
    | --------- | ------------- | ------------------------------------- |
    | page_num  | int           | current page number                   |
    | page_size | int           | current page size                     |
    | total     | int           | total items count                     |
    | data      | []ProxyConfig | [ProxyConfig Reference](#ProxyConfig) |

### Example

#### Request

`curl http://sash/proxy-configs?service_name=svc_1`

#### Response

```json5
{
  "data": [
    {
      "create_time":"0001-01-01T00:00:00Z",
      "update_time":"0001-01-01T00:00:00Z",
      "service_name": "svc_1",
      "config": null
    }
  ],
  "page_num": 0,
  "page_size": 10,
  "total":1
}
```

## `POST` /proxy-configs

### Description

Add a proxy config.

### Parameters

#### Header

- Content-Type: application/json

#### Body

| name         | type   | require | default | description                                                           |
| ------------ | ------ | ------- | ------- | --------------------------------------------------------------------- |
| service_name | string | true    |         | service name                                                          |
| config       | object | true    |         | [Reference](https://samaritan-proxy.github.io/docs/proto-ref/#config) |

### Response

- body: OK

### Example

#### Request

`curl -X POST -H 'Content-Type: application/json' -d '{"service_name": "svc_1", "config": {"protocol": "Redis"}}' http://sash/proxy-configs`

#### Response

`OK`

## `GET` /proxy-configs/:service

### Description

Get a proxy config by service name.

### Response

- header:
    - Content-Type: application/json

- body: [ProxyConfig Reference](#ProxyConfig)

### Example

#### Request

`curl http://sash/proxy-configs/svc_1`

#### Response

```json5
{
    "create_time":"0001-01-01T00:00:00Z",
    "update_time":"0001-01-01T00:00:00Z",
    "service_name": "svc_1",
    "config": null
}
```

## `PUT` /proxy-configs/:service

### Description

Update a proxy config by service name.

### Parameters

#### Header

- Content-Type: application/json

#### Body

| name   | type   | require | default | description                                                           |
| ------ | ------ | ------- | ------- | --------------------------------------------------------------------- |
| config | object | true    |         | [Reference](https://samaritan-proxy.github.io/docs/proto-ref/#config) |

### Response

- body: OK

### Example

#### Request

`curl -X PUT -H 'Content-Type: application/json' -d '{"config": {"protocol": "TCP"}}' http://sash/proxy-configs/svc_1`

#### Response

`OK`

## `DELETE` /proxy-configs/:service

### Description

Delete a proxy config by service name.

### Response

- body: OK

### Example

#### Request

`curl -X DELETE http://sash/proxy-configs/svc_1`

#### Response

`OK`
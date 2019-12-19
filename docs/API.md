# Sash API

## About
- type:
  - `regexp` : If the prefix is "re:", regular expression matching mode is enabled.

## GET /ping
- response:
  - 200: 
    - body: PONG

## GET /instances
- description: Get all instance
- parameters:
  - query:
    - page_num:
      - type: int
      - default: 0
      - description: page number
    - page_size:
      - type: int
      - default: 0
      - description: page size
    - id:
      - type: string, regexp
      - default: 
      - description: filter instances by id.
    - hostname:
      - type: string, regexp
      - default: 
      - description: filter instances by hostname.
    - ip:
      - type: string, regexp
      - default: 
      - description: filter instances by ip.
    - port:
      - type: int, regexp
      - default: 
      - description: filter instances by port.
    - version:
      - type: string, regexp
      - default: 
      - description: filter instances by version.
    - belong_service:
      - type: string, regexp
      - default: 
      - description: filter instances by belong service.
- response:
  - 500:
    - body: ${ERROR_MESSAGE}
  - 200:
    - header:
      - Content-Type: application/json
    - body:
        ```json
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

## GET /instances/{instance}
- description: Get an instance by instance id.
- parameters:
  - path:
    - instance:
      - type: string
      - require: true
      - description: the id of the instance you want to get.
- response:
  - 404:
    - description: instance not exist
    - body: instance[{instance}] not found
  - 500:
    - body: ${ERROR_MESSAGE}
  - 200:
    - header:
      - Content-Type: application/json
    - body:
        ```json
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

## GET /dependencies
- description: Get all dependencies
- parameters:
  - query:
    - page_num:
      - type: int
      - default: 0
      - description: page number
    - page_size:
      - type: int
      - default: 0
      - description: page size
    - service_name:
      - type: string, regexp
      - default: 
      - description: filter dependencies by service name.
- response:
  - 500:
    - body: ${ERROR_MESSAGE}
  - 200:
    - header:
      - Content-Type: application/json
    - body:
        ```json
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

## POST /dependencies
- description: add a dependency
- parameters:
  - body:
    ```json5
    {
        "service_name": "svc_1", // can not be null
        "dependencies": ["dep_1", "dep_2"]
    }
    ```
- response:
  - 400:
    - description: error in request
    - body: ${ERROR_MESSAGE}
  - 500:
    - body: ${ERROR_MESSAGE}
  - 200:
    - body: OK

## GET /dependencies/{service}
- description: Get a dependency by service name.
- parameters:
  - path:
    - service:
      - type: string
      - require: true
      - description: the name of the serivce you want to get.
- response:
  - 404:
    - description: service not exist
    - body: service[{service}] not found
  - 500:
    - body: ${ERROR_MESSAGE}
  - 200:
    - header:
      - Content-Type: application/json
    - body:
        ```json
        {
            "create_time":"0001-01-01T00:00:00Z",
            "update_time":"0001-01-01T00:00:00Z",
            "service_name": "svc_1",
            "dependencies": ["dep_1", "dep_2"]
        }
        ```

## PUT /dependencies/{service}
- description: Update a dependency by service name.
- parameters:
  - path:
    - service:
      - type: string
      - require: true
      - description: the name of the serivce you want to update.
  - body:
    ```json5
    {
        "dependencies": ["dep_1", "dep_2"]
    }
    ```
- response:
  - 400:
    - description: error in request
    - body: ${ERROR_MESSAGE}
  - 404:
    - description: service not exist
    - body: service[{service}] not found
  - 500:
    - body: ${ERROR_MESSAGE}
  - 200:
    - body: OK

## DELETE /dependencies/{service}
- description: delete a dependency by service name.
- parameters:
  - path:
    - service:
      - type: string
      - require: true
      - description: the name of the serivce you want to delete.
- response:
  - 404:
    - description: service not exist
    - body: service[{service}] not found
  - 500:
    - body: ${ERROR_MESSAGE}
  - 200:
    - body: OK

## GET /proxy-config
- description: Get all proxy configs.
- parameters:
  - query:
    - page_num:
      - type: int
      - default: 0
      - description: page number
    - page_size:
      - type: int
      - default: 0
      - description: page size
    - service_name:
      - type: string, regexp
      - default: 
      - description: filter proxy configs by service name.
- response:
  - 500:
    - body: ${ERROR_MESSAGE}
  - 200:
    - header:
      - Content-Type: application/json
    - body:
        ```json
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

## POST /proxy-config
- description: add a proxy config
- parameters:
  - body:
    ```json5
    {
        "service_name": "svc_1", // can not be null
        "config": {}
    }
    ```
- response:
  - 400:
    - description: error in request
    - body: ${ERROR_MESSAGE}
  - 500:
    - body: ${ERROR_MESSAGE}
  - 200:
    - body: OK

## GET /proxy-config/{service}
- description: Get a proxy config by service name.
- parameters:
  - path:
    - service:
      - type: string
      - require: true
      - description: the name of the proxy config you want to get.
- response:
  - 404:
    - description: service not exist
    - body: service[{service}] not found
  - 500:
    - body: ${ERROR_MESSAGE}
  - 200:
    - header:
      - Content-Type: application/json
    - body:
        ```json
        {
            "create_time":"0001-01-01T00:00:00Z",
            "update_time":"0001-01-01T00:00:00Z",
            "service_name": "svc_1",
            "config": null
        }
        ```

## PUT /proxy-config/{service}
- description: Update a proxy config by service name.
- parameters:
  - path:
    - service:
      - type: string
      - require: true
      - description: the name of the proxy config you want to update.
  - body:
    ```json5
    {
        "config": {}
    }
    ```
- response:
  - 400:
    - description: error in request
    - body: ${ERROR_MESSAGE}
  - 404:
    - description: service not exist
    - body: service[{service}] not found
  - 500:
    - body: ${ERROR_MESSAGE}
  - 200:
    - body: OK

## DELETE /proxy-config/{service}
- description: delete a proxy config by service name.
- parameters:
  - path:
    - service:
      - type: string
      - require: true
      - description: the name of the proxy config you want to delete.
- response:
  - 404:
    - description: service not exist
    - body: service[{service}] not found
  - 500:
    - body: ${ERROR_MESSAGE}
  - 200:
    - body: OK
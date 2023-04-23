<p align="center">
<img src="https://github.com/marle3003/mokapi/raw/v0.5.0/logo.svg" alt="Mokapi" title="Mokapi" width="300" />
</p>

<h3 align="center">Create and test API designs<br />before actually building them</h3>

<p align="center">
![build](https://github.com/marle3003/mokapi/actions/workflows/build.yml/badge.svg)
[![codecov](https://codecov.io/gh/marle3003/mokapi/branch/master/graph/badge.svg?token=C7X690ZOBI)](https://codecov.io/gh/marle3003/mokapi)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/marle3003/mokapi/blob/master/LICENSE)
</p>

## Overview

Mokapi is a tool that allows developers to create and test API designs before actually building them. It works by emulating API responses based on the designs that have been created, which allows developers to test different configurations and edge cases without having to rely on a fully built API.

With Mokapi, developers can make changes to their API configurations on the fly, which can help to speed up the testing process and reduce dependencies. This means that developers can quickly and easily test different scenarios, such as delayed or failed responses, without having to rely on a fully functional API.

By using Mokapi, developers can reduce the time and effort required for development and testing, since they can test different scenarios and configurations in a more efficient way. This can help to improve the quality of the API and reduce the risk of bugs or errors in production.

## Web UI

<img src="webui.png" alt="Mokapi Web UI" title="Mokapi Web UI" />

## Usage

```shell
docker run --env 'MOKAPI_Services_Swagger-Petstore_Config_Url'='https://raw.githubusercontent.com/OAI/OpenAPI-Specification/main/examples/v3.0/petstore.yaml' \
  --env 'MOKAPI_Services_Swagger-Petstore_Http_Servers[0]_Url'='http://:80' \
  -p 80:80 -p 8080:8080 \
  mokapi/mokapi:latest
```

## Documentation

- [Get Started](https://mokapi.io/docs/guides/get-started/welcome)
- [HTTP](https://mokapi.io/docs/guides/http/overview)
- [Kafka](https://mokapi.io/docs/guides/kafka/overview)
- [LDAP](https://mokapi.io/docs/guides/ldap/overview)
- [SMTP](https://mokapi.io/docs/guides/smtp/overview)
- [Javascript API](https://mokapi.io/docs/javascript-api)
- [Examples](https://mokapi.io/docs/examples)
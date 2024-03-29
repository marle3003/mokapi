---
title: Mokapi Dashboard to analyze and inspect your requests and responses
description: Mokapi provides you a dashboard where you can view your mocked APIs or analyze traffic like a HTTP request/response.
---
# Dashboard

Mokapi provides you a dashboard where you can view your 
mocked APIs or analyze traffic like a HTTP request/response.

<img src="/docs/guides/get-started/dashboard.png" width="700" alt="Mokapi Dashboard" title="Mokapi Dashboard" />

## API

Mokapi exposes a set of information through an API handler, 
such as the configurations, logs and metrics.

### Endpoints

You can find the OpenAPI specification of these endpoints 
[here](https://github.com/marle3003/mokapi/blob/master/examples/mokapi/dashboard.yml)

| Path                              | Description                            |
|-----------------------------------|----------------------------------------|
| /api/info                         | get information about mokapi's runtime |
 | /api/services                     | list of all services                   |
 | /api/services/http/{name}         | Information about the http service     |
 | /api/services/kafka/{name}        | Information about the kafka cluster    |
 | /api/services/smtp/{name}         | Information about the smtp server      |
 | /api/services/kafka/{name}/groups | list of kafka groups                   |
 | /api/events                       | list of events                         |
 | /api/events/{id}                  | get event by id                        |
 | /api/metrics                      | get list of metrics                    |
 | /api/schema/example               | returns example for given schema       |
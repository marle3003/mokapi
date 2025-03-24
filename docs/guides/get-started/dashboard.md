---
title: Using Mokapi Dashboard
description: Mokapi provides you a dashboard where you can view your mock APIs or analyze traffic like a HTTP request/response.
---
# Using Mokapi Dashboard

The dashboard provides graphical displays of your mock APIs and helps developers 
simulate real API responses during the development and testing phases. It allows 
you to validate data against the specification as well as generate new random, realistic and valid data.
This is useful during the early stages of development, facilitating rapid prototyping, testing error handling, 
and ensuring the mock APIs works seamlessly under various scenarios.

The dashboard is available on port 8080 by default. Open `http://localhost:8080` from your browser.

<img src="/dashboard-overview-mock-api.jpg" width="700" alt="Dashboard shows the current mock APIs and runtime metrics of Mokapi." title="" />

## Working with Dashboard

The dashboard displays key metrics like request logs, 
response times, and error rates, offering a comprehensive 
overview of how the mock API behaves under different conditions.

### Generate sample data from the specification

Some API types support generating sample data and validating data against the specification 
(namely HTTP and Kafka). Here is a quick overview with the HTTP example. To create sample data for a 
request body, navigate to an HTTP endpoint and look for the "Example & Validate" button.
In the opened dialog you can create, edit, validate, copy to the clipboard and download 
your sample data.

<img src="/dashboard-rest-api-endpoint-example-validate.jpg" alt="Localizes the sample and validation functionality in the REST API endpoint view." />

### Validate data

In the same dialog where we can create sample data, we can also validate it. The next image
shows an example of an HTTP REST request body that is not valid according to the specification.

<img src="/dashboard-rest-api-data-validation-with-error.jpg" width="700" alt="Indicates an error validating an HTTP REST request body" />

## How to customize?

There are several options to customize the dashboard to fit your environment.

### How to change the default port?

You can change the default port on which the dashboard should run as follows:

```bash tab=CLI
--api-port 5000
```
```bash tab=Env
MOKAPI_API_PORT=5000
```
```yaml tab=File (YAML)
api:
  port: 5000
```

### Adding a Path Prefix

Many applications are not hosted in the root (/) of their domain. For example,  
Mokapi's Dashboard could live at `example.com/mokapi`. In this case, a prefix must be 
added to the dashboard, so a link to `/dashboard` should be rewritten as `/mokapi/dashboard`.

```bash tab=CLI
--api-path /mokapi
```
```bash tab=Env
MOKAPI_API_PATH=/mokapi/dashboard
```
```yaml tab=File (YAML)
api:
 path: /mokapi
```

### Behind a Reverse Proxy

If the dashboard is running behind a reverse proxy and the proxy strips segments from the request 
URL, you must set the base path. For example, the URL to the dashboard is `example.com/app/mokapi`
but the request URL to your host running the dashboard is `/`

```bash tab=CLI
--api-base /app/mokapi
```
```bash tab=Env
MOKAPI_API_BASE=/app/mokapi
```
```yaml tab=File (YAML)
api:
 base: /app/mokapi
```

Further configuration options are documented [here](/docs/configuration/reference.md).

## API

Mokapi exposes a set of information through an API handler, 
such as the configurations, logs and metrics. The API is available at the same port 
as the dashboard but on the path /api

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
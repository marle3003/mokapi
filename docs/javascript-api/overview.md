---
title: JavaScript API Documentation
description: Explore Mokapi’s JavaScript API to extend its functionality. Learn how to use modules for HTTP, Kafka, YAML, Mustache, and more.
---
# Mokapi JavaScript API

Mokapi provides a powerful JavaScript API that allows you to extend its functionality. 
This documentation outlines the available modules and their capabilities.
To learn how Mokapi handles module imports, see [Modules](/docs/javascript-api/modules.md).

``` box=tip url=[@types/mokapi on npm](https://www.npmjs.com/package/@types/mokapi)
Mokapi offers a TypeScript definition package. Install it using:  
`npm install @types/mokapi --save-dev` 
```

## Available Modules

### mokapi (Core API)

Provides core functions for scheduling jobs, handling events, and accessing environment variables.

| Functions                                                                    | Description                                           |
|------------------------------------------------------------------------------|-------------------------------------------------------|
| [cron( expression, handler, \[args\] )](/docs/javascript-api/mokapi/cron.md) | Schedules a new periodic job using a cron expression. |
| [date( \[args\] )](/docs/javascript-api/mokapi/date.md)                      | Returns a formatted date string.                      |
| [env( name )](/docs/javascript-api/mokapi/env.md)                            | Gets the value of an environment variable.            |
| [every( interval, handler, \[args\] )](/docs/javascript-api/mokapi/every.md) | Runs a periodic job at a fixed interval.              |
| [on( event, handler, \[args\]](/docs/javascript-api/mokapi/on.md) )          | Registers an event handler.                           |
| [sleep( time )](/docs/javascript-api/mokapi/sleep.md)                        | Pauses execution for a specified duration.            |
| [marshal( value, \[encoding\] )](/docs/javascript-api/mokapi/marshal.md)     | Converts a value to a marshaled string.               |

### mokapi/http (HTTP Requests)

Functions to send HTTP requests within Mokapi scripts.

| Functions                                                                         | Description                   |
|-----------------------------------------------------------------------------------|-------------------------------|
| [get( url, \[args\] )](/docs/javascript-api/mokapi-http/get.md)                   | Sends an HTTP GET request.    |
| [post( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/post.md)       | Sends an HTTP POST request    |
| [put( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/put.md)         | Sends an HTTP PUT request     |
| [head( url, \[args\] )](/docs/javascript-api/mokapi-http/head.md)                 | Sends an HTTP HEAD request    |
| [patch( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/patch.md)     | Sends an HTTP PATCH request   |
| [delete( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/delete.md)   | Sends an HTTP DELETE request  |
| [options( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/options.md) | Sends an HTTP OPTIONS request |

### mokapi/faker (Mock Data Generator)

Generates random test data based on a schema.

| Functions                                                   | Description                             |
|-------------------------------------------------------------|-----------------------------------------|
| [fake( schema )](/docs/javascript-api/mokapi-faker/fake.md) | Generates mock data based on a schema.  |

### mokapi/kafka (Kafka Messaging)

Functions for interacting with Kafka topics.

| Functions                                                           | Description                            |
|---------------------------------------------------------------------|----------------------------------------|
| [produce( \[args\] )](/docs/javascript-api/mokapi-kafka/produce.md) | Publishes a message to a Kafka topic.  |

### mokapi/mustache (Template Engine)

Processes Mustache templates with dynamic data.

| Functions                                                                   | Description                                      |
|-----------------------------------------------------------------------------|--------------------------------------------------|
| [render( template, scope )](/docs/javascript-api/mokapi-mustache/render.md) | Renders a Mustache template with provided data.  |

### mokapi/yaml (YAML Parsing & Serialization)

Handles YAML data parsing and conversion.

| Functions                                                           | Description                                     |
|---------------------------------------------------------------------|-------------------------------------------------|
| [parse( text )](/docs/javascript-api/mokapi-yaml/parse.md)          | Parses a YAML string into a JavaScript object.  |
| [stringify( value )](/docs/javascript-api/mokapi-yaml/stringify.md) | Converts a JavaScript object into YAML format.  |

### mokapi/encoding (Encoding Utilities)

Functions for encoding and decoding data.

| Functions                                                                       | Description                       |
|---------------------------------------------------------------------------------|-----------------------------------|
| [base64.encode( input )](/docs/javascript-api/mokapi-encoding/base64-encode.md) | Encodes a string using Base64.    |
| [base64.decode( input )](/docs/javascript-api/mokapi-encoding/base64-decode.md) | Decodes a Base64-encoded string.  |







---
title: Mokapi Javascript API
description: Provides the documentation of the Mokapi Javascript API.
---
# Mokapi Javascript API

The list of Mokapi's Javascript modules usable to extend Mokapi's behavior.
To learn how Mokapi imports modules, see [Modules](/docs/javascript-api/modules.md).

## mokapi

| Functions                                                                    | Description                                                   |
|------------------------------------------------------------------------------|---------------------------------------------------------------|
| [cron( expression, handler, \[args\] )](/docs/javascript-api/mokapi/cron.md) | Schedules a new periodic job using cron expression.           |
| [date( \[args\] )](/docs/javascript-api/mokapi/date.md)                      | Returns a textual representation of the date.                 |
| [env( name )](/docs/javascript-api/mokapi/env.md)                            | Gets the value of an environment variable.                    |
| [every( interval, handler, \[args\] )](/docs/javascript-api/mokapi/every.md) | Schedules a new periodic job with interval.                   |
| [on( event, handler, \[args\]](/docs/javascript-api/mokapi/on.md) )          | Attaches an event handler for the given event.                |
| [sleep( time )](/docs/javascript-api/mokapi/sleep.md)                        | Suspends the execution for the specified duration.            |
| [marshal( value, \[encoding\] )](/docs/javascript-api/mokapi/marshal.md)     | Returns marshalled string representation of value (>= v0.9.7) |

## mokapi/http

| Functions                                                                         | Description                    |
|-----------------------------------------------------------------------------------|--------------------------------|
| [get( url, \[args\] )](/docs/javascript-api/mokapi-http/get.md)                   | Issues an HTTP GET request     |
| [post( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/post.md)       | Issues an HTTP POST request    |
| [put( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/put.md)         | Issues an HTTP PUT request     |
| [head( url, \[args\] )](/docs/javascript-api/mokapi-http/head.md)                 | Issues an HTTP HEAD request    |
| [patch( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/patch.md)     | Issues an HTTP PATCH request   |
| [delete( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/delete.md)   | Issues an HTTP DELETE request  |
| [options( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/options.md) | Issues an HTTP OPTIONS request |

## faker

| Functions                                                   | Description                                |
|-------------------------------------------------------------|--------------------------------------------|
| [fake( schema )](/docs/javascript-api/mokapi-faker/fake.md) | Creates a fake based on the given schema.  |

## mokapi/kafka

| Functions                                                           | Description                               |
|---------------------------------------------------------------------|-------------------------------------------|
| [produce( \[args\] )](/docs/javascript-api/mokapi-kafka/produce.md) | Sends a single message to a Kafka topic.  |

## mokapi/mustache

| Functions                                                                   | Description                                              |
|-----------------------------------------------------------------------------|----------------------------------------------------------|
| [render( template, scope )](/docs/javascript-api/mokapi-mustache/render.md) | Renders the given mustache template with the given data  |

## mokapi/yaml

| Functions                                                           | Description                                                                                |
|---------------------------------------------------------------------|--------------------------------------------------------------------------------------------|
| [parse( text )](/docs/javascript-api/mokapi-yaml/parse.md)          | Parses a YAML string, constructing the JavaScript value or object described by the string. |
| [stringify( value )](/docs/javascript-api/mokapi-yaml/stringify.md) | Converts a JavaScript value to a YAML string.                                              |





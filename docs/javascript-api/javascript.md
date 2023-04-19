# Mokapi's Javascript API

The list of Mokapi's Javascript modules usable to extend Mokapi's behavior.

## mokapi

| Functions                   | Description                                        |
|-----------------------------|----------------------------------------------------|
| open(filePath)              | Opens a file and read its content                  |
| env(name)                   | Returns the value of the environment variable      |
| on(event, function, args)   | Registers an event handler for the specified event |
| cron(expr, function, args)  | Creates a new cron job                             |
| every(expr, function, args) | Creates a new scheduled job                        |
| sleep(milliseconds)         | Suspends execution for the specified duration.     |

## kafka

| Functions         | Description                   |
|-------------------|-------------------------------|
| produce([params]) | Produces a new Kafka message  |

```javascript
import { produce } from 'kafka'
export default function() {
  var msg = produce({topic: 'topic', value: 'value', key: 'key', partition: 2})
  console.log(`key=${msg.key}, value=${msg.value}`)
}
```

## faker

| Functions    | Description                                                    |
|--------------|----------------------------------------------------------------|
| fake(schema) | Generates random data depending on given OpenAPI schema object |

```javascript
import {fake} from 'faker'
export default function() {
  var s = fake({type: 'string'})
  console.log(s)
}
```

## mustache

| Functions              | Description                                             |
|------------------------|---------------------------------------------------------|
| render(template, data) | Renders the given mustache template with the given data |

## yaml

| Functions   | Description                                                                                    |
|-------------|------------------------------------------------------------------------------------------------|
| parse(yaml) | Parses a YAML string, constructing the JavaScript value or object<br/> described by the string |

## http

| Functions                | Description                    |
|--------------------------|--------------------------------|
| get(url, args)           | Issues an HTTP GET request     |
| post(url, body, args)    | Issues an HTTP POST request    |
| put(url, body, args)     | Issues an HTTP PUT request     |
| head(url, args)          | Issues an HTTP HEAD request    |
| patch(url, body, args)   | Issues an HTTP PATCH request   |
| del(url, body, args)     | Issues an HTTP DELETE request  |
| options(url, body, args) | Issues an HTTP OPTIONS request |





---
title: Retry
description: Retry settings for producing, publishing and sending messages before a topic or channel is ready.
---
# Retry

Scripts can run before Mokapi has fully set up a Kafka topic, MQTT topic or WebSocket channel.
`Retry` controls how often Mokapi retries in that case. It is used by the `retry` option of:

- Kafka [`produce()` and `produceAsync()`](/docs/javascript-api/mokapi/app/kafka-produce.md#kafkaproduceargs)
- MQTT [`publish()` and `publishAsync()`](/docs/javascript-api/mokapi/app/mqtt-router.md#publish)
- WebSocket [`send()` and `sendAsync()`](/docs/javascript-api/mokapi/app/websocket-router.md#send)

Mokapi waits with exponential backoff. The wait time before each retry is
`min(initialRetryTime * factor^attempt, maxRetryTime)`. When all retries are exhausted, the
operation fails with an error.

| Name                        | Type             | Default   | Description                                                      |
|-----------------------------|------------------|-----------|------------------------------------------------------------------|
| retries (optional)          | number           | `5`       | Maximum number of retries. `0` disables retries.                 |
| initialRetryTime (optional) | string \| number | `200ms`   | Wait time before the first retry                                 |
| factor (optional)           | number           | `4`       | Multiplier applied to the wait time after each retry             |
| maxRetryTime (optional)     | string \| number | `30000ms` | Maximum wait time                                                |

Durations accept milliseconds as a number, or a string such as `"500ms"`, `"30s"` or `"2h45m"`.
Valid units are `ns`, `us` (or `µs`), `ms`, `s`, `m` and `h`.

With the default values, Mokapi waits 200ms, 800ms, 3200ms and so on, up to `maxRetryTime`.

## Example

```javascript
import { app } from 'mokapi'

export default function() {
    app.kafka().produce('order-placed', { data: { id: 1 } }, {
        retry: { retries: 10, initialRetryTime: '500ms', maxRetryTime: '60s', factor: 2 }
    })
}
```

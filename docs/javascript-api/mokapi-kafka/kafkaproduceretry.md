---
title: KafkaProduceRetry
description: The retry option can be used to customize the configuration of the retry mechanism.
---
# KafkaProduceRetry

The retry option can be used to customize the configuration of the retry mechanism.

| Name             | Type           | Description                                                                                                                                     |
|------------------|----------------|-------------------------------------------------------------------------------------------------------------------------------------------------|
| maxRetryTime     | string, number | Maximum wait time for a retry. Default 30s. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".                                     |
| initialRetryTime | string, number | Initial value used to calculate the wait time. Default 200ms. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".                   |
| factor           | number         | Factor for increasing the wait time for next retry. <br />1st retry: 200ms<br />2nd retry: 4 * 200ms = 800ms<br />3th retry: 4 * 800ms = 3200ms |
| retries          | number         | Max number of retries                                                                                                                           |

## Example

```javascript tab=kafka.js
import { produce } from 'mokapi/kafka'

export default function() {
    produce({
        topic: 'foo',
        messages: [
            {
                key: 'foo-1',
                data: { foo: 'bar-1' }
            },
        ],
        retry: { retries: 2 }
    })
}
```
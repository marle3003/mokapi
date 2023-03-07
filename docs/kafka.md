# AsyncApi

With [AsyncApi](http://asyncapi.com) you can mock your [Apache Kafka](https://kafka.apache.org/) streaming API. 

## Example

```yaml
asyncapi: '2.0.0'
info:
  title: Kafka Testserver
  description: A kafka test server
  version: '1.0'
servers:
  broker:
    url: demo:9092
    protocol: kafka
    bindings:
      kafka:
        group.initial.rebalance.delay.ms: 100
        log.retention.minutes: 5
        log.segment.bytes: 1048576 # 1 MB
channels:
  /message:
    subscribe:
      message:
        $ref: '#/components/messages/message'
    publish:
      message:
         $ref: '#/components/messages/message'
    bindings:
      kafka:
        partitions: 5

components:
  messages:
    message:
      contenttype: application/json
      payload:
        $ref: '#/components/schemas/Order'
      bindings:
        kafka:
          key:
            type: string
  schemas:
    Order:
      type: object
      properties:
        id:
          type: number
          x-faker: uuid
        price:
          type: number
          x-faker: price:1,1000
        shipTo:
          type: object
          properties:
            name:
              type: string
              x-faker: name
            address:
              type: string
              x-faker: street
            city:
              type: string
              x-faker: city
            zip:
              type: string
              x-faker: zip
```

## Produce messages
With Mokapi Actions you can produce scheduled messages to your Kafka channels.

```yaml
mokapi: 1.0
workflows:
  - name: produce kafka messages
    on:
      schedule:
        every: 5s
    steps:
      - uses: kafka-producer
        with:
          topic: message
```


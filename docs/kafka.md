#AsyncApi

With [AsyncApi](http://asyncapi.com) you can mock your [Apacahe Kafka](https://kafka.apache.org/) streaming API. 

##Example

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

## Kafka Server Bindings

### group.initial.rebalance.delay.ms
The amount of time Mokapi will wait for more consumers to join a new group before performing the first rebalance. Default value: *3000*

### log.retention.hours
The numbers of hours to keep a log before deleting it. Default value *24*

### log.retention.minutes
The numbers of minutes to keep a log before deleting it. If not set, the value in *log.retention.hours* is used

### log.retention.ms
The numbers of milliseconds to keep a log before deleting it. If not set, the value in *log.retention.minutes* is used

### log.retention.bytes
The maximum size of the log before deleting it. Default value is *-1*

### log.retention.check.interval.ms
The frequency in milliseconds that the log cleaner checks whether any log is eligible for deletion. Default value is *300000* (5 minutes)

### log.segment.bytes
The maximum size of a single log. Default value is *10485760* (10 megabytes)

### log.segment.delete.delay.ms
The amount of time to wait before deleting a log. Default value is *60000* (1 minute)

### log.cleaner.backoff.ms
The amount of time to sleep when there are no logs to clean. Default value is *15000* (15 seconds)

## Kafka Channel Bindings

### partitions
Sets the number of partition for the channel. Default value is *1*
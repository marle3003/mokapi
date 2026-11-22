---
title: Kafka Configuration
description: Configure Kafka servers, topics, schema validation and message retention in Mokapi with AsyncAPI Kafka bindings.
---
# Kafka Configuration

You configure Mokapi's Kafka behavior with the Kafka bindings of your AsyncAPI specification. Server
bindings control the broker (retention, cleanup, consumer groups), channel bindings control a single
topic (partitions, retention, schema validation), and message and operation bindings describe keys,
schema IDs and client IDs.

```yaml tab=3.0
asyncapi: 3.0.0
info:
  title: Orders Service
  version: 1.0.0

servers:
  mokapi:
    host: localhost:9092
    protocol: kafka
    bindings:
      kafka:
        log.retention.minutes: 30
        group.initial.rebalance.delay.ms: 500

channels:
  orders:
    address: orders
    bindings:
      kafka:
        partitions: 3
        retention.ms: 600000
        confluent.value.schema.validation: true
    messages:
      order:
        payload:
          type: object
          properties:
            id: { type: string }
```

```yaml tab=2.6
asyncapi: 2.6.0
info:
  title: Orders Service
  version: 1.0.0

servers:
  mokapi:
    url: localhost:9092
    protocol: kafka
    bindings:
      kafka:
        log.retention.minutes: 30
        group.initial.rebalance.delay.ms: 500

channels:
  orders:
    bindings:
      kafka:
        partitions: 3
        retention.ms: 600000
        confluent.value.schema.validation: true
    publish:
      message:
        payload:
          type: object
          properties:
            id: { type: string }
```

Keys that Mokapi doesn't know are ignored. You can also change bindings without editing the
specification, see [Patching](/docs/configuration/patching.md).

## Server Bindings

Set server bindings under `servers.<name>.bindings.kafka`. They apply to the Kafka server they are
defined on.

| Key                                | Default        | Description                                                                                                  |
|------------------------------------|----------------|--------------------------------------------------------------------------------------------------------------|
| `log.retention.ms`                 | 604800000      | How long Mokapi keeps a closed log segment before it deletes it. Default is 7 days.                          |
| `log.retention.minutes`            |                | Same as `log.retention.ms` in minutes. Used if `log.retention.ms` is not set.                                |
| `log.retention.hours`              |                | Same as `log.retention.ms` in hours. Used if neither `.ms` nor `.minutes` is set.                            |
| `log.retention.bytes`              | -1 (no limit)  | Maximum size of a partition. If a partition reaches this size, Mokapi deletes all of its closed segments.       |
| `log.roll.ms`                      | 604800000      | How long a segment stays open before Mokapi closes it. Default is 7 days.                                    |
| `log.roll.minutes`                 |                | Same as `log.roll.ms` in minutes. Used if `log.roll.ms` is not set.                                          |
| `log.roll.hours`                   |                | Same as `log.roll.ms` in hours. Used if neither `.ms` nor `.minutes` is set.                                 |
| `log.retention.check.interval.ms`  | 300000         | How often the log cleaner runs. Default is 5 minutes.                                                        |
| `group.initial.rebalance.delay.ms` | 3000           | How long Mokapi waits for more consumers to join a consumer group before it completes a rebalance.           |
| `group.min.session.timeout.ms`     | 6000           | How often Mokapi checks whether a member of a consumer group missed its heartbeat.                           |

A value of `0` or a missing key means the default. Mokapi doesn't treat `-1` as "no limit" for the time
based keys, only `log.retention.bytes` is unlimited by default.

The keys `log.segment.bytes`, `log.segment.delete.delay.ms`, `schemaRegistryUrl` and
`schemaRegistryVendor` are accepted, but they have no effect on Mokapi's behavior.

### How retention works

Mokapi keeps messages in memory in segments. The log cleaner runs every
`log.retention.check.interval.ms` and does two things for each partition:

1. It closes the open segment when it is older than `log.roll.ms` (or `segment.ms` of the topic).
2. It deletes a closed segment when it was closed longer ago than `log.retention.ms` (or `retention.ms`
   of the topic).

A message is therefore available for up to the roll time plus the retention time. With the defaults
that is up to 14 days.

Mokapi deletes only closed segments, also for the size limits `log.retention.bytes` and `retention.bytes`.
A size limit therefore has no effect until the open segment has been closed, which takes 7 days by default.
To let messages expire quickly in a test, lower the roll time and the retention:

```yaml
servers:
  mokapi:
    host: localhost:9092
    protocol: kafka
    bindings:
      kafka:
        log.roll.ms: 60000
        log.retention.ms: 60000
        log.retention.check.interval.ms: 10000
```

## Channel Bindings

Set channel bindings under `channels.<name>.bindings.kafka`. They apply to the topic of the channel and
override the server bindings of the same meaning. You can also nest the keys under `topicConfiguration`.

| Key                                 | Default       | Description                                                                                              |
|-------------------------------------|---------------|----------------------------------------------------------------------------------------------------------|
| `partitions`                        | 1             | Number of partitions of the topic. If you change it in a reloaded specification, Mokapi adds or removes partitions. |
| `retention.ms`                      | server value  | How long Mokapi keeps a closed segment of this topic. Overrides `log.retention.*`.                       |
| `retention.bytes`                   | server value  | Maximum size of a partition of this topic. Overrides `log.retention.bytes`.                              |
| `segment.ms`                        | server value  | How long a segment of this topic stays open. Overrides `log.roll.*`.                                     |
| `confluent.value.schema.validation` | true          | Validates the message value against the payload schema of the message.                                   |
| `confluent.key.schema.validation`   | true          | Validates the message key against the `key` schema of the message binding.                               |

The key `segment.bytes` is accepted but has no effect.

```yaml
channels:
  orders:
    address: orders
    bindings:
      kafka:
        partitions: 3
        retention.ms: 600000
        segment.ms: 300000
        retention.bytes: 1048576
```

### Schema validation

Both validation keys default to `true`, which means Mokapi validates. Set a key to `false` to skip
that validation. Mokapi then accepts any value or key for that part of the record.

When a produced record fails validation, Mokapi does not store it. The produce response contains the
error code `INVALID_RECORD` (87) and an error message that describes the schema violation. This
applies to the whole batch, so Mokapi doesn't store any record of a batch that contains an invalid
record. If a channel defines several messages, a record is valid when it matches at least one of them.

```yaml
channels:
  events:
    address: events
    bindings:
      kafka:
        confluent.value.schema.validation: false
        confluent.key.schema.validation: false
```

Header validation is not affected by these keys. If a message defines `headers`, Mokapi always
validates them.

```box=warning
With AsyncAPI 2.x, Mokapi ignores `confluent.key.schema.validation` and doesn't validate message keys.
Use AsyncAPI 3.0 if you need key validation.
```

## Message Bindings

Set message bindings under `bindings.kafka` of a message.

| Key                       | Default | Description                                                                                                  |
|---------------------------|---------|--------------------------------------------------------------------------------------------------------------|
| `key`                     |         | Schema of the message key. Mokapi validates the key against it.                    |
| `schemaIdLocation`        |         | Set to `payload` if the payload starts with a schema ID. Mokapi reads the ID and shows it in the message log. |
| `schemaIdPayloadEncoding` | 4       | Number of bytes of the schema ID. `confluent` is the same as `4`. Used if `schemaIdLocation` is `payload`.   |

With `schemaIdLocation: payload` Mokapi expects the Confluent wire format: one magic byte followed by the
schema ID as a big-endian integer. A record that is too short is rejected as invalid.

```yaml
components:
  messages:
    order:
      contentType: application/json
      bindings:
        kafka:
          key:
            type: string
      payload:
        type: object
        properties:
          id: { type: string }
```

With AsyncAPI 2.x, Mokapi supports only `key`.

## Operation Bindings

Set operation bindings under `bindings.kafka` of an operation.

| Key        | Description                                                                                                                       |
|------------|-----------------------------------------------------------------------------------------------------------------------------------|
| `clientId` | JSON Schema for the client ID. Mokapi checks the client ID of producers against `send` operations and of consumers against `receive` operations. |
| `groupId`  | JSON Schema for the consumer group ID. Mokapi checks the group ID of consumers against `receive` operations.                       |

```yaml
operations:
  sendOrder:
    action: send
    channel:
      $ref: '#/channels/orders'
    bindings:
      kafka:
        clientId:
          type: string
          pattern: '^order-service-.*$'
```

A produce request from a client whose ID doesn't match returns the error code `UNKNOWN_SERVER_ERROR` and
a message that contains the client ID and the schema violation. A consumer that commits or fetches offsets
with a non-matching client ID gets the same error code, and one with a non-matching group ID gets
`INVALID_GROUP_ID`. If a topic has several `receive` operations, it is enough that the consumer matches one
of them.

---
title: KafkaEventMessage
description: KafkaEventMessage is an object used by KafkaEventHandler
---
# KafkaEventMessage

KafkaMessage is an object used by [KafkaEventHandler](/docs/javascript-api/mokapi/eventhandler/kafkaeventhandler.md)
that contains Kafka-specific message data.

| Name    | Type   | Description                                    |
|---------|--------|------------------------------------------------|
| offset  | number | Kafka offset of the kafka message (read-only). |
| key     | string | Kafka message key                              |
| value   | string | Kafka message value                            |
| headers | object | Kafka message headers                          |


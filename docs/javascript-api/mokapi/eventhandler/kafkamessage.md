---
title: KafkaMessage
description: KafkaMessage is an object used by KafkaEventHandler
---
# KafkaMessage

KafkaMessage is an object used by [KafkaEventHandler](/docs/javascript-api/mokapi/eventhandler/kafkaeventhandler.md)
that contains Kafka-specific message data.

| Name    | Type   | Description                                                   |
|---------|--------|---------------------------------------------------------------|
| offset  | number | Kafka partition where the message was written to (read-only). |
| key     | string | Kafka message key                                             |
| value   | string | Kafka message value                                           |
| headers | object | Kafka message headers                                         |


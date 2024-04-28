---
title: Kafka Configuration
description: Mokapi provides you some configuration parameters to set up a Kafka topic
---
# Kafka Configuration

Mokapi provides you some configuration parameters to set up a Kafka topic

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

## Kafka Channel Bindings

### partitions
Sets the number of partition for the channel. Default value is *1*

### retention.bytes
The maximum size of a segment before deleting it. Default value is *-1*

### retention.ms
The numbers of milliseconds to keep a segment before deleting it.

### confluent.value.schema.validation
Skip validation of Kafka messages. Default value is *true*
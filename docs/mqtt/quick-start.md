---
title: "MQTT API Mocking: Publish Sensor Readings to AsyncAPI with Mokapi"
description: Learn how to set up Mokapi to publish sensor readings to an AsyncAPI-defined MQTT broker, including message publishing, validation, and monitoring.
---

# Publishing Sensor Readings to AsyncAPI MQTT

## Overview

This guide walks through a complete sensor workflow using Mokapi as a mock MQTT broker. By the end,
you'll have:

1. Defined an MQTT broker and a `sensors/temperature` topic using AsyncAPI.
2. Started Mokapi to simulate the broker and validate messages against the schema.
3. Published a sensor reading to the topic.
4. Subscribed to that topic to verify the message arrived.
5. Seen what happens when a message doesn't match the schema

## Define the AsyncAPI Specification

Create an AsyncAPI file (`asyncapi.yaml`) describing the MQTT broker and the `sensors/temperature` topic.

```yaml
asyncapi: 3.0.0
info:
  title: Sensor Service
  version: 1.0.0
  description: This AsyncAPI document defines the MQTT broker for publishing and subscribing to sensor readings.

servers:
  mokapi:
    host: 'localhost:1883'
    protocol: mqtt
    description: Mock MQTT broker provided by Mokapi.

channels:
  temperatureReading:
    address: 'sensors/temperature'
    description: Topic for temperature sensor readings.
    messages:
      TemperatureEvent:
        $ref: '#/components/messages/Temperature'

operations:
  publishTemperature:
    action: send
    summary: Publish a temperature reading to the topic.
    channel:
      $ref: '#/channels/temperatureReading'
    messages:
      - $ref: '#/channels/temperatureReading/messages/TemperatureEvent'
  subscribeTemperature:
    action: receive
    summary: Subscribe to temperature readings from the topic.
    channel:
      $ref: '#/channels/temperatureReading'
    messages:
      - $ref: '#/channels/temperatureReading/messages/TemperatureEvent'

components:
  messages:
    Temperature:
      name: TemperatureEvent
      title: A temperature reading
      payload:
        type: object
        properties:
          sensorId:
            type: string
            description: Unique identifier for the sensor.
          value:
            type: number
            format: float
            description: The measured temperature.
          unit:
            type: string
            enum: [celsius, fahrenheit]
            description: Unit of measurement.
        required:
          - sensorId
          - value
```

The `required` field ensures every reading includes at least a `sensorId` and a `value`. Mokapi enforces
this automatically once the broker is running.

## Start Mokapi

Start Mokapi with the AsyncAPI file to simulate the MQTT broker:

```bash
mokapi asyncapi.yaml
```

Mokapi's log output will confirm that an MQTT broker is running at `localhost:1883`, ready to accept
connections from publishers and subscribers exactly like a real broker.

## Publish a Sensor Reading

Use an MQTT client to publish a message to the `sensors/temperature` topic. Here's an example using
MQTTnet for .NET:

```csharp
using MQTTnet;
using MQTTnet.Client;
using System.Text.Json;

var factory = new MqttFactory();
using var client = factory.CreateMqttClient();

var options = new MqttClientOptionsBuilder()
    .WithTcpServer("localhost", 1883)
    .Build();

await client.ConnectAsync(options);

var reading = new
{
    sensorId = "sensor-001",
    value = 21.5,
    unit = "celsius"
};

var message = new MqttApplicationMessageBuilder()
    .WithTopic("sensors/temperature")
    .WithPayload(JsonSerializer.Serialize(reading))
    .Build();

await client.PublishAsync(message);

Console.WriteLine("Reading published successfully!");
```

Mokapi receives the message, validates it against the `TemperatureEvent` schema, and makes it available
to anyone subscribed to `sensors/temperature`.

## Subscribe to Sensor Readings

To verify the message arrived, subscribe to the `sensors/temperature` topic:

```csharp
using MQTTnet;
using MQTTnet.Client;

var factory = new MqttFactory();
using var client = factory.CreateMqttClient();

var options = new MqttClientOptionsBuilder()
    .WithTcpServer("localhost", 1883)
    .Build();

client.ApplicationMessageReceivedAsync += e =>
{
    var payload = e.ApplicationMessage.ConvertPayloadToString();
    Console.WriteLine($"Message: {payload}");
    return Task.CompletedTask;
};

await client.ConnectAsync(options);
await client.SubscribeAsync("sensors/temperature");

Console.WriteLine("Subscribed. Waiting for messages...");
await Task.Delay(Timeout.Infinite);
```

Running this alongside the publisher should print the sensor reading you just sent.

## Monitor Messages in the Dashboard

Instead of writing a subscriber just to check that a message arrived, you can use Mokapi's web dashboard
to inspect MQTT traffic directly. This is often the fastest way to debug message flow during development.

1. Open the dashboard, by default at http://localhost:8080.
2. Go to the MQTT tab.
3. Select the sensors/temperature topic to see all messages that have been published to it, including
their payload and QoS level.

## Schema Validation

Every message published to `sensors/temperature` is validated against the TemperatureEvent schema
defined in `asyncapi.yaml`. If a message doesn't conform to the schema:

- Mokapi rejects the message.
- Mokapi logs a validation error describing what doesn't match.

## Test an Invalid Message

Publish a message that's missing the required `value` field:

```csharp
var message = new MqttApplicationMessageBuilder()
    .WithTopic("sensors/temperature")
    .WithPayload(@"{ ""sensorId"": ""sensor-001"", ""unit"": ""celsius"" }")
    .Build();

await client.PublishAsync(message);
```

Mokapi rejects this message because `value` is required by the schema. Check Mokapi's logs, you'll see
a validation error identifying the missing field. The message won't appear in the dashboard either,
since it was never accepted.

## Next Steps

- [Mocking MQTT with AsyncAPI](/docs/mqtt/overview.md): Learn more about how Mokapi simulates MQTT.
- [JavaScript API](/docs/javascript-api/overview.md): Add dynamic behavior, generate test data, or simulate errors with Mokapi Scripts.
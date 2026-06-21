---
title: Mock MQTT Topics for Testing and Development
description: Mock MQTT brokers and topics using AsyncAPI specification for seamless testing and development of IoT and pub/sub systems
---

# Mocking MQTT with AsyncAPI

Mokapi turns an AsyncAPI specification into a working MQTT mock broker. Rather than simulating real
device hardware or network conditions, Mokapi focuses on message flow: topics, payloads, and the
publish/subscribe contract between your application and the devices or services it talks to. This
removes the need for physical devices or a real broker during development and testing, while still
ensuring your publishers and subscribers strictly adhere to their message contracts.

```yaml
asyncapi: 3.0.0
info:
  title: Sensor Service
  version: 1.0.0

servers:
  mokapi:
    host: 'localhost:1883'
    protocol: mqtt

channels:
  temperatureReading:
    address: 'sensors/temperature'
    messages:
      temperatureEvent:
        payload:
          type: object
          properties:
            sensorId: { type: string }
            value: { type: number }
            unit: { type: string, enum: [celsius, fahrenheit] }
```

``` box=tip title=Recommendation
Ready to dive in? Head over to the MQTT [Quick Start Guide](/docs/mqtt/quick-start.md) and run your first
MQTT mock in seconds.
```

## Why Use Mokapi for MQTT?

Testing IoT and pub/sub systems against real devices or a real broker is slow, hard to reproduce, and
often requires hardware you don't have on hand. Mokapi provides a lightweight, stable MQTT broker built
specifically for local development and CI/CD pipelines.

**Zero infrastructure overhead**: No physical devices, no real broker to install or configure. Mokapi runs as a single binary or container and works out of the box.

**Contract-first validation**: Every message is validated against your schema in real time, catching malformed payloads before they reach your codebase.

**Reproducible test suites**: Mokapi runs entirely in memory, so every test run starts from a clean slate with no retained messages or leftover subscriptions from previous runs.

## Supported Standards

Mokapi integrates with the existing MQTT ecosystem and supports modern industry standards:

**AsyncAPI specifications**: Full support for both version 2.x and version 3.0.

**Schema formats**: Built-in validation for JSON Schema.

**MQTT protocol**: Compatible with standard MQTT clients (Eclipse Paho, MQTT.js, and others) over MQTT 3.1.1 and 5.0.

## Key Features

### Automated Topic Provisioning

Mokapi reads your AsyncAPI definition and automatically provisions the topics it describes. No manual
setup required, topics are ready to publish and subscribe to as soon as Mokapi starts.

### Dynamic Data Generation

If a subscriber is waiting for messages but no publisher is currently active, Mokapi can generate
realistic mock messages based on your schema. This lets you test subscribers in complete isolation,
without needing a running publisher or a real device.

### Error and Latency Simulation

Use [Mokapi Scripts](/docs/javascript-api/overview.md) to simulate conditions that are difficult to
trigger with real devices:

- Inject network latency or jitter
- Simulate dropped connections or delayed delivery
- Build stateful mock behavior with JavaScript, for example a sensor that gradually drifts out of range

## Next Steps

- [Quick Start Guide](/docs/mqtt/quick-start.md): Run Mokapi and load your first AsyncAPI file.
- [Mokapi CLI:](/docs/configuration/static/cli-usage.md): Command-line options and runtime configuration.
- [JavaScript API](/docs/javascript-api/overview.md): Write scripts to control mock behavior dynamically.
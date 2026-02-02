---
title: Mock Kafka Topics for Testing and Development
description: Mock Kafka Topics Using AsyncAPI Specification for Seamless Testing and Development
---
# Mocking Kafka with AsyncAPI

Mokapi transforms your AsyncAPI Specification into a functional Kafka mock server.
Mokapi focuses on mocking Kafka topics and message flow rather than simulating Kafka’s internal cluster mechanics.
It eliminates the overhead of managing local clusters while ensuring your producers and
consumers strictly adhere to their API contracts.

```yaml tab=3.0
asyncapi: 3.0.0
info:
  title: User Service
  version: 1.0.0

channels:
  userSignedUp:
    address: 'users.signedup'
    messages:
      userEvent:
        payload:
          type: object
          properties:
            id: { type: string, format: uuid }
            email: { type: string, format: email }
```

```yaml tab=2.6
asyncapi: 2.6.0
info:
  title: User Service
  version: 1.0.0

channels:
  users.signedup:
    publish:
      message:
        name: userEvent
        payload:
          type: object
          properties:
            id: { type: string, format: uuid }
            email: { type: string, format: email }
```

```box=tip
Ready to dive in? Head over to the Kafka [Quick Start Guide](/docs/guides/kafka/quick-start.md) and run your first
Kafka mock in seconds.
```

## Why Use Mokapi for Kafka?

Testing event-driven architectures with a real Kafka instance is often heavy and non-deterministic. Mokapi provides a
lightweight, stable broker designed specifically for local development and CI/CD pipelines.

- <p><strong>Zero Infrastructure Overhead:</strong><br/>Skip the complexity of Kafka cluster setup (ZooKeeper or KRaft). Mokapi is a single binary/container that works out of the box.
- <p><strong>Contract-First Validation:</strong><br/>Every message is validated against your schemas in real-time. Catch breaking changes before they even reach your codebase.
- <p><strong>Reproducible Test Suites:</strong><br/>Operating entirely in-memory, Mokapi ensures every test run starts with a clean slate—no leftover offsets or orphaned topics.

## Supported Standards

Mokapi integrates seamlessly into the existing ecosystem, supporting modern industry standards:

- <p><strong>AsyncAPI Specifications:</strong><br/>Full support for both Version 2.x and Version 3.0.
- <p><strong>Schema Formats:</strong><br/>Built-in validation for JSON Schema and Avro.
- <p><strong>Kafka Protocol:</strong><br/>Compatible with standard Kafka clients (Java, Go, Python, .NET, etc.).

## Key Features

### 1. Automated Topic Provisioning

Based on your AsyncAPI definition, Mokapi automatically initializes the required channels (topics). No manual admin commands or startup scripts required.

### 2. Dynamic Data Generation

If a consumer requests data but no producer is active, with Mokapi you can generate realistic mock data based on your schema. 
This allows you to test consumers in complete isolation.

### 3. Error & Latency Simulation

Use Mokapi Scripts to simulate edge cases that are difficult to trigger in a real environment:

- Inject network latency or jitter.
- Simulate broker-specific Kafka Error Codes.
- Create stateful mock behavior using JavaScript.

## Architectural Design

To ensure speed and determinism, Mokapi simulates Kafka's application behavior rather than its cluster administration:

- <p><strong>Single Stable Broker:</strong><br/>Focuses on message flow rather than 
  leader election, partition replication, or broker coordination.
- <p><strong>Ephemeral by Design:</strong><br/>Data is kept in-memory to provide 
  lightning-fast feedback loops during development.
- <p><strong>Deterministic Broker Address Resolution:</strong><br/>
  Mokapi resolves the advertised broker address based on the listener port.
  If multiple AsyncAPI servers share the same port, the first matching server
  definition is used. AsyncAPI servers are treated as environment-specific
  configurations rather than simultaneously addressable brokers.</p>

Kafka clients do not transmit the DNS name used to establish a connection.
When multiple servers share the same listener port, it is therefore not
possible to determine which server was used at runtime. Mokapi follows
Kafka’s networking model and applies a deterministic resolution strategy
instead of guessing.

``` box=tip title=Recommendation
If you define multiple AsyncAPI servers, use different ports for each server
or use Mokapi’s [patching](/docs/configuration/patching.md) mechanism.
```

## Next Steps

- [Quick Start Guide](/docs/guides/kafka/quick-start.md): Learn how to run Mokapi and load your first AsyncAPI file.
- [Mokapi CLI:](/docs/configuration/static/cli-usage.md): Detailed command-line options and runtime configuration.
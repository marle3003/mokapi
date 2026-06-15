---
title: Mock Kafka Topics for Testing and Development
description: Mock Kafka Topics Using AsyncAPI Specification for Seamless Testing and Development
---
# Mocking Kafka with AsyncAPI

Mokapi turns an AsyncAPI specification into a working Kafka mock server. Rather than simulating Kafka's
cluster internals, Mokapi focuses on message flow: topics, schemas, and the producer/consumer contract.
This removes the overhead of running a local cluster while still ensuring your producers and consumers
strictly adhere to their API contracts.

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
Ready to dive in? Head over to the Kafka [Quick Start Guide](/docs/kafka/quick-start.md) and run your first
Kafka mock in seconds.
```

## Why Use Mokapi for Kafka?

Testing event-driven architectures with a real Kafka instance is often heavy and non-deterministic. Mokapi provides a
lightweight, stable broker designed specifically for local development and CI/CD pipelines.

- **Zero infrastructure overhead**:  
  Skip cluster setup entirely, no ZooKeeper, no KRaft. Mokapi runs as a single binary or container and works out of the box.
- **Contract-first validation**:  
  Every message is validated against your schema in real time, catching breaking changes before they reach your codebase.
- **Reproducible test suites**:  
  Mokapi runs entirely in memory, so every test run starts from a clean slate with no leftover offsets or orphaned topics.

## Supported Standards

Mokapi integrates seamlessly into the existing ecosystem, supporting modern industry standards:

- **AsyncAPI Specifications**:  
  Full support for both Version 2.x and Version 3.x.
- **Schema Formats**:  
  Built-in validation for JSON Schema and Avro.
- **Kafka Protocol**:  
  Compatible with standard Kafka clients (Java, Go, Python, .NET, etc.).

## Key Features

### Automated Topic Provisioning

Mokapi reads your AsyncAPI definition and automatically provisions the channels (topics) it describes.
No manual admin commands or startup scripts required, topics are ready as soon as Mokapi starts.

### Dynamic Data Generation

If a consumer requests data but no producer is currently active, Mokapi can generate realistic mock
messages based on your schema. This lets you test consumers in complete isolation, without needing a
running producer.

### Error and Latency Simulation

Use [Mokapi Scripts](/docs/javascript-api/overview.md) to simulate conditions that are difficult to
trigger in a real environment:

- Inject network latency or jitter
- Simulate broker-specific Kafka error codes
- Build stateful mock behavior with JavaScript

## Architectural Design

Mokapi is designed for speed and determinism by simulating Kafka's application-level behavior rather
than its cluster administration.

- **Single stable broker**:  
  Mokapi focuses on message flow rather than partition replication, or broker coordination.
- **Ephemeral by design**:  
  All data is kept in memory, giving you fast, predictable feedback loops during development.
- **Deterministic broker address resolution**:  
  Mokapi resolves the advertised broker address based on the listener port. If multiple AsyncAPI servers
  share the same port, the first matching server definition is used. AsyncAPI servers are treated as
  environment-specific configurations rather than simultaneously addressable brokers.

Kafka clients don't transmit the DNS name used to establish a connection. When multiple servers share
the same listener port, it's therefore not possible to determine at runtime which server was used.
Mokapi follows Kafka's networking model and applies a deterministic resolution strategy instead of guessing.

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

- [Quick Start Guide](/docs/kafka/quick-start.md): Run Mokapi and load your first AsyncAPI file.
- [Mokapi CLI:](/docs/configuration/static/cli-usage.md): Command-line options and runtime configuration.
- [JavaScript API](/docs/javascript-api/overview.md): Write scripts to control mock behavior dynamically.
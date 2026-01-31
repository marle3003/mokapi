# AsyncAPI 3.0 – Mokapi Example

This example demonstrates how to mock a Kafka topic using an **AsyncAPI 3.0** specification and **Mokapi**, 
and how to produce and consume messages locally.

## Run the example

### 1. Install dependencies:

```bash
npm install
```

### 2. Start Mokapi with the AsyncAPI spec and producer script

```bash
mokapi ./api.yaml ./producer.js &
```

### 3. Run the consumer

```bash
node consumer.js
```
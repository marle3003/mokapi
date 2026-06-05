---
title: How to Test Kafka Workflows End-to-End Without a Real Broker
description: Mock a Kafka broker with Mokapi and test your Node.js event-driven service end-to-end. No Docker, no Zookeeper.
subtitle: "Testing event-driven services is hard. You either spin up a real Kafka cluster and deal with the infrastructure overhead, or you stub out your Kafka client and end up testing nothing real. This guide shows a third way: mocking the broker with Mokapi so your actual service code runs against a lightweight, spec-validated fake. No Docker, no Zookeeper, and no changes to your application code."
icon: bi-lightning
tech: kafka
---

# Testing Kafka Workflows Without Spinning Up a Real Cluster

Here's a situation I kept running into when building Mokapi. Developers would come to me with the same problem: they had an event-driven backend that worked great in production, but testing it was a nightmare.

A message comes in on one Kafka topic. The service processes it. An event goes out on another. Simple flow. But how do you test that reliably?

Do you spin up a real Kafka cluster locally? Set up Docker Compose with Zookeeper and a broker and hope nothing goes wrong in CI? Stub out your Kafka client library so consumer.run() never actually runs, and call it a unit test?

None of those feel great. The real cluster is heavy and fragile in CI. And stubbing the Kafka client is the sneaky bad one: it feels like a test, but you're not testing anything real. You're just confirming that your mock behaves the way you told it to. Things like message serialization, offset handling, schema validation, none of that gets exercised. You can have a green test suite and still ship broken event handling. I got tired of watching teams skip this kind of testing entirely because the setup cost was too high. That frustration is a big part of why I built Mokapi.

Let me show you how I'd test a Kafka workflow using Mokapi today.

---

## The Stack: Mokapi & Playwright

Mokapi lets you define your Kafka topics using AsyncAPI specs and then interact with those topics over a simple
REST API. No real broker. No Zookeeper. Your service connects to localhost:9092 exactly like it would in
production, but Mokapi is the one answering. And because the spec is always enforced, every message gets
validated against your schema automatically, whether it comes from your real service or your test.

In this article we're using Mokapi's HTTP API to produce test messages, which means no Kafka client library
needed in the test code at all. Mokapi actually supports several other ways to produce messages too,
including a JavaScript API and OpenAPI-backed endpoints with custom handlers.

For the test runner I'm using Playwright. Yes, the browser testing tool. But hear me out. Playwright is a
fantastic runner for async workflows even when there's no browser involved. The step model gives you really
clean test output, and the async handling is solid. Jest works too, honestly, but Playwright's reporting
is nicer to read when something goes wrong.

---

## The Scenario We're Testing

Let's make this concrete. The workflow looks like this:

1. A foreign system publishes a command to the topic document.send-command
2. Our Node.js backend picks it up, does some processing, and publishes a result to document.send-event
3. Our test acts as the foreign system: sends the command, waits, checks that the right event came out

Simple enough to follow, but realistic enough to be meaningful. This is the kind of flow you'd find in any document delivery system, order processing pipeline, notification service... you name it.

<img src="/kafka-workflow.png" alt="Sequence diagram of Kafka workflow testing with Mokapi, Playwright, and a backend service" title="Kafka workflow testing with Mokapi and Playwright">


---

## Step 1: Define Your Topics with AsyncAPI

The first thing you need is an AsyncAPI spec. This is what Mokapi uses to know about your topics and message schemas.
Here's a trimmed version of the spec for this example. You can find the [full spec on GitHub](https://github.com/marle3003/mokapi-kafka-workflow/blob/main/mocks/kafka.yaml).

```yaml
asyncapi: 3.0.0
info:
  title: Kafka Mock
  version: 1.0.0
servers:
  mock:
    host: localhost:9092
    protocol: kafka

channels:
  document.send-command:
    description: Sending documents to customers
    messages:
      documentCommand:
        $ref: '#/components/messages/documentCommand'
  document.send-event:
    description: Events when sending documents
    messages:
      documentEvent:
        $ref: '#/components/messages/documentEvent'
```

You define the full message schema in the components section. This is really important because it means your test
and your backend are both working against the same contract, not just some informal agreement between you and
a colleague. When I designed Mokapi around AsyncAPI, this was one of the core ideas: the spec is the source of
truth, not the implementation.

---

## Step 2: The Node.js Backend

The backend is straightforward. It consumes from document.send-command, pretends to send a document, then publishes a result to document.send-event.

```javascript
import { Kafka } from 'kafkajs';

const kafka = new Kafka({
  clientId: 'backend',
  brokers: ['localhost:9092']
});

const consumer = kafka.consumer({ groupId: 'backend-group' });
const producer = kafka.producer();

async function start() {
  await consumer.connect();
  await producer.connect();

  await consumer.subscribe({ topic: 'document.send-command', fromBeginning: true });

  await consumer.run({
    eachMessage: async ({ message }) => {
      const value = JSON.parse(message.value.toString());
      console.log('Received command:', value);

      // Simulate processing time
      await new Promise(res => setTimeout(res, 500));

      const event = {
        documentId: value.documentId,
        status: 'SENT'
      };

      await producer.send({
        topic: 'document.send-event',
        messages: [{ key: value.documentId, value: JSON.stringify(event) }]
      });

      console.log('Published event:', event);
    }
  });
}

start().catch(console.error);
```

Notice that the backend connects to localhost:9092 just like it would in production. It doesn't know or care that
Mokapi is running there instead of a real broker. That transparency was a deliberate design goal. You shouldn't
have to change your application code to make it testable.

## Step 3: Writing the Playwright Test

This is the interesting part. The test does three things:

1. Records the current offset on the output topic so we only look at new messages
2. Produces a command message via Mokapi's REST API
3. Polls the output topic until it finds the matching event, then asserts on it

```typescript
import { test, expect } from '@playwright/test';
import fetch from 'node-fetch';

const MOKAPI_API = 'http://localhost:8080/api/services/kafka/Kafka%20Mock';
const TOPIC_COMMAND = 'document.send-command';
const TOPIC_EVENT = 'document.send-event';

test('Kafka document send workflow', async () => {
  const documentId = 'doc-' + Date.now();
  let startOffset = -1;

  await test.step('Get current offset for events', async () => {
    startOffset = await getPartitionOffset(TOPIC_EVENT, 0);
    console.log('current partition offset is: ' + startOffset);
  });

  await test.step('Produce a message to document.send-command', async () => {
    await produce(TOPIC_COMMAND, {
      key: documentId,
      value: {
        documentId: documentId,
        recipient: 'alice@mokapi.io',
        document: {
          mediaType: 'text/plain',
          fileName: 'test.txt',
          content: 'Hello Alice'
        }
      }
    });
  });

  await test.step('Get messages from document.send-event', async () => {
    let record: any = undefined;
    const timeout = Date.now() + 5000;

    while (Date.now() < timeout && !record) {
      const records: any = await read(TOPIC_EVENT, 0, startOffset);
      record = records.find(x => x.value.documentId === documentId);
      if (record) break;
      startOffset += records.length;
      await new Promise(res => setTimeout(res, 200));
    }

    expect(record, 'record should be found').not.toBeNull();
    expect(record.value.status).toBe('SENT');
  });
});

async function getPartitionOffset(topic, partition) {
  const res = await fetch(`${MOKAPI_API}/topics/${topic}/partitions/${partition}`);
  const data: any = await res.json();
  return data.offset;
}

async function produce(topic: string, record: { key: string; value: any }) {
  const res = await fetch(`${MOKAPI_API}/topics/${topic}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ records: [{ key: record.key, value: record.value }] })
  });
  expect(res.status).toBe(200);
  const data: any = await res.json();
  expect(data.offsets.every(x => !('error' in x))).toBe(true);
}

async function read(topic: string, partition: number, offset: number) {
  const res = await fetch(
    `${MOKAPI_API}/topics/${topic}/partitions/${partition}/offsets?offset=${offset}`
  );
  expect(res.status).toBe(200);
  return await res.json();
}
```

A few things worth pointing out here.
The offset tracking is important. If you don't record where the topic is before you produce your command,
you might accidentally pick up a message from a previous test run. It's a small thing, but it's the kind
of bug that makes you lose an hour.

The polling loop with a timeout is deliberate. You need to give the backend time to consume and process the
command before the event shows up. Five seconds is generous for a test environment, but you can tune it down
once you know your typical processing time.

And the `documentId` based on `Date.now()` means each test run uses a unique ID, so parallel runs don't interfere
with each other.

---

## Why This Pattern Is Worth It

I want to be honest with you: setting this up the first time takes maybe an hour. Writing the AsyncAPI
spec, getting Mokapi running, wiring the test together. It's not zero effort.

But here's what you get in return.

Spec-first testing. Your AsyncAPI file is the contract. Both the backend and the test work against it. If your
backend sends a field the spec doesn't expect, Mokapi will tell you. If you change the schema and forget to update
the backend, you'll catch it immediately. This was the core problem I wanted to solve with Mokapi.

Real behavior, fake infrastructure. The backend isn't mocked. It's running for real, consuming real messages,
publishing real events. The only thing that's fake is the broker, and that's exactly what you want. You're testing
your code, not Kafka's reliability.

Fast CI. No Docker Compose. No Zookeeper. Mokapi starts in seconds. Your CI pipeline doesn't need special
infrastructure or teardown logic.

It scales. Add another backend service, another topic, another step in the workflow. The test structure stays the same.
Produce, wait, verify. That's the whole pattern.

---

## One More Thing

You might be wondering if this works for more complex scenarios. Multiple services talking to each other? Topics
with Avro schemas? Consumer groups? Yes to all of it. The pattern doesn't change. Produce, wait, verify. The
AsyncAPI spec handles the schema validation across all of it, whether messages come from your real Kafka producer,
a JavaScript handler, or an HTTP call like in this example.

The full working example is in the repository: [mokapi-kafka-workflow](https://github.com/marle3003/mokapi-kafka-workflow).
Clone it, run it, and see for yourself.

If you've been putting off writing real end-to-end tests for your Kafka services because it felt too heavy, this is
exactly the problem Mokapi was built to solve.

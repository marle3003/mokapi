---
title: Testing Kafka Workflows with Playwright and Mokapi
description: Simulating real message flows end-to-end with Node.js, Kafka topics, and browser-driven tests.
icon: bi-lightning
tech: kafka
---

# Testing Kafka Workflows with Playwright and Mokapi

## Introduction

Modern applications rarely work in isolation. Instead, they communicate through **events** on Kafka topics, making
testing a challenge. How do you know your backend consumes the right command messages, processes them correctly,
and publishes the expected events?

This is where Mokapi comes in. Mokapi lets you **mock Kafka brokers** using
[AsyncAPI](https://www.asyncapi.com/) specifications or dynamic configuration. You can define your topics and schemas,
and then interact with them over a simple REST API. That means you don’t need to spin up a real Kafka cluster just
for testing.

For end-to-end testing, we combine Mokapi with [Playwright](https://playwright.dev). Normally Playwright is used for
browser UI tests, but in this tutorial we’ll use it to drive our workflow tests: producing Kafka messages, waiting for
backend processing, and validating the resulting events. (If you’re wondering: Jest would also work here, but 
Playwright makes it easy to integrate with workflows that already involve UI.)

By the end, you’ll see how a full event-driven workflow can be tested reliably without the overhead of managing a real
Kafka cluster.

---

## The Testing Scenario

Imagine this workflow:

- A foreign system publishes a **command** to the topic `document.send-command`.
- Our **backend service** (written in Node.js) consumes this command, simulates sending the document, and then publishes a **resulting event** to `document.send-event`.
- Our **test** (written in Playwright) acts as the foreign system:
  1. It produces a command message to `document.send-command` using Mokapi’s REST API.
  2. It waits for the backend to process the command.
  3. It retrieves the resulting event from `document.send-event` and verifies the contents (such as `documentId` and `status`).  

This mirrors what happens in production: a message-driven workflow across multiple services. The only difference is
that for testing, Kafka itself is mocked by Mokapi.

## Defining Kafka Topics with AsyncAPI

The heart of this setup is the AsyncAPI specification. It declares the two topics we use:
- document.send-command (input command)
- document.send-event (output event)

Here is part of the API specification:

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

## Setting Up the Backend

Our backend is a small Node.js service. It consumes commands from document.send-command, simulates sending a document, 
and publishes an event to document.send-event.

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
    eachMessage: async ({ topic, partition, message }) => {
      const value = JSON.parse(message.value.toString());
      console.log('Received command:', value);

      // Simulate sending the document
      await new Promise(res => setTimeout(res, 500));

      // Publish send-event
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

---

## Writing the Test with Playwright

Now let’s test this workflow with Playwright and Mokapi.

1. Produce a command to document.send-command. 
2. Wait for the backend to publish an event to document.send-event. 
3. Retrieve the event and assert its contents.

We can access Mokapi's REST API either through Playwright's Page object or via fetch.

```typescript
import { test, expect } from '@playwright/test';
import fetch from 'node-fetch';

const MOKAPI_API = 'http://localhost:8080/api/services/kafka/Kafka%20Mock';
const TOPIC_COMMAND = 'document.send-command';
const TOPIC_EVENT = 'document.send-event';

test('Kafka document send workflow', async () => {
    const documentId = 'doc-' + Date.now();
    let startOffset = -1
    console.log('using document ID: ' + documentId)

    await test.step('Get current offset for events', async () => {
        startOffset = await getPartitionOffset(TOPIC_EVENT, 0)
        console.log('current partition offset is: ' + startOffset)
    })

    await test.step('Produce a message to document.send-command topic', async () => {
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
        })
    })

    await test.step('Get messages from document.send-event', async () => {
        let record: any = undefined;
        const timeout = Date.now() + 5000;

        while (Date.now() < timeout && !record) {
            const records: any = await read(TOPIC_EVENT, 0, startOffset);
            console.log(records)
            record = records.find(x => x.value.documentId === documentId);
            if (record) {
                break    
            }
            startOffset += records.length
            // short delay before retry
            await new Promise(res => setTimeout(res, 200));
        }
        expect(record, 'record should be found').not.toBeNull();
        expect(record.value.status).toBe('SENT')
    })
})

async function getPartitionOffset(topic, partition) {
  const res = await fetch(`${MOKAPI_API}/topics/${topic}/partitions/${partition}`);
  const data: any = await res.json();
  return data.offset
}

async function produce(topic: string, record: {key: string, value: any}) {
  const res = await fetch(`${MOKAPI_API}/topics/${topic}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      records: [
        {
          key: record.key,
          value: record.value
        }
      ]
    })
  });
  expect(res.status).toBe(200);
  const data: any = await res.json();
  expect(data.offsets.every(x => !('error' in x))).toBe(true);
}

async function read(topic: string, partition: number, offset: number) {
  const res = await fetch(`${MOKAPI_API}/topics/${topic}/partitions/${partition}/offsets?offset=${offset}`)
  expect(res.status).toBe(200);
  return await res.json()
}
```

This test simulates the entire flow end-to-end — no mocks inside the backend, no shortcuts. Just a command going in, and an event coming out.

## Why This Matters

This example might feel simple, but the pattern is powerful:

- Spec-first: Mokapi uses AsyncAPI to define topics and message schemas. Tests validate against this shared contract. 
- Realistic simulation: you’re testing exactly how your backend would behave in production. 
- Fast feedback: no need to set up and maintain a real Kafka cluster in your CI pipeline. 
- Scalability: add more backends and topics, and the test structure remains the same.

With Mokapi and Playwright, you get a full end-to-end test for an event-driven workflow without the heavy setup.

---

## Conclusion

We’ve built a full workflow test where:

- A command is produced to Kafka. 
- A backend consumes it and publishes an event. 
- A test verifies the event using Mokapi’s REST API.

This pattern can be applied to much more complex event-driven architectures: multiple services, multiple topics, 
or more complex event payloads. However, the testing approach remains the same: Produce → Consume → Verify.

By mocking Kafka with Mokapi, you avoid the overhead of running real brokers in tests, while still validating real 
message flows. Combined with Playwright, this gives you a powerful way to ensure your event-driven applications 
work exactly as expected.

👉 You can find the full working example in the repository: [mokapi-kafka-workflow](https://github.com/marle3003/mokapi-kafka-workflow).
---
title: "Kafka API Mocking: Produce Orders with AsyncAPI & Mokapi"
description: Learn how to mock a Kafka API using AsyncAPI and Mokapi. Step-by-step guide to simulate order production, message validation, and real-time monitoring.
---
# Producing Mock Orders to an AsyncAPI Kafka Broker

## Overview

This guide walks through a complete order workflow using Mokapi as a mock Kafka broker. By the end,
you'll have:

1. Defined a Kafka instance and an `orders` topic using [AsyncAPI](https://www.asyncapi.com/).
2. Started Mokapi to simulate the broker and validate messages against the schema.
3. Produced an order message to the topic.
4. Consumed that message to verify it arrived correctly.
5. Seen what happens when a message doesn't match the schema.

## Define the AsyncAPI Specification

Create an AsyncAPI file (`asyncapi.yaml`) describing the Kafka broker and the `orders` topic.

```yaml
asyncapi: '3.0.0'
info:
  title: Orders Kafka Service
  version: 1.0.0
  description: This AsyncAPI document defines the Kafka instance for producing and consuming orders.
servers:
  mokapi:
    host: 'localhost:9092'
    protocol: kafka
    description: Mock Kafka broker provided by Mokapi.
channels:
  orders:
    address: orders
    description: Kafka topic for order messages.
    messages:
      OrderMessage:
        $ref: '#/components/messages/Order'
operations:
  publishOrder:
    action: send
    summary: Publish an order to the topic.
    channel:
      $ref: '#/channels/orders'
    messages:
      - $ref: '#/channels/orders/messages/OrderMessage'
  subscribeOrder:
    action: receive
    summary: Subscribe to orders from the topic.
    channel:
      $ref: '#/channels/orders'
    messages:
      - $ref: '#/channels/orders/messages/OrderMessage'
components:
  messages:
    Order:
      name: OrderMessage
      title: An order message
      payload:
        type: object
        properties:
          orderId:
            type: string
            description: Unique identifier for the order.
          customerId:
            type: string
            description: Unique identifier for the customer.
          totalAmount:
            type: number
            format: float
            description: Total order amount.
          items:
            type: array
            description: List of items in the order.
            items:
              type: object
              properties:
                productId:
                  type: string
                  description: Unique identifier for the product.
                quantity:
                  type: integer
                  description: Quantity of the product.
        required:
          - orderId
```

The `required` field ensures every order message includes at least an `orderId`. Mokapi enforces this
automatically once the broker is running.

## Start Mokapi

Start Mokapi with the AsyncAPI file to simulate the Kafka broker:

```bash
mokapi asyncapi.yaml
```

Mokapi's log output will confirm that a Kafka broker is running at `localhost:9092`, ready to accept
connections from producers and consumers exactly like a real broker.

## Produce an Order

Use a Kafka producer to send a message to the `orders` topic. Here's an example using the Confluent
.NET client:

```csharp
using System.Text.Json;
using Confluent.Kafka;

const string topic = "orders";
var config = new ProducerConfig
{
    BootstrapServers = "localhost:9092",
};

using var producer = new ProducerBuilder<string, string>(config).Build();

Order order = new() {
    OrderId = "12345",
    CustomerId = "0001",
    TotalAmount = 250.75F,
    Items = new() { 
        new Item{ProductId = "101", Quantity = 2},
        new Item{ProductId = "102", Quantity = 1}
    }
};

await producer.ProduceAsync(topic, new Message<string, string> { Key = "foo", Value = JsonSerializer.Serialize(order) });

Console.WriteLine("Order sent successfully!");
```

Mokapi receives the message, validates it against the `OrderMessage` schema, and stores it on the
`orders` topic.

## Consume the Order

To verify the message arrived, subscribe to the `orders` topic with a consumer:

```csharp
using Confluent.Kafka;

var config = new ConsumerConfig
{
    BootstrapServers = "localhost:9092",
    GroupId = "foo",
    AutoOffsetReset = AutoOffsetReset.Earliest
};

using var consumer = new ConsumerBuilder<Ignore, string>(config).Build();
consumer.Subscribe("orders");
while (true)
{
    var result = consumer.Consume();
    Console.WriteLine($"Message: {result.Message.Value}");
}
```

Running this alongside the producer should print the order message you just sent.

## Monitor Messages in the Dashboard

Instead of writing a consumer just to check that a message arrived, you can use Mokapi's web dashboard
to inspect Kafka traffic directly. This is often the fastest way to debug message flow during development.

1. Open the dashboard, by default at http://localhost:8080.
2. Go to the Kafka tab.
3. Click on the mocked Kafka API.
4. Select the orders topic to see all messages that have been produced to it, including their key,
value, and headers.

## Schema Validation

Every message produced to the `orders` topic is validated against the `OrderMessage` schema defined
in `asyncapi.yaml`. If a message doesn't conform to the schema:

- Mokapi rejects the message.
- Mokapi logs a validation error describing what doesn't match.

### Test an Invalid Message

Send a message that's missing the required `orderId` field:

```csharp
await producer.ProduceAsync(
    topic, 
    new Message<string, string> { 
        Key = "foo", 
        Value = @"{ ""customerId"": ""0001"", ""TotalAmount"": 250.75 }"
     }
 );
```

Mokapi rejects this message because `orderId` is required by the schema. Check Mokapi's logs, you'll
see a validation error identifying the missing field. The message won't appear on the `orders` topic
in the dashboard either, since it was never accepted.

## Next Steps

- [Mocking Kafka with AsyncAPI](/docs/kafka/overview.md): Learn more about how Mokapi simulates Kafka.
- [JavaScript API](/docs/javascript-api/overview.md): Add dynamic behavior, generate test data, or simulate errors with Mokapi Scripts.
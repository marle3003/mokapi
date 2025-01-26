---
title: Producing Orders to AsyncAPI Kafka with Mokapi
description: Learn how to set up Mokapi to produce orders to an AsyncAPI-defined Kafka instance, including message production, validation, and monitoring.
---
# Producing Orders to AsyncAPI Kafka with Mokapi

## Overview

In this guide, you will:

1. Define a Kafka instance and topics using [AsyncAPI](https://www.asyncapi.com/).
2. Use Mokapi to simulate the Kafka broker and validate messages against the schema.
3. Simulate a producer that sends order messages into the Kafka topic.
4. Simulate a consumer to verify the flow of messages.

## Define the AsyncAPI Specification

Create an AsyncAPI file (asyncapi.yaml) to define the Kafka instance and orders topic.

```yaml
asyncapi: '2.6.0'
info:
  title: Orders Kafka Service
  version: 1.0.0
  description: This AsyncAPI document defines the Kafka instance for producing and consuming orders.
servers:
  mokapi:
    url: localhost:9092
    protocol: kafka
    description: Mock Kafka broker provided by Mokapi.
channels:
  orders:
    description: Kafka topic for order messages.
    publish:
      summary: Publish an order to the topic.
      message: 
        $ref: '#/components/messages/Order'
    subscribe:
      summary: Subscribe to orders from the topic.
      message:
        $ref: '#/components/messages/Order'
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

## Simulate the Kafka Instance with Mokapi

Start the Mokapi server with the AsyncAPI file to simulate the Kafka broker.

```bash
mokapi --providers-file-filename asyncapi.yaml
```
You should see output confirming that Mokapi is running a Kafka broker at localhost:9092.


## Produce Orders to the Kafka Instance
Use a Kafka producer to send messages to the orders topic.

```c#
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

## Consume Orders from the Kafka Instance

To verify the messages, use a Kafka consumer to subscribe to the `orders` topic.

```c#
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

While the consumer application can process messages programmatically, Mokapi provides 
a web-based dashboard to easily monitor and inspect the messages produced to Kafka topics. 
This can be helpful during development or debugging to ensure that messages are reaching the intended topic.

1. Open the dashboard in your browser, typically available at [http://localhost:8080](http://localhost:8080)
2. Go to the Kafka tab
3. Here, you'll see all Kafka topics defined in your AsyncAPI file

## Schema Validation

When using Mokapi, messages produced to the orders topic are validated against the OrderMessage schema defined in the asyncapi.yaml. If a message doesn’t conform to the schema:

- The message is rejected.
- Errors are logged by Mokapi.

### Test Invalid Message

Try sending an invalid message (e.g., missing orderId):

```c#
await producer.ProduceAsync(
    topic, 
    new Message<string, string> { 
        Key = "foo", 
        Value = @"{ ""customerId"": ""0001"", ""TotalAmount"": 250.75 }"
     }
 );
```
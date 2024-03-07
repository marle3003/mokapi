---
title: Apache Kafka Quick Start - Mock a topic in seconds with AsyncAPI
description: A simple Use Case to use Mokapi to produce orders into a Kafka topic
---
# Quick Start

A simple Use Case to use Mokapi to produce orders into an [AsyncAPI](https://www.asyncapi.com/) Kafka instance

## Create a Kafka Topic
First, create an AsyncAPI specification `orders.yaml` for your Kafka instance that defines the topic and messages you want to produce and consume.

```yaml
asyncapi: '2.0.0'
info:
  title: Orders API
  version: 1.0.0
servers:
  mokapi:
    url: localhost:9092
    protocol: kafka
channels:
  orders:
    publish:
      message: 
        $ref: '#/components/messages/Order'
    subscribe:
      message:
        $ref: '#/components/messages/Order'
components:
  messages:
    Order:
      payload:
        type: object
        properties:
          orderId:
            type: integer
          customer:
            type: string
          items:
            type: array
            items:
              type: object
              properties:
                itemId:
                  type: integer
                quantity:
                  type: integer
```
This specification defines a topic called "orders" that accepts messages in the format of an "Order" object which has an orderId, a customer name, and an array of items, each with an itemId and a quantity.

## Create a Producer
Next, create a JavaScript file `orders.js` to generate messages for your topic

```javascript
import { produce } from 'mokapi/kafka'

export default function () {
    produce({ topic: 'orders' })
    produce({ topic: 'orders', value: {orderId: 1, customer: 'Alice', items: [{itemId: 200, quantity: 3}]} })
}
```
This script creates two messages to the "orders" topic. With the first call, Mokapi creates a random message.

## Create a Dockerfile
Next create a `Dockerfile` to configure Mokapi
```dockerfile
FROM mokapi/mokapi:latest

COPY ./orders.yaml /demo/

COPY ./orders.js /demo/

CMD ["--Providers.File.Directory=/demo"]
```

## Start Mokapi

```
docker run -p 8080:8080 -p 9092:9092 --rm -it $(docker build -q .)
```
You can now open a browser and go to Mokapi's Dashboard (`http://localhost:8080`) to see the Mokapi's Kafka Topic and the produced messages.

## Use a Kafka Consumer

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
You can see the group and consumer on Mokapi's Dashboard. You can find more information about Kafka Consumers [here](https://docs.confluent.io/home/clients/overview.html#clients)  
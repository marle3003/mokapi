---
title: Get started with Kafka
description: Learn how to mock a Kafka Topic and verify that your producer generates valid messages according your AsyncAPI specification.
icon: bi-lightning
---

# Get started with Kafka

This tutorial provides a step-by-step guide to mocking a Kafka topic using Mokapi 
and verifying that a producer generates valid messages based on an AsyncAPI specification. 
This approach enables the simulation of Kafka environments without requiring a live Kafka cluster.

## Create AsyncAPI file

Let's start by creating a file named `kafka.yml` with the following content.

```yaml
asyncapi: '2.0.0'
info:
  title: Kafka Cluster
  description: A kafka test cluster
  version: '1.0'
  contact:
    name: Mokapi
    url: https://mokapi.io
    email: mokapi@mokapi.io
servers:
  broker:
    url: 127.0.0.1:9092
    protocol: kafka
channels:
  users:
    subscribe:
      message:
        $ref: '#/components/messages/user'
    publish:
      message:
         $ref: '#/components/messages/user'
    bindings:
      kafka:
        partitions: 2

components:
  messages:
    user:
      contentType: application/json
      payload:
        $ref: '#/components/schemas/user'
      bindings:
        kafka:
          key:
            type: string

  schemas:
    user:
      type: object
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        email:
          type: string
      required:
        - id
        - name
        - email
```

This creates a Kafka topic `users` with two partitions. The content type of the message's payload 
is `application/json` and the object must have the properties `id`, `name` and `email`.

``` box=info
When Mokapi receives an invalid message it will return the error `CORRUPT_MESSAGE`and logs an
error, for instance `kafka: invalid message received for topic users: missing required field name...`
```

## Create a Dockerfile

Next create a `Dockerfile` to configure Mokapi to use the AsyncAPI specification file

```dockerfile
FROM mokapi/mokapi:latest

COPY ./kafka.yaml /demo/

CMD ["--Providers.File.Directory=/demo"]
```

## Start Mokapi

Now we can start our container and verify in Mokapi's Dashboard (http://localhost:8080) our mocked Kafka topic

```
docker run -p 9092:9092 -p 8080:8080 --rm -it $(docker build -q .)
```

## Create a Kafka Producer

You can now produce messages to the mocked Kafka topic. Below is an example using C#.

```csharp
public class Producer
{
    public static async Task Run()
    {
        var config = new ProducerConfig { BootstrapServers = "localhost:9092" };

        using var producer = new ProducerBuilder<string, string>(config).Build();
        var topic = new TopicPartition("users", new Partition(1));

        var serializeOptions = new JsonSerializerOptions
        {
            PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
        };
        
        var result = await producer.ProduceAsync(topic, new Message<string, string> {
            Key = "alice",
            Value = JsonSerializer.Serialize(new {
                Id = "dd5742d1-82ad-4d42-8960-cb21bd02f3e7",
                Name = "Alice",
                Email = "alice@foo.bar",
            }, serializeOptions),
        });
    }
}
```

## Verifying Messages with Mokapi

Mokapi provides a dashboard to monitor and verify messages sent to the mocked Kafka topics.

1. <p><strong>Access the Mokapi Dashboard:</strong><br />Open your browser and navigate to http://localhost:8080/dashboard.

2. <p><strong>Navigate to the Kafka Section:</strong><br />In the dashboard, select the Kafka section to view the topics and messages.

3. <p><strong>Verify the Message:</strong><br />Locate the users topic and verify that the message sent by your producer appears as expected. Mokapi will validate the message against the AsyncAPI specification.

<img src="/docs/resources/tutorials/simple-kafka-example.png" alt="Mokapi Kafka Dashboard" title="Mokapi Kafka Dashboard" />

## Create a Kafka Consumer

Now we can use a .NET Consumer to consume our first sent message.

```csharp
public class Consumer
{
    public static void Run()
    {
        var config = new ConsumerConfig
        {
            BootstrapServers = "localhost:9092",
            GroupId = "foo",
            AutoOffsetReset = AutoOffsetReset.Earliest
        };

        using var consumer = new ConsumerBuilder<Ignore, string>(config).Build();
        consumer.Subscribe("users");

        CancellationTokenSource cts = new CancellationTokenSource();

        while (true)
        {
            var result = consumer.Consume(cts.Token);
            Console.WriteLine($"Consumed message '{result.Message.Value}' offset: {result.TopicPartitionOffset.Offset} partition: {result.TopicPartition.Partition}");
        }
    }
}
```
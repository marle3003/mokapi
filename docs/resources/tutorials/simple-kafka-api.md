---
title: Get started with Kafka
description: Learn how to mock a Kafka Topic and verify that your producer generates valid messages according your AsyncAPI specification.
subtitle: Learn how to mock a Kafka topic using AsyncAPI specifications with Mokapi. Validate producer messages and test consumers without requiring a live Kafka cluster.
tags: [Kafka, AsyncAPI]
icon: bi-lightning
tech: kafka
cards:
  items:
    - title: Testing Kafka Workflows with Playwright and Mokapi
      href: /resources/blogs/testing-kafka-workflows-playwright
      description: Simulating real message flows end-to-end with Node.js, Kafka topics, and browser-driven tests.
---

# Get started with Kafka

This tutorial provides a step-by-step guide to mocking a Kafka topic using Mokapi and verifying
that a producer generates valid messages based on an AsyncAPI specification. This approach enables
simulation of Kafka environments without requiring a live Kafka cluster.

In this tutorial, you will:
- Create an AsyncAPI specification defining a Kafka topic with schema validation
- Configure and run Mokapi as a mock Kafka broker in Docker
- Write a .NET producer that sends messages to the mocked topic
- Verify message validation and monitoring using the Mokapi dashboard
- Create a .NET consumer to read messages from the mocked topic

``` box=tree title="Project Structure"
📄 kafka.yaml
📄 Dockerfile
📁 Kafka.GetStarted (optional)
    📄 Consumer.cs
    📄 Kafka.GetStarted.csproj
    📄 Producer.cs
    📄 Program.cs
```

``` box=info
You can find the [full working example](https://github.com/marle3003/mokapi/tree/main/examples/kafka/get-started) in the examples.
```

## <i class="bi bi-1-circle-fill align-baseline"></i> Create AsyncAPI file

Create a file named `kafka.yaml` that defines a Kafka topic with message validation:

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

### Specification Breakdown
- **Topic:** Creates a `users` topic with 2 partitions
- **Content Type:** Messages must be JSON (`application/json`)
- **Schema:** Each message must have `id` (UUID), `name` (string), and `email` (string)
- **Validation:** All three fields are required; missing fields will cause validation errors
- **Key Type:** Message keys must be strings

``` box=info title="Message Validation"
When Mokapi receives an invalid message, it will return the error CORRUPT_MESSAGE and log an error message 
like: kafka: invalid message received for topic users: missing required field name...
```

## <i class="bi bi-2-circle-fill align-baseline"></i> Create a Dockerfile

Create a `Dockerfile` to configure Mokapi with the AsyncAPI specification:

```dockerfile
FROM mokapi/mokapi:latest

COPY ./kafka.yaml /demo/

CMD ["--Providers.File.Directory=/demo"]
```

### Dockerfile Breakdown
- **Base Image:** Uses the official Mokapi Docker image
- **Copy Specification:** Copies the AsyncAPI specification into the container
- **Configuration:** Instructs Mokapi to load specifications from the `/demo` directory

## <i class="bi bi-3-circle-fill align-baseline"></i> Start Mokapi

Build and run the Docker container, exposing the Kafka broker and dashboard ports:

```
docker run -p 9092:9092 -p 8080:8080 --rm -it $(docker build -q .)
```

This command:
- **Builds** the Docker image from your Dockerfile
- **Exposes port 9092** for the Kafka broker protocol
- **Exposes port 8080** for the Mokapi dashboard
- **Runs interactively** with automatic cleanup when stopped

``` box=result title="Mokapi is Running"
Open your browser and navigate to the Mokapi Dashboard at <code>http://localhost:8080</code> to see your mocked Kafka 
topic. You can verify the topic configuration, view messages, and monitor producer/consumer activity.
```

## <i class="bi bi-4-circle-fill align-baseline"></i> Create a Kafka Producer

Now you can produce messages to the mocked Kafka topic. Below is an example using C# with the Confluent.Kafka library:

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

### Producer Code Breakdown

- **Connection:** Connects to the Mokapi broker at `localhost:9092`
- **Topic & Partition:** Targets the `users` topic, partition 1
- **Message Key:** Uses `"alice"` as the message key (required by the AsyncAPI spec)
- **Message Value:** Serializes a JSON object with `id`, `name`, and `email` fields
- **JSON Format:** Uses camelCase naming to match the AsyncAPI schema

``` box=info title="Schema Validation"
Mokapi validates every message against the AsyncAPI schema. If you send a message missing the <code>name</code> field or 
with an invalid UUID format for <code>id</code>, Mokapi will reject it with a <code>CORRUPT_MESSAGE</code> error.
```

## <i class="bi bi-5-circle-fill align-baseline"></i> Verifying Messages with Mokapi

Mokapi provides a comprehensive dashboard to monitor and verify messages sent to mocked Kafka topics.

### Dashboard Verification Steps

1. **Access the Mokapi Dashboard:** Open your browser and navigate to `http://localhost:8080/dashboard`
2. **Navigate to the Kafka Section:** In the dashboard, select the Kafka section to view configured topics and their messages
3. **Verify the Message:** Locate the `users` topic and verify that the message sent by your producer appears as expected. Mokapi displays the message key, value, partition, offset, and validation status

![Mokapi Kafka Dashboard](/docs/resources/tutorials/simple-kafka-example.png "Mokapi Kafka Dashboard displaying validated messages in the users topic")

The dashboard shows:
- **Message Content:** Full JSON payload with schema validation status
- **Partition & Offset:** Where the message was stored in the topic
- **Validation Errors:** Any schema mismatches or missing required fields
- **Timestamp:** When the message was received by Mokapi

## <i class="bi bi-6-circle-fill align-baseline"></i> Create a Kafka Consumer

Now create a .NET consumer to read messages from the mocked Kafka topic:

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

### Consumer Code Breakdown
- **Connection:** Connects to the Mokapi broker at `localhost:9092`
- **Consumer Group:** Joins consumer group `"foo"`
- **Offset Reset:** Starts reading from the earliest available 
- **Subscribe:** Subscribes to the `users` topic
- **Consume Loop:** Continuously polls for new messages and prints them to the console

``` box=result title="Expected Console Output"
Consumed message '{"id":"dd5742d1-82ad-4d42-8960-cb21bd02f3e7","name":"Alice","email":"alice@foo.bar"}' offset: 0 partition: 1
```

## Next Steps

Now that you have a working Kafka mock with producer and consumer, explore these advanced topics:

{{ card-grid key="cards" }}
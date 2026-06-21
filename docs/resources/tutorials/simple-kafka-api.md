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

By the end of this tutorial you'll have a Kafka broker running entirely in Docker, with no real
cluster anywhere in sight, validating every message against a schema you define. You'll produce a
message with a .NET producer, watch Mokapi accept or reject it based on your AsyncAPI spec, and
consume it back out with a .NET consumer.

You'll work through six steps:

- Write an AsyncAPI specification defining a Kafka topic with schema validation
- Run Mokapi as a mock Kafka broker in Docker
- Produce a message from a .NET producer
- Watch Mokapi validate it in real time
- Inspect the message in the Mokapi dashboard
- Consume the message back out with a .NET consumer

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

Create a file named `kafka.yaml`. This is the contract Mokapi will enforce: it defines the topic,
the partitions, and the schema every message must satisfy.

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

This spec creates a `users` topic with two partitions. Every message must be JSON, must include
`id` (a UUID), `name`, and `email`, and all three fields are required. Send a message missing
any of them and Mokapi will reject it, which you'll see for yourself in step four.

``` box=info title="Message Validation"
When Mokapi receives an invalid message, it returns the error <code>CORRUPT_MESSAGE</code> and logs
something like: <code>kafka: invalid message received for topic users: missing required field name...</code>
```

## <i class="bi bi-2-circle-fill align-baseline"></i> Package Mokapi with Your Spec

Create a `Dockerfile` that bundles Mokapi with the spec you just wrote:

```dockerfile
FROM mokapi/mokapi:latest

COPY ./kafka.yaml /demo/

CMD ["--providers-file-directory", "/demo"]
```

The base image is the official Mokapi container. The `COPY` line bundles your spec into the
image, and the `CMD` tells Mokapi to load every spec it finds in `/demo` on startup.

## <i class="bi bi-3-circle-fill align-baseline"></i> Start the Broker

Build and run the container, exposing both the Kafka protocol port and the dashboard:

```
docker run -p 9092:9092 -p 8080:8080 --rm -it $(docker build -q .)
```

Port 9092 is where Kafka clients connect, exactly as they would to a real broker. Port 8080 is
Mokapi's dashboard, where you'll inspect traffic in a moment. The `--rm` flag cleans up the
container automatically once you stop it.

``` box=result title="Mokapi is Running"
Open your browser and navigate to the Mokapi Dashboard at <code>http://localhost:8080</code> to see your mocked Kafka 
topic. You can verify the topic configuration, view messages, and monitor producer/consumer activity.
```

## <i class="bi bi-4-circle-fill align-baseline"></i> Produce a Message

With the broker running, produce a message using a standard Kafka client. Here's an example in
C# with Confluent.Kafka, but any Kafka client works exactly the same way since Mokapi speaks the
real Kafka protocol.

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

This connects to Mokapi at `localhost:9092`, same as a real broker. It targets partition 1 of the
`users` topic, uses `"alice"` as the key, and sends a JSON payload with camelCase fields to match
the schema. Mokapi validates this message the instant it arrives.

``` box=info title="Schema Validation"
Remove the <code>Name</code> field from the payload and run the producer again. Mokapi rejects the
message with a <code>CORRUPT_MESSAGE</code> error instead of silently accepting bad data, exactly
as it would reject a message that violates your contract in production.
```

## <i class="bi bi-5-circle-fill align-baseline"></i> Inspect the Message in the Dashboard

Instead of writing a consumer just to confirm your message arrived, check the dashboard first.

1. Open http://localhost:8080/dashboard in your browser.
2. Go to the Kafka section to see your configured topics.
3. Select the users topic and find the message you just sent. Mokapi shows the key, value, partition, offset, and validation status for every message that's passed through.

![Mokapi Kafka Dashboard](/docs/resources/tutorials/simple-kafka-example.png "Mokapi Kafka Dashboard displaying validated messages in the users topic")

If you sent the broken message from the previous step too, you'll see it flagged with a
validation error right there in the dashboard, no log diving required.

## <i class="bi bi-6-circle-fill align-baseline"></i> Consume the Message

Now read it back out with a consumer:

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

This joins consumer group `foo`, starts reading from the earliest offset, and subscribes to
`users`. Run it alongside the producer, and you'll see the message print to the console exactly
as it was sent.

``` box=result title="Expected Console Output"
Consumed message '{"id":"dd5742d1-82ad-4d42-8960-cb21bd02f3e7","name":"Alice","email":"alice@foo.bar"}' offset: 0 partition: 1
```

## Next Steps

You now have a working Kafka mock with a validated producer and a working consumer, all without a
real cluster anywhere in the picture. From here, explore these advanced topics:

{{ card-grid key="cards" }}
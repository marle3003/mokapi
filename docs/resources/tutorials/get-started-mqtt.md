---
title: Get started with MQTT
description: Learn how to mock an MQTT broker and verify that your publisher generates valid messages according to your AsyncAPI specification.
subtitle: Learn how to mock an MQTT broker using AsyncAPI specifications with Mokapi. Validate publisher messages and test subscribers without requiring real devices or a live broker.
tags: [MQTT, AsyncAPI]
icon: bi-broadcast
tech: mqtt
---

# Get started with MQTT

By the end of this tutorial you'll have an MQTT broker running entirely in Docker, with no real
devices or infrastructure anywhere in sight, validating every message against a schema you define.
You'll publish a device status update with a .NET client, watch Mokapi accept or reject it based
on your AsyncAPI spec, and subscribe to read it back out.

You'll work through six steps:
- Write an AsyncAPI specification defining an MQTT topic with schema validation
- Run Mokapi as a mock MQTT broker in Docker
- Publish a message from a .NET client
- Watch Mokapi validate it in real time
- Inspect the message in the Mokapi dashboard
- Subscribe and read the message back out

``` box=tree title="Project Structure"
📄 mqtt.yaml
📄 Dockerfile
📁 Mqtt.GetStarted (optional)
    📄 Publisher.cs
    📄 Subscriber.cs
    📄 Mqtt.GetStarted.csproj
    📄 Program.cs
```

## <i class="bi bi-1-circle-fill align-baseline"></i> Define the Topic in AsyncAPI

Create a file named `mqtt.yaml`. This is the contract Mokapi will enforce: it defines the topic
and the schema every message must satisfy.

```yaml
asyncapi: 3.0.0
info:
  title: Device Fleet
  description: A fleet of connected devices reporting status over MQTT
  version: '1.0'
  contact:
    name: Mokapi
    url: https://mokapi.io
    email: mokapi@mokapi.io

servers:
  broker:
    host: 127.0.0.1:1883
    protocol: mqtt

channels:
  deviceStatus:
    address: 'devices/status'
    messages:
      status:
        $ref: '#/components/messages/status'

operations:
  publishStatus:
    action: send
    channel:
      $ref: '#/channels/deviceStatus'
    messages:
      - $ref: '#/channels/deviceStatus/messages/status'
  subscribeStatus:
    action: receive
    channel:
      $ref: '#/channels/deviceStatus'
    messages:
      - $ref: '#/channels/deviceStatus/messages/status'

components:
  messages:
    status:
      contentType: application/json
      payload:
        $ref: '#/components/schemas/status'

  schemas:
    status:
      type: object
      properties:
        deviceId:
          type: string
          format: uuid
        state:
          type: string
          enum: [online, offline, error]
        batteryLevel:
          type: integer
          minimum: 0
          maximum: 100
      required:
        - deviceId
        - state
```

This spec creates a `devices/status` topic. Every message must be JSON, must include `deviceId`
(a UUID) and `state` (one of `online`, `offline`, or `error`), and may optionally include a
`batteryLevel` between 0 and 100. Publish a message missing a required field, or with a `state`
value outside the allowed list, and Mokapi rejects it, which you'll see for yourself in step four.

``` box=info title="Message Validation"
When Mokapi receives an invalid message, it rejects it and logs something like:
<code>mqtt: invalid message received for topic devices/status: 'state' must be one of online, offline, error...</code>
```

## <i class="bi bi-2-circle-fill align-baseline"></i> Package Mokapi with Your Spec

Create a `Dockerfile` that bundles Mokapi with the spec you just wrote:

```dockerfile
FROM mokapi/mokapi:latest

COPY ./mqtt.yaml /demo/

CMD ["--Providers.File.Directory=/demo"]
```

The base image is the official Mokapi container. The `COPY` line bundles your spec into the
image, and the `CMD` tells Mokapi to load every spec it finds in `/demo` on startup.

## <i class="bi bi-3-circle-fill align-baseline"></i> Start the Broker

Build and run the container, exposing both the MQTT protocol port and the dashboard:

```
docker run -p 1883:1883 -p 8080:8080 --rm -it $(docker build -q .)
```

Port 1883 is the standard MQTT port, where any MQTT client connects exactly as it would to a real
broker. Port 8080 is Mokapi's dashboard, where you'll inspect traffic in a moment. The `--rm` flag
cleans up the container automatically once you stop it.

``` box=result title="Mokapi is Running"
Open your browser and navigate to the Mokapi Dashboard at <code>http://localhost:8080</code> to see your mocked MQTT 
topic. You can verify the topic configuration, view messages, and monitor publisher/subscriber activity.
```

## <i class="bi bi-4-circle-fill align-baseline"></i> Publish a Message

With the broker running, publish a message using a standard MQTT client. Here's an example in C#
with MQTTnet, but any MQTT client works the same way since Mokapi speaks the real MQTT protocol.

```csharp
public class Publisher
{
    public static async Task Run()
    {
        var factory = new MqttFactory();
        using var client = factory.CreateMqttClient();

        var options = new MqttClientOptionsBuilder()
            .WithTcpServer("localhost", 1883)
            .Build();

        await client.ConnectAsync(options);

        var status = new
        {
            deviceId = "dd5742d1-82ad-4d42-8960-cb21bd02f3e7",
            state = "online",
            batteryLevel = 87
        };

        var message = new MqttApplicationMessageBuilder()
            .WithTopic("devices/status")
            .WithPayload(JsonSerializer.Serialize(status))
            .Build();

        await client.PublishAsync(message);
    }
}
```

This connects to Mokapi at `localhost:1883`, same as a real broker, and publishes a JSON payload
to `devices/status` matching the schema. Mokapi validates this message the instant it arrives.

``` box=info title="Try Breaking It"
Change <code>state</code> to something not in the schema, like <code>"sleeping"</code>, and run the
publisher again. Mokapi rejects the message instead of silently accepting bad data, exactly as it
would reject a message that violates your contract in production.
```

## <i class="bi bi-5-circle-fill align-baseline"></i> Inspect the Message in the Dashboard

Instead of writing a subscriber just to confirm your message arrived, check the dashboard first.

1. Open `http://localhost:8080/dashboard` in your browser.
2. Go to the **MQTT** section to see your configured topics.
3. Select the `devices/status` topic and find the message you just published. Mokapi shows the
   payload and validation status for every message that's passed through.

If you published the broken message from the previous step too, you'll see it flagged with a
validation error right there in the dashboard, no log diving required.

## <i class="bi bi-6-circle-fill align-baseline"></i> Subscribe and Read the Message

Now read it back out with a subscriber:

```csharp
public class Subscriber
{
    public static async Task Run()
    {
        var factory = new MqttFactory();
        using var client = factory.CreateMqttClient();

        var options = new MqttClientOptionsBuilder()
            .WithTcpServer("localhost", 1883)
            .Build();

        client.ApplicationMessageReceivedAsync += e =>
        {
            var payload = e.ApplicationMessage.ConvertPayloadToString();
            Console.WriteLine($"Received on '{e.ApplicationMessage.Topic}': {payload}");
            return Task.CompletedTask;
        };

        await client.ConnectAsync(options);
        await client.SubscribeAsync("devices/status");

        Console.WriteLine("Subscribed. Waiting for messages...");
        await Task.Delay(Timeout.Infinite);
    }
}
```

This connects to Mokapi, subscribes to `devices/status`, and prints every message it receives.
Run it alongside the publisher, and you'll see the status update print to the console exactly as
it was sent.

``` box=result title="Expected Console Output"
Subscribed. Waiting for messages...
Received on 'devices/status': {"deviceId":"dd5742d1-82ad-4d42-8960-cb21bd02f3e7","state":"online","batteryLevel":87}
```

## Next Steps

You now have a working MQTT mock with a validated publisher and a working subscriber, all without
a real broker or any physical devices anywhere in the picture. From here, explore Mokapi's
[JavaScript API](/docs/javascript-api/overview.md) to add dynamic behavior, simulate device
failures, or build out a fuller fleet simulation.
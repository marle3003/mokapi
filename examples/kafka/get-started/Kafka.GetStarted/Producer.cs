using System;
using System.Text.Json;
using System.Threading.Tasks;
using Confluent.Kafka;

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

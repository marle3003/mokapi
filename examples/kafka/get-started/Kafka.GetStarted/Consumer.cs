using System;
using System.Threading;
using Confluent.Kafka;

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

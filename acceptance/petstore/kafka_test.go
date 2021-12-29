package petstore

import (
	"mokapi/acceptance/cmd"
	"mokapi/config/static"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/metaData"
	"mokapi/kafka/protocol/produce"
	"mokapi/test"
	"testing"
	"time"
)

func TestKafka_TopicConfig(t *testing.T) {
	cfg := static.NewConfig()
	cfg.Providers.File.Filename = "./asyncapi.yml"
	cmd, err := cmd.Start(cfg)
	test.Ok(t, err)
	defer cmd.Stop()

	// wait for kafka start
	time.Sleep(500 * time.Millisecond)
	c := kafkatest.NewClient("127.0.0.1:9092", "test")
	defer c.Close()

	r, err := c.Metadata(0, &metaData.Request{})
	test.Ok(t, err)
	test.Equals(t, 1, len(r.Topics))
	test.Equals(t, "petstore.order-event", r.Topics[0].Name)
	test.Equals(t, 2, len(r.Topics[0].Partitions))
}

func TestKafka_Produce_InvalidFormat(t *testing.T) {
	cfg := static.NewConfig()
	cfg.Providers.File.Filename = "./asyncapi.yml"
	cmd, err := cmd.Start(cfg)
	test.Ok(t, err)
	defer cmd.Stop()

	// wait for kafka start
	time.Sleep(500 * time.Millisecond)
	c := kafkatest.NewClient("127.0.0.1:9092", "test")
	defer c.Close()

	r, err := c.Produce(0, &produce.Request{Topics: []produce.RequestTopic{
		{Name: "petstore.order-event", Partitions: []produce.RequestPartition{
			{
				Index: 0,
				Record: protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Offset:  0,
							Time:    time.Now(),
							Key:     protocol.NewBytes([]byte(`foo`)),
							Value:   protocol.NewBytes([]byte(`{}`)),
							Headers: nil,
						},
					},
				},
			},
		},
		}},
	})
	test.Ok(t, err)
	test.Equals(t, "petstore.order-event", r.Topics[0].Name)
	test.Equals(t, protocol.CorruptMessage, r.Topics[0].Partitions[0].ErrorCode)
	test.Equals(t, int64(0), r.Topics[0].Partitions[0].BaseOffset)
}

func TestKafka_Produce(t *testing.T) {
	cfg := static.NewConfig()
	cfg.Providers.File.Filename = "./asyncapi.yml"
	cmd, err := cmd.Start(cfg)
	test.Ok(t, err)
	defer cmd.Stop()

	// wait for kafka start
	time.Sleep(500 * time.Millisecond)
	c := kafkatest.NewClient("127.0.0.1:9092", "test")
	defer c.Close()

	r, err := c.Produce(0, &produce.Request{Topics: []produce.RequestTopic{
		{Name: "petstore.order-event", Partitions: []produce.RequestPartition{
			{
				Index: 0,
				Record: protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Offset:  0,
							Time:    time.Now(),
							Key:     protocol.NewBytes([]byte(`foo`)),
							Value:   protocol.NewBytes([]byte(`{"id": 12345}`)),
							Headers: nil,
						},
					},
				},
			},
		},
		}},
	})
	test.Ok(t, err)
	test.Equals(t, "petstore.order-event", r.Topics[0].Name)
	test.Equals(t, protocol.None, r.Topics[0].Partitions[0].ErrorCode)
	test.Equals(t, int64(0), r.Topics[0].Partitions[0].BaseOffset)
}

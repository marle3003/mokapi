package kafka_test

import (
	"mokapi/server/kafka/kafkatest"
	"mokapi/server/kafka/memory"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/produce"
	"mokapi/test"
	"testing"
	"time"
)

func TestProduce(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, b *kafkatest.Broker)
	}{
		{
			"default",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetCluster(memory.NewCluster(memory.Schema{Topics: []memory.TopicSchema{{Name: "foo", Partitions: []memory.PartitionSchema{{Index: 0}}}}}))
				r, err := b.Client().Produce(3, &produce.Request{Topics: []produce.RequestTopic{
					{Name: "foo", Data: produce.RequestPartition{
						Partition: 0,
						Record: protocol.RecordBatch{
							Records: []protocol.Record{
								{
									Offset:  0,
									Time:    time.Now(),
									Key:     []byte("foo"),
									Value:   []byte("bar"),
									Headers: nil,
								},
							},
						},
					},
					}},
				})
				test.Ok(t, err)
				test.Equals(t, "foo", r.Topics[0].Name)
				test.Equals(t, protocol.None, r.Topics[0].ErrorCode)
			},
		},
	}

	for _, data := range testdata {
		d := data
		t.Parallel()
		t.Run(d.name, func(t *testing.T) {
			b := kafkatest.NewBroker()
			defer b.Close()
			d.fn(t, b)
		})
	}
}

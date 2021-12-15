package kafka_test

import (
	"fmt"
	"mokapi/server/kafka/kafkatest"
	"mokapi/server/kafka/memory"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/metaData"
	"mokapi/test"
	"strings"
	"testing"
)

func TestMetadata(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, b *kafkatest.Broker)
	}{
		{
			"default",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetCluster(memory.NewCluster(memory.Schema{
					Topics: []memory.TopicSchema{
						{Name: "foo", Partitions: []memory.PartitionSchema{{Index: 1, Replicas: []int{0}}}}},
					Brokers: []memory.BrokerSchema{{Id: 0, Host: "foohost", Port: 9092}},
				}))
				r, err := b.Client().Metadata(4, &metaData.Request{})
				test.Ok(t, err)

				// controller
				test.Equals(t, int32(0), r.ControllerId)

				// brokers
				test.Equals(t, 1, len(r.Brokers))
				test.Equals(t, int32(0), r.Brokers[0].NodeId)
				test.Equals(t, "foohost", r.Brokers[0].Host)
				test.Equals(t, int32(9092), r.Brokers[0].Port)
				test.Equals(t, "", r.Brokers[0].Rack)

				// topics
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, "foo", r.Topics[0].Name)
				test.Equals(t, protocol.None, r.Topics[0].ErrorCode)
				test.Equals(t, 1, len(r.Topics[0].Partitions))
				test.Equals(t, int32(0), r.Topics[0].Partitions[0].PartitionIndex)
				test.Equals(t, int32(0), r.Topics[0].Partitions[0].LeaderId) // default broker id is 0
				test.Equals(t, 1, len(r.Topics[0].Partitions[0].ReplicaNodes))
				test.Equals(t, int32(0), r.Topics[0].Partitions[0].ReplicaNodes[0])
				test.Equals(t, 1, len(r.Topics[0].Partitions[0].IsrNodes))
				test.Equals(t, int32(0), r.Topics[0].Partitions[0].IsrNodes[0])
				test.Equals(t, false, r.Topics[0].IsInternal)
			},
		},
		{
			"with specific topic and two partitions",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetCluster(memory.NewCluster(memory.Schema{
					Topics: []memory.TopicSchema{
						{Name: "foo", Partitions: []memory.PartitionSchema{
							{Index: 1, Replicas: []int{0}},
							{Index: 2, Replicas: []int{0}},
						}},
						{Name: "foo2", Partitions: []memory.PartitionSchema{
							{Index: 1, Replicas: []int{0}},
						}}},
					Brokers: []memory.BrokerSchema{{Id: 0, Host: "foohost", Port: 9092}},
				}))

				r, err := b.Client().Metadata(4, &metaData.Request{
					Topics: []metaData.TopicName{{Name: "foo"}},
				})
				test.Ok(t, err)
				test.Equals(t, 2, len(r.Topics[0].Partitions))
			},
		},
		{
			"with invalid topic",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetCluster(memory.NewCluster(memory.Schema{
					Topics: []memory.TopicSchema{
						{Name: "foo", Partitions: []memory.PartitionSchema{
							{Index: 1, Replicas: []int{0}},
						}}},
					Brokers: []memory.BrokerSchema{{Id: 0, Host: "foohost", Port: 9092}},
				}))
				r, err := b.Client().Metadata(4, &metaData.Request{
					Topics: []metaData.TopicName{{Name: "foo"}, {Name: "bar"}},
				})
				test.Ok(t, err)
				test.Equals(t, 2, len(r.Topics))
				test.Equals(t, protocol.None, r.Topics[0].ErrorCode)
				test.Equals(t, protocol.UnknownTopicOrPartition, r.Topics[1].ErrorCode)
			},
		},
		{
			"with invalid topic name",
			func(t *testing.T, b *kafkatest.Broker) {
				for _, name := range []string{"", ".", "..", "event?", strings.Repeat("a", 250)} {
					testName := name
					if len(name) > 10 {
						testName = testName[:10] + "..."
					}
					t.Run(fmt.Sprintf("name %q", testName), func(t *testing.T) {
						r, err := b.Client().Metadata(4, &metaData.Request{
							Topics: []metaData.TopicName{{Name: name}},
						})
						test.Ok(t, err)
						test.Equals(t, 1, len(r.Topics))
						test.Equals(t, protocol.InvalidTopic, r.Topics[0].ErrorCode)
					})
				}
			},
		},
	}

	for _, data := range testdata {
		d := data
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			b := kafkatest.NewBroker()
			defer b.Close()

			d.fn(t, b)
		})
	}
}

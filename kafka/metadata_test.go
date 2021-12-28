package kafka_test

import (
	"fmt"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/metaData"
	"mokapi/test"
	"net"
	"strconv"
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
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{
					Brokers: []string{b.Listener.Addr().String()},
					Topics:  []kafkatest.TopicConfig{{"foo", 1}}}))
				r, err := b.Client().Metadata(4, &metaData.Request{})
				test.Ok(t, err)

				// controller
				test.Equals(t, int32(0), r.ControllerId)

				// brokers
				test.Equals(t, 1, len(r.Brokers))
				test.Equals(t, int32(0), r.Brokers[0].NodeId)
				test.Equals(t, "127.0.0.1", r.Brokers[0].Host)
				_, ps, _ := net.SplitHostPort(b.Listener.Addr().String())
				p, _ := strconv.ParseInt(ps, 10, 32)
				test.Equals(t, int32(p), r.Brokers[0].Port)
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
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{
					Brokers: []string{b.Listener.Addr().String()},
					Topics:  []kafkatest.TopicConfig{{"foo", 2}, {"foo2", 1}}}))

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
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{
					Brokers: []string{b.Listener.Addr().String()},
					Topics:  []kafkatest.TopicConfig{{"foo", 1}}}))
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

package kafka_test

import (
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/offset"
	"mokapi/test"
	"testing"
)

func TestOffsetsFetch(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, b *kafkatest.Broker)
	}{
		{
			"empty earliest",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{
					Brokers: []string{b.Listener.Addr().String()},
					Topics:  []kafkatest.TopicConfig{{"foo", 1}}}))

				r, err := b.Client().Offset(3, &offset.Request{Topics: []offset.RequestTopic{
					{
						Name: "foo",
						Partitions: []offset.RequestPartition{
							{
								Index:     0,
								Timestamp: protocol.Earliest,
							},
						},
					},
				}})
				test.Ok(t, err)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, 1, len(r.Topics[0].Partitions))

				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.None, p.ErrorCode)
				test.Equals(t, int64(-1), p.Offset)
			},
		},
		{
			"empty latest",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{
					Brokers: []string{b.Listener.Addr().String()},
					Topics:  []kafkatest.TopicConfig{{"foo", 1}}}))

				r, err := b.Client().Offset(3, &offset.Request{Topics: []offset.RequestTopic{
					{
						Name: "foo",
						Partitions: []offset.RequestPartition{
							{
								Index:     0,
								Timestamp: protocol.Latest,
							},
						},
					},
				}})
				test.Ok(t, err)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, 1, len(r.Topics[0].Partitions))

				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.None, p.ErrorCode)
				test.Equals(t, int64(-1), p.Offset)
			},
		},
		{
			"one record earliest",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{
					Brokers: []string{b.Listener.Addr().String()},
					Topics:  []kafkatest.TopicConfig{{"foo", 1}}}))

				b.Store().Topic("foo").Partition(0).Write(protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Key:   protocol.NewBytes([]byte("foo")),
							Value: protocol.NewBytes([]byte("bar")),
						},
					},
				})

				r, err := b.Client().Offset(3, &offset.Request{Topics: []offset.RequestTopic{
					{
						Name: "foo",
						Partitions: []offset.RequestPartition{
							{
								Index:     0,
								Timestamp: protocol.Earliest,
							},
						},
					},
				}})

				test.Ok(t, err)
				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.None, p.ErrorCode)
				test.Equals(t, protocol.Earliest, p.Timestamp)
				test.Equals(t, int64(0), p.Offset)
			},
		},
		{
			"one record latest",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{
					Brokers: []string{b.Listener.Addr().String()},
					Topics:  []kafkatest.TopicConfig{{"foo", 1}}}))

				b.Store().Topic("foo").Partition(0).Write(protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Key:   protocol.NewBytes([]byte("foo")),
							Value: protocol.NewBytes([]byte("bar")),
						},
					},
				})

				r, err := b.Client().Offset(3, &offset.Request{Topics: []offset.RequestTopic{
					{
						Name: "foo",
						Partitions: []offset.RequestPartition{
							{
								Index:     0,
								Timestamp: protocol.Latest,
							},
						},
					},
				}})

				test.Ok(t, err)
				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.None, p.ErrorCode)
				test.Equals(t, protocol.Latest, p.Timestamp)
				test.Equals(t, int64(0), p.Offset)
			},
		},
		{
			"topic not exists",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{
					Brokers: []string{b.Listener.Addr().String()},
					Topics:  []kafkatest.TopicConfig{{"foo", 1}}}))

				r, err := b.Client().Offset(3, &offset.Request{Topics: []offset.RequestTopic{
					{
						Name: "foo",
						Partitions: []offset.RequestPartition{
							{
								Index:     1,
								Timestamp: protocol.Latest,
							},
						},
					},
				}})

				test.Ok(t, err)
				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.UnknownTopicOrPartition, p.ErrorCode)
			},
		},
		{
			"partition not exists",
			func(t *testing.T, b *kafkatest.Broker) {
				r, err := b.Client().Offset(3, &offset.Request{Topics: []offset.RequestTopic{
					{
						Name: "foo",
						Partitions: []offset.RequestPartition{
							{
								Index:     0,
								Timestamp: protocol.Latest,
							},
						},
					},
				}})

				test.Ok(t, err)
				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.UnknownTopicOrPartition, p.ErrorCode)
			},
		},
	}

	t.Parallel()
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

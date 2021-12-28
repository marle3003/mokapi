package kafka_test

import (
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/offsetCommit"
	"mokapi/test"
	"testing"
)

func TestOffsetCommit(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, b *kafkatest.Broker)
	}{
		{
			"group not exists",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{
					Brokers: []string{b.Listener.Addr().String()},
					Topics:  []kafkatest.TopicConfig{{"foo", 1}}}))

				r, err := b.Client().OffsetCommit(2, &offsetCommit.Request{
					GroupId:       "TestGroup",
					GenerationId:  0,
					MemberId:      "foo",
					RetentionTime: 0,
					Topics: []offsetCommit.Topic{
						{
							Name: "foo",
							Partitions: []offsetCommit.Partition{
								{
									Index:  0,
									Offset: 0,
								},
							},
						},
					},
				})
				test.Ok(t, err)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, "foo", r.Topics[0].Name)
				test.Equals(t, 1, len(r.Topics[0].Partitions))
				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.UnknownMemberId, p.ErrorCode)
			},
		},
		{
			"offset out of range",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{
					Brokers: []string{b.Listener.Addr().String()},
					Topics:  []kafkatest.TopicConfig{{"foo", 1}}}))

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)
				test.Ok(t, err)

				r, err := b.Client().OffsetCommit(2, &offsetCommit.Request{
					GroupId:       "bar",
					GenerationId:  0,
					MemberId:      "foo",
					RetentionTime: 0,
					Topics: []offsetCommit.Topic{
						{
							Name: "foo",
							Partitions: []offsetCommit.Partition{
								{
									Index:    0,
									Offset:   99999,
									Metadata: "",
								},
							},
						},
					},
				})
				test.Ok(t, err)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, "foo", r.Topics[0].Name)
				test.Equals(t, 1, len(r.Topics[0].Partitions))
				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.OffsetOutOfRange, p.ErrorCode)
			},
		},
		{
			"offset commit successfully",
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

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)
				test.Ok(t, err)

				r, err := b.Client().OffsetCommit(2, &offsetCommit.Request{
					GroupId:       "bar",
					GenerationId:  0,
					MemberId:      "foo",
					RetentionTime: 0,
					Topics: []offsetCommit.Topic{
						{
							Name: "foo",
							Partitions: []offsetCommit.Partition{
								{
									Index:    0,
									Offset:   0,
									Metadata: "",
								},
							},
						},
					},
				})
				test.Ok(t, err)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, "foo", r.Topics[0].Name)
				test.Equals(t, 1, len(r.Topics[0].Partitions))
				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.None, p.ErrorCode)
			},
		},
		{
			"topic not exists",
			func(t *testing.T, b *kafkatest.Broker) {
				r, err := b.Client().OffsetCommit(2, &offsetCommit.Request{
					GroupId:       "TestGroup",
					GenerationId:  0,
					MemberId:      "foo",
					RetentionTime: 0,
					Topics: []offsetCommit.Topic{
						{
							Name: "foo",
							Partitions: []offsetCommit.Partition{
								{
									Index:  0,
									Offset: 0,
								},
							},
						},
					},
				})
				test.Ok(t, err)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, "foo", r.Topics[0].Name)
				test.Equals(t, 1, len(r.Topics[0].Partitions))
				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.UnknownTopicOrPartition, p.ErrorCode)
			},
		},
		{
			"partition not exists",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{
					Brokers: []string{b.Listener.Addr().String()},
					Topics:  []kafkatest.TopicConfig{{"foo", 0}}}))

				r, err := b.Client().OffsetCommit(2, &offsetCommit.Request{
					GroupId:       "TestGroup",
					GenerationId:  0,
					MemberId:      "foo",
					RetentionTime: 0,
					Topics: []offsetCommit.Topic{
						{
							Name: "foo",
							Partitions: []offsetCommit.Partition{
								{
									Index:  0,
									Offset: 0,
								},
							},
						},
					},
				})
				test.Ok(t, err)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, "foo", r.Topics[0].Name)
				test.Equals(t, 1, len(r.Topics[0].Partitions))
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

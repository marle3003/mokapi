package kafka_test

import (
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/offsetCommit"
	"mokapi/kafka/protocol/offsetFetch"
	"mokapi/kafka/schema"
	"mokapi/kafka/store"
	"mokapi/test"
	"testing"
)

func TestOffsetFetch(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, b *kafkatest.Broker)
	}{
		{
			"empty",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(store.New(schema.Cluster{Topics: []schema.Topic{{Name: "foo", Partitions: []schema.Partition{{Index: 0}}}}}))

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)

				r, err := b.Client().OffsetFetch(3, &offsetFetch.Request{
					GroupId: "bar",
					Topics: []offsetFetch.RequestTopic{
						{
							Name:             "foo",
							PartitionIndexes: []int32{0},
						},
					}})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, 1, len(r.Topics[0].Partitions))

				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.None, p.ErrorCode)
				test.Equals(t, int64(-1), p.CommittedOffset)
			},
		},
		{
			"empty with api version 0",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(store.New(schema.Cluster{Topics: []schema.Topic{{Name: "foo", Partitions: []schema.Partition{{Index: 0}}}}}))

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)

				r, err := b.Client().OffsetFetch(0, &offsetFetch.Request{
					GroupId: "bar",
					Topics: []offsetFetch.RequestTopic{
						{
							Name:             "foo",
							PartitionIndexes: []int32{0},
						},
					}})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, 1, len(r.Topics[0].Partitions))

				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.UnknownTopicOrPartition, p.ErrorCode)
				test.Equals(t, int64(-1), p.CommittedOffset)
			},
		},
		{
			"invalid partition request",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(store.New(schema.Cluster{Topics: []schema.Topic{{Name: "foo", Partitions: []schema.Partition{{Index: 0}}}}}))

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)

				r, err := b.Client().OffsetFetch(3, &offsetFetch.Request{Topics: []offsetFetch.RequestTopic{
					{
						Name:             "foo",
						PartitionIndexes: []int32{9999},
					},
				}})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, 1, len(r.Topics[0].Partitions))

				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.UnknownTopicOrPartition, p.ErrorCode)
				test.Equals(t, int64(-1), p.CommittedOffset)

			},
		},
		{
			"invalid partition request with api version 0",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(store.New(schema.Cluster{Topics: []schema.Topic{{Name: "foo", Partitions: []schema.Partition{{Index: 0}}}}}))

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)

				r, err := b.Client().OffsetFetch(0, &offsetFetch.Request{Topics: []offsetFetch.RequestTopic{
					{
						Name:             "foo",
						PartitionIndexes: []int32{9999},
					},
				}})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, 1, len(r.Topics[0].Partitions))

				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.UnknownTopicOrPartition, p.ErrorCode)
				test.Equals(t, int64(-1), p.CommittedOffset)

			},
		},
		{
			"unknown topic request",
			func(t *testing.T, b *kafkatest.Broker) {
				r, err := b.Client().OffsetFetch(3, &offsetFetch.Request{Topics: []offsetFetch.RequestTopic{
					{
						Name:             "unknown",
						PartitionIndexes: []int32{9999},
					},
				}})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, 1, len(r.Topics[0].Partitions))

				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.UnknownTopicOrPartition, p.ErrorCode)
				test.Equals(t, int64(-1), p.CommittedOffset)
			},
		},
		{
			"unknown member",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(store.New(schema.Cluster{Topics: []schema.Topic{{Name: "foo", Partitions: []schema.Partition{{Index: 0}}}}}))

				r, err := b.Client().OffsetFetch(3, &offsetFetch.Request{
					GroupId: "bar",
					Topics: []offsetFetch.RequestTopic{
						{
							Name:             "foo",
							PartitionIndexes: []int32{0},
						},
					}})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, 1, len(r.Topics[0].Partitions))

				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.UnknownMemberId, p.ErrorCode)
				test.Equals(t, int64(-1), p.CommittedOffset)
			},
		},
		{
			"offset fetch",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(store.New(schema.Cluster{Topics: []schema.Topic{{Name: "foo", Partitions: []schema.Partition{{Index: 0}}}}}))
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

				_, err = b.Client().OffsetCommit(2, &offsetCommit.Request{
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

				r, err := b.Client().OffsetFetch(3, &offsetFetch.Request{
					GroupId: "bar",
					Topics: []offsetFetch.RequestTopic{
						{
							Name:             "foo",
							PartitionIndexes: []int32{0},
						},
					}})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, 1, len(r.Topics[0].Partitions))

				p := r.Topics[0].Partitions[0]
				test.Equals(t, protocol.None, p.ErrorCode)
				test.Equals(t, int64(0), p.CommittedOffset)
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

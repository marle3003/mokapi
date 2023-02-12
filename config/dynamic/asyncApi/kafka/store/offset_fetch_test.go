package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/offsetCommit"
	"mokapi/kafka/offsetFetch"
	"testing"
)

func TestOffsetFetch(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, s *store.Store)
	}{
		{
			"empty",
			func(t *testing.T, s *store.Store) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				defer b.Close()
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithServer("", "kafka", b.Addr),
					asyncapitest.WithChannel("foo"),
				))

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)

				r, err := b.Client().OffsetFetch(3, &offsetFetch.Request{
					GroupId: "bar",
					Topics: []offsetFetch.RequestTopic{
						{
							Name:             "foo",
							PartitionIndexes: []int32{0},
						},
					}})
				require.NoError(t, err)
				require.Equal(t, kafka.None, r.ErrorCode)
				require.Len(t, r.Topics, 1)
				require.Len(t, r.Topics[0].Partitions, 1)

				p := r.Topics[0].Partitions[0]
				require.Equal(t, kafka.None, p.ErrorCode)
				require.Equal(t, int64(-1), p.CommittedOffset)
			},
		},
		{
			"empty with api version 0",
			func(t *testing.T, s *store.Store) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				defer b.Close()
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithServer("", "kafka", b.Addr),
					asyncapitest.WithChannel("foo"),
				))

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)

				r, err := b.Client().OffsetFetch(0, &offsetFetch.Request{
					GroupId: "bar",
					Topics: []offsetFetch.RequestTopic{
						{
							Name:             "foo",
							PartitionIndexes: []int32{0},
						},
					}})
				require.NoError(t, err)
				require.Equal(t, kafka.None, r.ErrorCode)
				require.Len(t, r.Topics, 1)
				require.Len(t, r.Topics[0].Partitions, 1)

				p := r.Topics[0].Partitions[0]
				require.Equal(t, kafka.UnknownTopicOrPartition, p.ErrorCode)
				require.Equal(t, int64(-1), p.CommittedOffset)
			},
		},
		{
			"invalid partition request",
			func(t *testing.T, s *store.Store) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				defer b.Close()
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithServer("", "kafka", b.Addr),
					asyncapitest.WithChannel("foo"),
				))

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)

				r, err := b.Client().OffsetFetch(3, &offsetFetch.Request{
					Topics: []offsetFetch.RequestTopic{
						{
							Name:             "foo",
							PartitionIndexes: []int32{9999},
						},
					}})
				require.NoError(t, err)
				require.Equal(t, kafka.None, r.ErrorCode)
				require.Len(t, r.Topics, 1)
				require.Len(t, r.Topics[0].Partitions, 1)

				p := r.Topics[0].Partitions[0]
				require.Equal(t, kafka.UnknownTopicOrPartition, p.ErrorCode)
				require.Equal(t, int64(-1), p.CommittedOffset)

			},
		},
		{
			"invalid partition request with api version 0",
			func(t *testing.T, s *store.Store) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				defer b.Close()
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithServer("", "kafka", b.Addr),
					asyncapitest.WithChannel("foo"),
				))

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)

				r, err := b.Client().OffsetFetch(0, &offsetFetch.Request{Topics: []offsetFetch.RequestTopic{
					{
						Name:             "foo",
						PartitionIndexes: []int32{9999},
					},
				}})
				require.NoError(t, err)
				require.Equal(t, kafka.None, r.ErrorCode)
				require.Len(t, r.Topics, 1)
				require.Len(t, r.Topics[0].Partitions, 1)

				p := r.Topics[0].Partitions[0]
				require.Equal(t, kafka.UnknownTopicOrPartition, p.ErrorCode)
				require.Equal(t, int64(-1), p.CommittedOffset)

			},
		},
		{
			"unknown topic request",
			func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &offsetFetch.Request{
					Topics: []offsetFetch.RequestTopic{
						{
							Name:             "unknown",
							PartitionIndexes: []int32{9999},
						},
					}}))

				res, ok := rr.Message.(*offsetFetch.Response)
				require.True(t, ok)
				require.Equal(t, kafka.None, res.ErrorCode)
				require.Len(t, res.Topics, 1)
				require.Len(t, res.Topics[0].Partitions, 1)

				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.UnknownTopicOrPartition, p.ErrorCode)
				require.Equal(t, int64(-1), p.CommittedOffset)
			},
		},
		{
			"unknown member",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithChannel("foo"),
				))

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &offsetFetch.Request{
					Topics: []offsetFetch.RequestTopic{
						{
							Name:             "foo",
							PartitionIndexes: []int32{0},
						},
					}}))

				res, ok := rr.Message.(*offsetFetch.Response)
				require.True(t, ok)
				require.Equal(t, kafka.None, res.ErrorCode)
				require.Len(t, res.Topics, 1)
				require.Len(t, res.Topics[0].Partitions, 1)

				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.UnknownMemberId, p.ErrorCode)
				require.Equal(t, int64(-1), p.CommittedOffset)
			},
		},
		{
			"offset fetch",
			func(t *testing.T, s *store.Store) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				defer b.Close()
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithServer("", "kafka", b.Addr),
					asyncapitest.WithChannel("foo"),
				))
				s.Topic("foo").Partition(0).Write(kafka.RecordBatch{
					Records: []kafka.Record{
						{
							Key:   kafka.NewBytes([]byte("foo")),
							Value: kafka.NewBytes([]byte("bar")),
						},
					},
				})

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)
				require.NoError(t, err)

				_, err = b.Client().OffsetCommit(2, &offsetCommit.Request{
					GroupId:  "bar",
					MemberId: "foo",
					Topics: []offsetCommit.Topic{
						{
							Name:       "foo",
							Partitions: []offsetCommit.Partition{{}},
						},
					},
				})
				require.NoError(t, err)

				r, err := b.Client().OffsetFetch(3, &offsetFetch.Request{
					GroupId: "bar",
					Topics: []offsetFetch.RequestTopic{
						{
							Name:             "foo",
							PartitionIndexes: []int32{0},
						},
					}})
				require.NoError(t, err)
				require.Equal(t, kafka.None, r.ErrorCode)
				require.Len(t, r.Topics, 1)
				require.Len(t, r.Topics[0].Partitions, 1)

				p := r.Topics[0].Partitions[0]
				require.Equal(t, kafka.None, p.ErrorCode)
				require.Equal(t, int64(0), p.CommittedOffset)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := store.New(asyncapitest.NewConfig(), enginetest.NewEngine())
			defer s.Close()
			tc.fn(t, s)
		})
	}
}

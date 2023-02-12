package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/offset"
	"testing"
)

func TestOffsets(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, s *store.Store)
	}{
		{
			"empty earliest",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithChannel("foo")))

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &offset.Request{
					Topics: []offset.RequestTopic{
						{
							Name: "foo",
							Partitions: []offset.RequestPartition{
								{
									Timestamp: kafka.Earliest,
								},
							},
						},
					}}))

				res, ok := rr.Message.(*offset.Response)
				require.True(t, ok)
				require.Len(t, res.Topics, 1)
				require.Len(t, res.Topics[0].Partitions, 1)

				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.None, p.ErrorCode)
				require.Equal(t, int64(0), p.Offset)
			},
		},
		{
			"empty latest",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithChannel("foo")))

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &offset.Request{
					Topics: []offset.RequestTopic{
						{
							Name: "foo",
							Partitions: []offset.RequestPartition{
								{
									Timestamp: kafka.Latest,
								},
							},
						},
					}}))

				res, ok := rr.Message.(*offset.Response)
				require.True(t, ok)
				require.Len(t, res.Topics, 1)
				require.Len(t, res.Topics[0].Partitions, 1)

				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.None, p.ErrorCode)
				require.Equal(t, int64(0), p.Offset)
			},
		},
		{
			"one record earliest",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithChannel("foo")))
				s.Topic("foo").Partition(0).Write(kafka.RecordBatch{
					Records: []kafka.Record{
						{
							Key:   kafka.NewBytes([]byte("foo")),
							Value: kafka.NewBytes([]byte("bar")),
						},
					},
				})

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &offset.Request{
					Topics: []offset.RequestTopic{
						{
							Name: "foo",
							Partitions: []offset.RequestPartition{
								{
									Timestamp: kafka.Earliest,
								},
							},
						},
					}}))

				res, ok := rr.Message.(*offset.Response)
				require.True(t, ok)
				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.None, p.ErrorCode)
				require.Equal(t, kafka.Earliest, p.Timestamp)
				require.Equal(t, int64(0), p.Offset)
			},
		},
		{
			"one record latest",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithChannel("foo")))
				s.Topic("foo").Partition(0).Write(kafka.RecordBatch{
					Records: []kafka.Record{
						{
							Key:   kafka.NewBytes([]byte("foo")),
							Value: kafka.NewBytes([]byte("bar")),
						},
					},
				})

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &offset.Request{
					Topics: []offset.RequestTopic{
						{
							Name: "foo",
							Partitions: []offset.RequestPartition{
								{
									Timestamp: kafka.Latest,
								},
							},
						},
					}}))

				res, ok := rr.Message.(*offset.Response)
				require.True(t, ok)
				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.None, p.ErrorCode)
				require.Equal(t, kafka.Latest, p.Timestamp)
				require.Equal(t, int64(1), p.Offset)
			},
		},
		{
			"topic not exists",
			func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &offset.Request{
					Topics: []offset.RequestTopic{
						{
							Name: "foo",
							Partitions: []offset.RequestPartition{
								{
									Timestamp: kafka.Latest,
								},
							},
						},
					}}))

				res, ok := rr.Message.(*offset.Response)
				require.True(t, ok)
				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.UnknownTopicOrPartition, p.ErrorCode)
			},
		},
		{
			"partition not exists",
			func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &offset.Request{
					Topics: []offset.RequestTopic{
						{
							Name: "foo",
							Partitions: []offset.RequestPartition{
								{
									Timestamp: kafka.Latest,
								},
							},
						},
					}}))

				res, ok := rr.Message.(*offset.Response)
				require.True(t, ok)
				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.UnknownTopicOrPartition, p.ErrorCode)
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

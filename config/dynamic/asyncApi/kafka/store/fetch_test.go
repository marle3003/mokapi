package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	kafka2 "mokapi/config/dynamic/asyncApi/kafka"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/fetch"
	"mokapi/kafka/kafkatest"
	"testing"
	"time"
)

func TestFetch(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, s *store.Store)
	}{
		{
			"topic not exists",
			func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name:       "foo",
						Partitions: []fetch.RequestPartition{{}},
					},
				}}))

				res, ok := rr.Message.(*fetch.Response)
				require.True(t, ok)
				require.Equal(t, kafka.None, res.ErrorCode)
				require.Equal(t, 1, len(res.Topics))
				require.Equal(t, "foo", res.Topics[0].Name)
				require.Len(t, res.Topics[0].Partitions, 1)
				require.Equal(t, kafka.UnknownTopicOrPartition, res.Topics[0].Partitions[0].ErrorCode)
			},
		},
		{
			"partition not exists",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(asyncapitest.WithChannel("foo")))

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name:       "bar",
						Partitions: []fetch.RequestPartition{{}},
					},
				}}))

				res, ok := rr.Message.(*fetch.Response)
				require.True(t, ok)
				require.Equal(t, kafka.None, res.ErrorCode)
				require.Equal(t, 1, len(res.Topics))
				require.Equal(t, "bar", res.Topics[0].Name)
				require.Len(t, res.Topics[0].Partitions, 1)
				require.Equal(t, kafka.UnknownTopicOrPartition, res.Topics[0].Partitions[0].ErrorCode)
			},
		},
		{
			"empty",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(asyncapitest.WithChannel("foo", asyncapitest.WithChannelKafka(kafka2.TopicBindings{Partitions: 1}))))

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name:       "foo",
						Partitions: []fetch.RequestPartition{{}},
					},
				}}))

				res, ok := rr.Message.(*fetch.Response)
				require.True(t, ok)
				require.Equal(t, kafka.None, res.ErrorCode)
				require.Len(t, res.Topics, 1)
				require.Equal(t, "foo", res.Topics[0].Name)
				require.Len(t, res.Topics[0].Partitions, 1)
				require.Equal(t, kafka.None, res.Topics[0].Partitions[0].ErrorCode)
				require.Len(t, res.Topics[0].Partitions[0].RecordSet.Records, 0)
			},
		},
		{
			"empty with max wait time",
			func(t *testing.T, s *store.Store) {
				start := time.Now()

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name:       "foo",
						Partitions: []fetch.RequestPartition{{}},
					},
				}, MaxWaitMs: 1000}))
				end := time.Now()

				_, ok := rr.Message.(*fetch.Response)
				require.True(t, ok)
				waitTime := end.Sub(start).Milliseconds()
				// fetch request waits for MaxWaitMs - 200ms
				require.Less(t, waitTime, int64(100), "wait time should be 800ms but was %v", waitTime)
			},
		},
		{
			"empty with max wait time and min bytes",
			func(t *testing.T, s *store.Store) {
				start := time.Now()

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name:       "foo",
						Partitions: []fetch.RequestPartition{{}},
					},
				}, MaxWaitMs: 1000, MinBytes: 1}))
				end := time.Now()

				_, ok := rr.Message.(*fetch.Response)
				require.True(t, ok)
				waitTime := end.Sub(start).Milliseconds()
				require.Greater(t, waitTime, int64(799))
				require.Less(t, waitTime, int64(1000), "wait time should be 800ms but was %v", waitTime)
			},
		},
		{
			"fetch one record",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(asyncapitest.WithChannel("foo")))
				_, err := s.Topic("foo").Partition(0).Write(kafka.RecordBatch{Records: []kafka.Record{
					{
						Key:   kafka.NewBytes([]byte("foo")),
						Value: kafka.NewBytes([]byte("bar")),
					},
				}})
				require.NoError(t, err)

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &fetch.Request{
					MaxBytes: 1000,
					Topics: []fetch.Topic{
						{
							Name:       "foo",
							Partitions: []fetch.RequestPartition{{MaxBytes: 1000}},
						},
					}}))

				res, ok := rr.Message.(*fetch.Response)
				require.True(t, ok)
				require.Equal(t, 1, len(res.Topics[0].Partitions[0].RecordSet.Records))
				require.Equal(t, int64(1), res.Topics[0].Partitions[0].HighWatermark)
				require.Equal(t, int64(1), res.Topics[0].Partitions[0].LastStableOffset)

				record := res.Topics[0].Partitions[0].RecordSet.Records[0]
				require.Equal(t, int64(0), record.Offset)
				require.Equal(t, "foo", kafkatest.BytesToString(record.Key))
				require.Equal(t, "bar", kafkatest.BytesToString(record.Value))
			},
		},
		{
			"fetch one record with MaxBytes 15",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(asyncapitest.WithChannel("foo")))
				_, err := s.Topic("foo").Partition(0).Write(kafka.RecordBatch{Records: []kafka.Record{
					{
						Key:   kafka.NewBytes([]byte("key-1")),
						Value: kafka.NewBytes([]byte("value-1")),
					},
					{
						Key:   kafka.NewBytes([]byte("key-2")),
						Value: kafka.NewBytes([]byte("value-2")),
					},
				}})
				require.NoError(t, err)

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &fetch.Request{
					MaxBytes: 1000,
					Topics: []fetch.Topic{
						{
							Name:       "foo",
							Partitions: []fetch.RequestPartition{{MaxBytes: 15}},
						},
					}}))

				res, ok := rr.Message.(*fetch.Response)
				require.True(t, ok)
				// only one record returned because of MaxBytes 1
				require.Len(t, res.Topics[0].Partitions[0].RecordSet.Records, 1)
				require.Equal(t, int64(2), res.Topics[0].Partitions[0].HighWatermark)

				record := res.Topics[0].Partitions[0].RecordSet.Records[0]
				require.Equal(t, int64(0), record.Offset)
				require.Equal(t, "key-1", kafkatest.BytesToString(record.Key))
				require.Equal(t, "value-1", kafkatest.BytesToString(record.Value))
			},
		},
		{
			"fetch next not available record",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(asyncapitest.WithChannel("foo")))
				_, err := s.Topic("foo").Partition(0).Write(kafka.RecordBatch{Records: []kafka.Record{
					{
						Key:   kafka.NewBytes([]byte("foo")),
						Value: kafka.NewBytes([]byte("bar")),
					},
				}})
				require.NoError(t, err)

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name:       "foo",
						Partitions: []fetch.RequestPartition{{MaxBytes: 1, FetchOffset: 1}},
					},
				}}))

				res, ok := rr.Message.(*fetch.Response)
				require.True(t, ok)
				require.Equal(t, 0, len(res.Topics[0].Partitions[0].RecordSet.Records))
				require.Equal(t, int64(1), res.Topics[0].Partitions[0].HighWatermark)
			},
		},
		{
			"fetch both records",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(asyncapitest.WithChannel("foo")))
				_, err := s.Topic("foo").Partition(0).Write(kafka.RecordBatch{Records: []kafka.Record{
					{
						Key:   kafka.NewBytes([]byte("key-1")),
						Value: kafka.NewBytes([]byte("value-1")),
					},
					{
						Key:   kafka.NewBytes([]byte("key-2")),
						Value: kafka.NewBytes([]byte("value-2")),
					},
				}})
				require.NoError(t, err)

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &fetch.Request{
					MaxBytes: 1000,
					Topics: []fetch.Topic{
						{
							Name:       "foo",
							Partitions: []fetch.RequestPartition{{MaxBytes: 24}},
						},
					}}))

				res, ok := rr.Message.(*fetch.Response)
				require.True(t, ok)
				require.Equal(t, 2, len(res.Topics[0].Partitions[0].RecordSet.Records))
				require.Equal(t, int64(2), res.Topics[0].Partitions[0].HighWatermark)

				record1 := res.Topics[0].Partitions[0].RecordSet.Records[0]
				require.Equal(t, int64(0), record1.Offset)
				require.Equal(t, "key-1", kafkatest.BytesToString(record1.Key))
				require.Equal(t, "value-1", kafkatest.BytesToString(record1.Value))

				record2 := res.Topics[0].Partitions[0].RecordSet.Records[1]
				require.Equal(t, int64(1), record2.Offset)
				require.Equal(t, "key-2", kafkatest.BytesToString(record2.Key))
				require.Equal(t, "value-2", kafkatest.BytesToString(record2.Value))
			},
		},
		{
			"wait fetch for MinBytes",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(asyncapitest.WithChannel("foo")))

				ch := make(chan *fetch.Response, 1)
				go func() {
					rr := kafkatest.NewRecorder()
					s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &fetch.Request{
						MaxBytes: 1000,
						Topics: []fetch.Topic{
							{
								Name:       "foo",
								Partitions: []fetch.RequestPartition{{MaxBytes: 12}},
							},
						}, MinBytes: 1, MaxWaitMs: 5000}))
					res, ok := rr.Message.(*fetch.Response)
					require.True(t, ok)
					ch <- res
				}()
				time.Sleep(300 * time.Millisecond)
				_, err := s.Topic("foo").Partition(0).Write(kafka.RecordBatch{
					[]kafka.Record{
						{
							Key:   kafka.NewBytes([]byte("foo")),
							Value: kafka.NewBytes([]byte("bar")),
						},
					},
				})
				require.NoError(t, err)

				r := <-ch

				require.Equal(t, 1, len(r.Topics[0].Partitions[0].RecordSet.Records))
				require.Equal(t, int64(1), r.Topics[0].Partitions[0].HighWatermark)

				record := r.Topics[0].Partitions[0].RecordSet.Records[0]
				require.Equal(t, int64(0), record.Offset)
				require.Equal(t, "foo", kafkatest.BytesToString(record.Key))
				require.Equal(t, "bar", kafkatest.BytesToString(record.Value))
			},
		},
		{
			"fetch offset out of range when empty",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(asyncapitest.WithChannel("foo")))

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name:       "foo",
						Partitions: []fetch.RequestPartition{{FetchOffset: 1}},
					},
				}}))

				res, ok := rr.Message.(*fetch.Response)
				require.True(t, ok)
				require.Equal(t, kafka.None, res.ErrorCode)
				require.Equal(t, kafka.None, res.Topics[0].Partitions[0].ErrorCode)
			},
		},
		{
			"fetch offset out of range",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(asyncapitest.WithChannel("foo")))
				_, err := s.Topic("foo").Partition(0).Write(kafka.RecordBatch{Records: []kafka.Record{
					{
						Key:   kafka.NewBytes([]byte("foo")),
						Value: kafka.NewBytes([]byte("bar")),
					},
				}})
				require.NoError(t, err)

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name:       "foo",
						Partitions: []fetch.RequestPartition{{FetchOffset: -10}},
					},
				}}))

				res, ok := rr.Message.(*fetch.Response)
				require.True(t, ok)
				require.Equal(t, kafka.None, res.ErrorCode)
				require.Equal(t, kafka.OffsetOutOfRange, res.Topics[0].Partitions[0].ErrorCode)
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

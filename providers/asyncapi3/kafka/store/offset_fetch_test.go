package store_test

import (
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/offsetCommit"
	"mokapi/kafka/offsetFetch"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/schema/json/schema/schematest"
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
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithServer("", "kafka", b.Addr),
					asyncapi3test.WithChannel("foo"),
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
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithServer("", "kafka", b.Addr),
					asyncapi3test.WithChannel("foo"),
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
				require.Equal(t, kafka.None, p.ErrorCode)
				require.Equal(t, int64(-1), p.CommittedOffset)
			},
		},
		{
			"invalid partition request",
			func(t *testing.T, s *store.Store) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				defer b.Close()
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithServer("", "kafka", b.Addr),
					asyncapi3test.WithChannel("foo"),
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
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithServer("", "kafka", b.Addr),
					asyncapi3test.WithChannel("foo"),
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
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithChannel("foo"),
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
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithServer("", "kafka", b.Addr),
					asyncapi3test.WithChannel("foo"),
				))
				s.Topic("foo").Partition(0).Write(kafka.RecordBatch{
					Records: []*kafka.Record{
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

			s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngine())
			defer s.Close()
			tc.fn(t, s)
		})
	}
}

func TestOffsetFetch_Validation(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, s *store.Store, hook *test.Hook)
	}{
		{
			"invalid clientId",
			func(t *testing.T, s *store.Store, hook *test.Hook) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				defer b.Close()

				ch := asyncapi3test.NewChannel()
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithServer("", "kafka", b.Addr),
					asyncapi3test.AddChannel("foo", ch),
					asyncapi3test.WithOperation("foo",
						asyncapi3test.WithOperationAction("receive"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.WithOperationBinding(asyncapi3.KafkaOperationBinding{ClientId: schematest.New("string", schematest.WithPattern("^[A-Z]{10}[0-5]$"))}),
					)))

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)
				require.NoError(t, err)

				res, err := b.Client().OffsetFetch(3, &offsetFetch.Request{
					GroupId: "bar",
					Topics: []offsetFetch.RequestTopic{
						{
							Name:             "foo",
							PartitionIndexes: []int32{0},
						},
					}})

				require.Equal(t, kafka.None, res.ErrorCode)
				require.Len(t, res.Topics, 1)
				require.Len(t, res.Topics[0].Partitions, 1)

				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.UnknownServerError, p.ErrorCode)
				require.Equal(t, int64(-1), p.CommittedOffset)

				require.Equal(t, 6, len(hook.Entries))
				require.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
				require.Equal(t, "kafka OffsetFetch: invalid consumer 'kafkatest' for topic foo: invalid clientId: error count 1:\n- #/pattern: string 'kafkatest' does not match regex pattern '^[A-Z]{10}[0-5]$'", hook.LastEntry().Message)
			},
		},
		{
			"invalid groupId",
			func(t *testing.T, s *store.Store, hook *test.Hook) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				defer b.Close()

				ch := asyncapi3test.NewChannel()
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithServer("", "kafka", b.Addr),
					asyncapi3test.AddChannel("foo", ch),
					asyncapi3test.WithOperation("foo",
						asyncapi3test.WithOperationAction("receive"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.WithOperationBinding(asyncapi3.KafkaOperationBinding{GroupId: schematest.New("string", schematest.WithPattern("^[A-Z]{10}[0-5]$"))}),
					)))

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)
				require.NoError(t, err)

				res, err := b.Client().OffsetFetch(3, &offsetFetch.Request{
					GroupId: "bar",
					Topics: []offsetFetch.RequestTopic{
						{
							Name:             "foo",
							PartitionIndexes: []int32{0},
						},
					}})

				require.Equal(t, kafka.None, res.ErrorCode)
				require.Len(t, res.Topics, 1)
				require.Len(t, res.Topics[0].Partitions, 1)

				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.InvalidGroupId, p.ErrorCode)
				require.Equal(t, int64(-1), p.CommittedOffset)

				require.Equal(t, 6, len(hook.Entries))
				require.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
				require.Equal(t, "kafka OffsetFetch: invalid consumer 'kafkatest' for topic foo: invalid groupId: error count 1:\n- #/pattern: string 'bar' does not match regex pattern '^[A-Z]{10}[0-5]$'", hook.LastEntry().Message)
			},
		},
		{
			"valid groupId and clientId",
			func(t *testing.T, s *store.Store, hook *test.Hook) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				c := kafkatest.NewClient(b.Addr, "MOKAPITEST1")
				defer b.Close()

				ch := asyncapi3test.NewChannel()
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithServer("", "kafka", b.Addr),
					asyncapi3test.AddChannel("foo", ch),
					asyncapi3test.WithOperation("foo",
						asyncapi3test.WithOperationAction("receive"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.WithOperationBinding(asyncapi3.KafkaOperationBinding{
							ClientId: schematest.New("string", schematest.WithPattern("^[A-Z]{10}[0-5]$")),
							GroupId:  schematest.New("string", schematest.WithPattern("^[A-Z]{5}[0-5]$")),
						}),
					)))

				err := c.JoinSyncGroup("foo", "GROUP1", 3, 3)
				require.NoError(t, err)

				res, err := c.OffsetFetch(3, &offsetFetch.Request{
					GroupId: "GROUP1",
					Topics: []offsetFetch.RequestTopic{
						{
							Name:             "foo",
							PartitionIndexes: []int32{0},
						},
					}})

				require.Equal(t, kafka.None, res.ErrorCode)
				require.Len(t, res.Topics, 1)
				require.Len(t, res.Topics[0].Partitions, 1)

				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.None, p.ErrorCode)
				require.Equal(t, int64(-1), p.CommittedOffset)

				require.Equal(t, 5, len(hook.Entries))
				require.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngine())
			defer s.Close()
			hook := test.NewGlobal()
			tc.fn(t, s, hook)
		})
	}
}

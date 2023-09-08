package store_test

import (
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	kafka2 "mokapi/config/dynamic/asyncApi/kafka"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/offsetCommit"
	"testing"
)

func TestOffsetCommit(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, s *store.Store)
	}{
		{
			"group not exists",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithChannel("foo")))

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 2,
					&offsetCommit.Request{
						GroupId:  "TestGroup",
						MemberId: "foo",
						Topics: []offsetCommit.Topic{
							{
								Name:       "foo",
								Partitions: []offsetCommit.Partition{{}},
							},
						},
					},
				))

				res, ok := rr.Message.(*offsetCommit.Response)
				require.True(t, ok)
				require.Len(t, res.Topics, 1)
				require.Equal(t, "foo", res.Topics[0].Name)
				require.Len(t, res.Topics[0].Partitions, 1)
				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.UnknownMemberId, p.ErrorCode)
			},
		},
		{
			"offset out of range",
			func(t *testing.T, s *store.Store) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				defer b.Close()
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithServer("", "kafka", b.Addr),
					asyncapitest.WithChannel("foo")),
				)

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)
				require.NoError(t, err)

				r, err := b.Client().OffsetCommit(2, &offsetCommit.Request{
					GroupId:  "bar",
					MemberId: "foo",
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
				require.NoError(t, err)
				require.Len(t, r.Topics, 1)
				require.Equal(t, "foo", r.Topics[0].Name)
				require.Len(t, r.Topics[0].Partitions, 1)
				p := r.Topics[0].Partitions[0]
				require.Equal(t, kafka.OffsetOutOfRange, p.ErrorCode)
			},
		},
		{
			"offset commit successfully",
			func(t *testing.T, s *store.Store) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				defer b.Close()
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithServer("", "kafka", b.Addr),
					asyncapitest.WithChannel("foo")))
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

				r, err := b.Client().OffsetCommit(2, &offsetCommit.Request{
					GroupId:  "bar",
					MemberId: "foo",
					Topics: []offsetCommit.Topic{
						{
							Name: "foo",
							Partitions: []offsetCommit.Partition{
								{},
							},
						},
					},
				})
				require.NoError(t, err)
				require.Len(t, r.Topics, 1)
				require.Equal(t, "foo", r.Topics[0].Name)
				require.Len(t, r.Topics[0].Partitions, 1)
				p := r.Topics[0].Partitions[0]
				require.Equal(t, kafka.None, p.ErrorCode)
			},
		},
		{
			"topic not exists",
			func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 2, &offsetCommit.Request{
					GroupId:  "TestGroup",
					MemberId: "foo",
					Topics: []offsetCommit.Topic{
						{
							Name:       "foo",
							Partitions: []offsetCommit.Partition{{}},
						},
					},
				}))

				res, ok := rr.Message.(*offsetCommit.Response)
				require.True(t, ok)
				require.Len(t, res.Topics, 1)
				require.Equal(t, "foo", res.Topics[0].Name)
				require.Len(t, res.Topics[0].Partitions, 1)
				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.UnknownTopicOrPartition, p.ErrorCode)
			},
		},
		{
			"partition not exists",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithChannel("foo", asyncapitest.WithChannelKafka(kafka2.TopicBindings{Partitions: 1}))))

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 2, &offsetCommit.Request{
					GroupId:  "TestGroup",
					MemberId: "foo",
					Topics: []offsetCommit.Topic{
						{
							Name:       "foo",
							Partitions: []offsetCommit.Partition{{Index: 10}},
						},
					},
				}))

				res, ok := rr.Message.(*offsetCommit.Response)
				require.True(t, ok)
				require.Len(t, res.Topics, 1)
				require.Equal(t, "foo", res.Topics[0].Name)
				require.Len(t, res.Topics[0].Partitions, 1)
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

func TestOffsetCommit_Validation(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, s *store.Store, hook *test.Hook)
	}{
		{
			"invalid clientId",
			func(t *testing.T, s *store.Store, hook *test.Hook) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				defer b.Close()

				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithServer("", "kafka", b.Addr),
					asyncapitest.WithChannel("foo", asyncapitest.WithSubscribeAndPublish(
						asyncapitest.WithOperationBinding(kafka2.Operation{ClientId: schematest.New("string", schematest.WithPattern("^[A-Z]{10}[0-5]$"))}),
					))))

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)
				require.NoError(t, err)

				res, err := b.Client().OffsetCommit(2, &offsetCommit.Request{
					GroupId: "bar",
					Topics: []offsetCommit.Topic{
						{
							Name: "foo",
							Partitions: []offsetCommit.Partition{
								{},
							},
						},
					}})

				require.Len(t, res.Topics, 1)
				require.Len(t, res.Topics[0].Partitions, 1)

				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.UnknownServerError, p.ErrorCode)

				require.Equal(t, 7, len(hook.Entries))
				require.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
				require.Equal(t, "kafka OffsetCommit: invalid consumer 'kafkatest' for topic foo: invalid clientId: value 'kafkatest' does not match pattern, expected schema type=string pattern=^[A-Z]{10}[0-5]$", hook.LastEntry().Message)
			},
		},
		{
			"invalid groupId",
			func(t *testing.T, s *store.Store, hook *test.Hook) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				defer b.Close()

				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithServer("", "kafka", b.Addr),
					asyncapitest.WithChannel("foo", asyncapitest.WithSubscribeAndPublish(
						asyncapitest.WithOperationBinding(kafka2.Operation{GroupId: schematest.New("string", schematest.WithPattern("^[A-Z]{10}[0-5]$"))}),
					))))

				err := b.Client().JoinSyncGroup("foo", "bar", 3, 3)
				require.NoError(t, err)

				res, err := b.Client().OffsetCommit(2, &offsetCommit.Request{
					GroupId: "bar",
					Topics: []offsetCommit.Topic{
						{
							Name: "foo",
							Partitions: []offsetCommit.Partition{
								{},
							},
						},
					}})

				require.Len(t, res.Topics, 1)
				require.Len(t, res.Topics[0].Partitions, 1)

				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.InvalidGroupId, p.ErrorCode)

				require.Equal(t, 7, len(hook.Entries))
				require.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
				require.Equal(t, "kafka OffsetCommit: invalid consumer 'kafkatest' for topic foo: invalid groupId: value 'bar' does not match pattern, expected schema type=string pattern=^[A-Z]{10}[0-5]$", hook.LastEntry().Message)
			},
		},
		{
			"valid groupId and clientId",
			func(t *testing.T, s *store.Store, hook *test.Hook) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				c := kafkatest.NewClient(b.Addr, "MOKAPITEST1")
				defer b.Close()

				s.Update(asyncapitest.NewConfig(
					asyncapitest.WithServer("", "kafka", b.Addr),
					asyncapitest.WithChannel("foo", asyncapitest.WithSubscribeAndPublish(
						asyncapitest.WithOperationBinding(kafka2.Operation{
							ClientId: schematest.New("string", schematest.WithPattern("^[A-Z]{10}[0-5]$")),
							GroupId:  schematest.New("string", schematest.WithPattern("^[A-Z]{5}[0-5]$")),
						}),
					))))

				err := c.JoinSyncGroup("member1", "GROUP1", 3, 3)
				require.NoError(t, err)

				res, err := c.OffsetCommit(2, &offsetCommit.Request{
					MemberId: "member1",
					GroupId:  "GROUP1",
					Topics: []offsetCommit.Topic{
						{
							Name: "foo",
							Partitions: []offsetCommit.Partition{
								{},
							},
						},
					}})

				require.Len(t, res.Topics, 1)
				require.Len(t, res.Topics[0].Partitions, 1)

				p := res.Topics[0].Partitions[0]
				require.Equal(t, kafka.None, p.ErrorCode)

				require.Equal(t, 7, len(hook.Entries))
				require.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := store.New(asyncapitest.NewConfig(), enginetest.NewEngine())
			defer s.Close()
			hook := test.NewGlobal()
			tc.fn(t, s, hook)
		})
	}
}

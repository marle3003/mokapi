package store_test

import (
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/fetch"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/offset"
	"mokapi/kafka/produce"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/schema/json/schematest"
	"testing"
	"time"
)

func TestProduce(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, s *store.Store)
	}{
		{
			"default",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithServer("foo", "kafka", "127.0.0.1"),
					asyncapi3test.WithChannel("foo")))
				g := s.GetOrCreateGroup("foo", 0)
				g.Commit("foo", 0, 0)
				events.SetStore(5, events.NewTraits().WithNamespace("kafka"))

				rr := kafkatest.NewRecorder()
				r := kafkatest.NewRequest("kafkatest", 3, &produce.Request{
					Topics: []produce.RequestTopic{
						{Name: "foo", Partitions: []produce.RequestPartition{
							{
								Record: kafka.RecordBatch{
									Records: []*kafka.Record{
										{
											Offset:  0,
											Time:    time.Now(),
											Key:     kafka.NewBytes([]byte("foo-1")),
											Value:   kafka.NewBytes([]byte("bar-1")),
											Headers: nil,
										},
										{
											Offset:  1,
											Time:    time.Now(),
											Key:     kafka.NewBytes([]byte("foo-2")),
											Value:   kafka.NewBytes([]byte("bar-2")),
											Headers: nil,
										},
									},
								},
							},
						},
						}}})
				m := monitor.New()
				r.Context = monitor.NewKafkaContext(r.Context, m.Kafka)
				s.ServeMessage(rr, r)

				res, ok := rr.Message.(*produce.Response)
				require.True(t, ok)
				require.Equal(t, "foo", res.Topics[0].Name)
				require.Equal(t, kafka.None, res.Topics[0].Partitions[0].ErrorCode)
				require.Equal(t, int64(0), res.Topics[0].Partitions[0].BaseOffset)

				rr = kafkatest.NewRecorder()
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

				oRes, ok := rr.Message.(*offset.Response)
				require.True(t, ok)
				require.Equal(t, kafka.None, oRes.Topics[0].Partitions[0].ErrorCode)
				require.Equal(t, int64(2), oRes.Topics[0].Partitions[0].Offset)

				rr = kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &fetch.Request{
					MaxBytes: 1000,
					Topics: []fetch.Topic{
						{
							Name: "foo",
							Partitions: []fetch.RequestPartition{{
								MaxBytes: 1000,
							}},
						},
					}}))
				fRes, ok := rr.Message.(*fetch.Response)
				require.True(t, ok)
				require.Equal(t, kafka.None, fRes.ErrorCode)
				require.Equal(t, kafka.None, fRes.Topics[0].Partitions[0].ErrorCode)

				require.Len(t, fRes.Topics[0].Partitions[0].RecordSet.Records, 2)
				require.Equal(t, int64(2), fRes.Topics[0].Partitions[0].HighWatermark)

				record1 := fRes.Topics[0].Partitions[0].RecordSet.Records[0]
				require.Equal(t, int64(0), record1.Offset)
				require.Equal(t, "foo-1", kafkatest.BytesToString(record1.Key))
				require.Equal(t, "bar-1", kafkatest.BytesToString(record1.Value))

				record2 := fRes.Topics[0].Partitions[0].RecordSet.Records[1]
				require.Equal(t, int64(1), record2.Offset)
				require.Equal(t, "foo-2", kafkatest.BytesToString(record2.Key))
				require.Equal(t, "bar-2", kafkatest.BytesToString(record2.Value))

				// monitor
				time.Sleep(100 * time.Millisecond)
				require.Equal(t, 2.0, m.Kafka.Messages.WithLabel("test", "foo").Value())
				require.Less(t, 0.0, m.Kafka.LastMessage.WithLabel("test", "foo").Value())
				require.Equal(t, 2.0, m.Kafka.Lags.WithLabel("test", "foo", "foo", "0").Value())

				logs := events.GetEvents(events.NewTraits().WithNamespace("kafka").WithName("test").With("topic", "foo"))
				require.Len(t, logs, 2)
				require.Equal(t, "foo-2", logs[0].Data.(*store.KafkaLog).Key)
				require.Equal(t, "bar-2", logs[0].Data.(*store.KafkaLog).Message)
				require.Equal(t, int64(1), logs[0].Data.(*store.KafkaLog).Offset)

				require.Equal(t, int64(0), logs[1].Data.(*store.KafkaLog).Offset)
			},
		},
		{
			"Base Offset",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithChannel("foo")))
				s.Topic("foo").Partition(0).Write(kafka.RecordBatch{Records: []*kafka.Record{
					{
						Key:   kafka.NewBytes([]byte("foo")),
						Value: kafka.NewBytes([]byte("bar")),
					},
				},
				})

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &produce.Request{
					Topics: []produce.RequestTopic{
						{Name: "foo", Partitions: []produce.RequestPartition{
							{
								Index: 0,
								Record: kafka.RecordBatch{
									Records: []*kafka.Record{
										{
											Time:    time.Now(),
											Key:     kafka.NewBytes([]byte("foo-1")),
											Value:   kafka.NewBytes([]byte("bar-1")),
											Headers: nil,
										},
									},
								},
							},
						},
						}},
				}))

				res, ok := rr.Message.(*produce.Response)
				require.True(t, ok)
				require.Equal(t, "foo", res.Topics[0].Name)
				require.Equal(t, kafka.None, res.Topics[0].Partitions[0].ErrorCode)
				require.Equal(t, int64(1), res.Topics[0].Partitions[0].BaseOffset)
			},
		},
		{
			"invalid message value format",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithChannel("foo",
						asyncapi3test.WithMessage("foo",
							asyncapi3test.WithContentType("application/json"),
							asyncapi3test.WithPayload(schematest.New("integer"))),
						asyncapi3test.WithTopicBinding(asyncapi3.TopicBindings{ValueSchemaValidation: true}),
					),
				))

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &produce.Request{
					Topics: []produce.RequestTopic{
						{Name: "foo", Partitions: []produce.RequestPartition{
							{
								Index: 0,
								Record: kafka.RecordBatch{
									Records: []*kafka.Record{
										{
											Offset:  0,
											Time:    time.Now(),
											Key:     kafka.NewBytes([]byte(`"foo-1"`)),
											Value:   kafka.NewBytes([]byte(`"bar-1"`)),
											Headers: nil,
										},
									},
								},
							},
						},
						}},
				}))

				res, ok := rr.Message.(*produce.Response)
				require.True(t, ok)
				require.Equal(t, "foo", res.Topics[0].Name)
				require.Equal(t, kafka.InvalidRecord, res.Topics[0].Partitions[0].ErrorCode)
				require.Equal(t, int64(0), res.Topics[0].Partitions[0].BaseOffset)
			},
		},
		{
			"invalid client id",
			func(t *testing.T, s *store.Store) {
				ch := asyncapi3test.NewChannel(
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithContentType("application/json"),
						asyncapi3test.WithPayload(schematest.New("integer"))),
				)
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("foo", ch),
					asyncapi3test.WithOperation("foo",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.WithOperationBinding(asyncapi3.KafkaOperationBinding{ClientId: schematest.New("string", schematest.WithPattern("^[A-Z]{10}[0-5]$"))}),
					),
				))
				hook := test.NewGlobal()

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &produce.Request{
					Topics: []produce.RequestTopic{
						{Name: "foo", Partitions: []produce.RequestPartition{
							{
								Index: 0,
								Record: kafka.RecordBatch{
									Records: []*kafka.Record{
										{
											Offset:  0,
											Time:    time.Now(),
											Key:     kafka.NewBytes([]byte(`"foo-1"`)),
											Value:   kafka.NewBytes([]byte(`4`)),
											Headers: nil,
										},
									},
								},
							},
						},
						}},
				}))

				res, ok := rr.Message.(*produce.Response)
				require.True(t, ok)
				require.Equal(t, "foo", res.Topics[0].Name)
				require.Equal(t, kafka.UnknownServerError, res.Topics[0].Partitions[0].ErrorCode, "expected kafka error UnknownServerError")
				require.Equal(t, int64(0), res.Topics[0].Partitions[0].BaseOffset)
				require.Equal(t, "invalid producer clientId 'kafkatest' for topic foo: found 1 error:\nstring 'kafkatest' does not match regex pattern '^[A-Z]{10}[0-5]$'\nschema path #/pattern", res.Topics[0].Partitions[0].ErrorMessage)

				require.Equal(t, 1, len(hook.Entries))
				require.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
				require.Equal(t, "kafka Produce: invalid producer clientId 'kafkatest' for topic foo: found 1 error:\nstring 'kafkatest' does not match regex pattern '^[A-Z]{10}[0-5]$'\nschema path #/pattern", hook.LastEntry().Message)
			},
		},
		{
			"valid client id",
			func(t *testing.T, s *store.Store) {
				ch := asyncapi3test.NewChannel(
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithContentType("application/json"),
						asyncapi3test.WithPayload(schematest.New("integer"))),
				)
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("foo", ch),
					asyncapi3test.WithOperation("foo",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.WithOperationBinding(asyncapi3.KafkaOperationBinding{ClientId: schematest.New("string", schematest.WithPattern("^[A-Z]{10}[0-5]$"))}),
					),
				))
				hook := test.NewGlobal()

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("MOKAPITEST1", 3, &produce.Request{
					Topics: []produce.RequestTopic{
						{Name: "foo", Partitions: []produce.RequestPartition{
							{
								Index: 0,
								Record: kafka.RecordBatch{
									Records: []*kafka.Record{
										{
											Offset:  0,
											Time:    time.Now(),
											Key:     kafka.NewBytes([]byte(`"foo-1"`)),
											Value:   kafka.NewBytes([]byte(`4`)),
											Headers: nil,
										},
									},
								},
							},
						},
						}},
				}))

				res, ok := rr.Message.(*produce.Response)
				require.True(t, ok)
				require.Equal(t, "foo", res.Topics[0].Name)
				require.Equal(t, kafka.None, res.Topics[0].Partitions[0].ErrorCode, "expected no kafka error")
				require.Equal(t, int64(0), res.Topics[0].Partitions[0].BaseOffset)
				require.Equal(t, "", res.Topics[0].Partitions[0].ErrorMessage)

				require.Equal(t, 0, len(hook.Entries))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer events.Reset()

			s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngine())
			defer s.Close()
			tc.fn(t, s)
		})
	}
}

func TestProduceTriggersEvent(t *testing.T) {
	defer events.Reset()
	triggerCount := 0
	s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngineWithHandler(func(event string, args ...interface{}) []*common.Action {
		triggerCount++
		return nil
	}))
	defer s.Close()

	s.Update(asyncapi3test.NewConfig(
		asyncapi3test.WithServer("foo", "kafka", "127.0.0.1"),
		asyncapi3test.WithChannel("foo")))
	g := s.GetOrCreateGroup("foo", 0)
	g.Commit("foo", 0, 0)
	events.SetStore(5, events.NewTraits().WithNamespace("kafka"))

	rr := kafkatest.NewRecorder()
	r := kafkatest.NewRequest("kafkatest", 3, &produce.Request{
		Topics: []produce.RequestTopic{
			{Name: "foo", Partitions: []produce.RequestPartition{
				{
					Record: kafka.RecordBatch{
						Records: []*kafka.Record{
							{
								Offset:  0,
								Time:    time.Now(),
								Key:     kafka.NewBytes([]byte("foo-1")),
								Value:   kafka.NewBytes([]byte("bar-1")),
								Headers: nil,
							},
							{
								Offset:  1,
								Time:    time.Now(),
								Key:     kafka.NewBytes([]byte("foo-2")),
								Value:   kafka.NewBytes([]byte("bar-2")),
								Headers: nil,
							},
						},
					},
				},
			},
			}}})

	s.ServeMessage(rr, r)

	require.Equal(t, 2, triggerCount)
}

package kafka_test

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/fetch"
	"mokapi/kafka/protocol/offset"
	"mokapi/kafka/protocol/produce"
	"mokapi/kafka/store"
	"mokapi/test"
	"testing"
	"time"
)

func TestProduce(t *testing.T) {
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

				r, err := b.Client().Produce(3, &produce.Request{Topics: []produce.RequestTopic{
					{Name: "foo", Partitions: []produce.RequestPartition{
						{
							Index: 0,
							Record: protocol.RecordBatch{
								Records: []protocol.Record{
									{
										Offset:  0,
										Time:    time.Now(),
										Key:     protocol.NewBytes([]byte("foo-1")),
										Value:   protocol.NewBytes([]byte("bar-1")),
										Headers: nil,
									},
									{
										Offset:  1,
										Time:    time.Now(),
										Key:     protocol.NewBytes([]byte("foo-2")),
										Value:   protocol.NewBytes([]byte("bar-2")),
										Headers: nil,
									},
								},
							},
						},
					},
					}},
				})
				test.Ok(t, err)
				test.Equals(t, "foo", r.Topics[0].Name)
				test.Equals(t, protocol.None, r.Topics[0].Partitions[0].ErrorCode)
				test.Equals(t, int64(0), r.Topics[0].Partitions[0].BaseOffset)

				offset, err := b.Client().Offset(3, &offset.Request{Topics: []offset.RequestTopic{
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
				test.Equals(t, protocol.None, r.Topics[0].Partitions[0].ErrorCode)
				test.Equals(t, int64(1), offset.Topics[0].Partitions[0].Offset)

				fetch, err := b.Client().Fetch(3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name: "foo",
						Partitions: []fetch.RequestPartition{{
							Index:              0,
							CurrentLeaderEpoch: 0,
							FetchOffset:        0,
							LogStartOffset:     0,
							MaxBytes:           1000,
						}},
					},
				}})
				test.Ok(t, err)
				test.Equals(t, protocol.None, fetch.ErrorCode)
				test.Equals(t, protocol.None, fetch.Topics[0].Partitions[0].ErrorCode)

				test.Equals(t, 2, len(fetch.Topics[0].Partitions[0].RecordSet.Records))
				test.Equals(t, int64(1), fetch.Topics[0].Partitions[0].HighWatermark)

				record1 := fetch.Topics[0].Partitions[0].RecordSet.Records[0]
				test.Equals(t, int64(0), record1.Offset)
				test.Equals(t, "foo-1", kafkatest.BytesToString(record1.Key))
				test.Equals(t, "bar-1", kafkatest.BytesToString(record1.Value))

				record2 := fetch.Topics[0].Partitions[0].RecordSet.Records[1]
				test.Equals(t, int64(1), record2.Offset)
				test.Equals(t, "foo-2", kafkatest.BytesToString(record2.Key))
				test.Equals(t, "bar-2", kafkatest.BytesToString(record2.Value))
			},
		},
		{
			"Base Offset",
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
				r, err := b.Client().Produce(3, &produce.Request{Topics: []produce.RequestTopic{
					{Name: "foo", Partitions: []produce.RequestPartition{
						{
							Index: 0,
							Record: protocol.RecordBatch{
								Records: []protocol.Record{
									{
										Offset:  0,
										Time:    time.Now(),
										Key:     protocol.NewBytes([]byte("foo-1")),
										Value:   protocol.NewBytes([]byte("bar-1")),
										Headers: nil,
									},
								},
							},
						},
					},
					}},
				})
				test.Ok(t, err)
				test.Equals(t, "foo", r.Topics[0].Name)
				test.Equals(t, protocol.None, r.Topics[0].Partitions[0].ErrorCode)
				test.Equals(t, int64(1), r.Topics[0].Partitions[0].BaseOffset)
			},
		},
		{
			"invalid message value format",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(store.New(asyncapitest.NewConfig(
					asyncapitest.WithServer(b.Listener.Addr().String(), "kafka", b.Listener.Addr().String()),
					asyncapitest.WithChannel("foo", asyncapitest.WithSubscribeAndPublish(
						asyncapitest.WithMessage(
							asyncapitest.WithContentType("application/json"),
							asyncapitest.WithPayload(openapitest.NewSchema("integer"))),
					)),
				)))

				r, err := b.Client().Produce(3, &produce.Request{Topics: []produce.RequestTopic{
					{Name: "foo", Partitions: []produce.RequestPartition{
						{
							Index: 0,
							Record: protocol.RecordBatch{
								Records: []protocol.Record{
									{
										Offset:  0,
										Time:    time.Now(),
										Key:     protocol.NewBytes([]byte(`"foo-1"`)),
										Value:   protocol.NewBytes([]byte(`"bar-1"`)),
										Headers: nil,
									},
								},
							},
						},
					},
					}},
				})
				test.Ok(t, err)
				test.Equals(t, "foo", r.Topics[0].Name)
				test.Equals(t, protocol.CorruptMessage, r.Topics[0].Partitions[0].ErrorCode)
				test.Equals(t, int64(0), r.Topics[0].Partitions[0].BaseOffset)
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

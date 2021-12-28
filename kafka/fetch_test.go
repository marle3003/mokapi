package kafka_test

import (
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/fetch"
	"mokapi/test"
	"testing"
	"time"
)

func TestFetch(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, b *kafkatest.Broker)
	}{
		{
			"topic not exists",
			func(t *testing.T, b *kafkatest.Broker) {
				r, err := b.Client().Fetch(3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name:       "foo",
						Partitions: []fetch.RequestPartition{{}},
					},
				}})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, "foo", r.Topics[0].Name)
				test.Equals(t, 1, len(r.Topics[0].Partitions))
				test.Equals(t, protocol.UnknownTopicOrPartition, r.Topics[0].Partitions[0].ErrorCode)
			},
		},
		{
			"partition not exists",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{Topics: []kafkatest.TopicConfig{{"foo", 0}}}))
				r, err := b.Client().Fetch(3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name:       "foo",
						Partitions: []fetch.RequestPartition{{}},
					},
				}})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, "foo", r.Topics[0].Name)
				test.Equals(t, 1, len(r.Topics[0].Partitions))
				test.Equals(t, protocol.UnknownTopicOrPartition, r.Topics[0].Partitions[0].ErrorCode)
			},
		},
		{
			"empty",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{Topics: []kafkatest.TopicConfig{{"foo", 1}}}))
				r, err := b.Client().Fetch(3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name:       "foo",
						Partitions: []fetch.RequestPartition{{}},
					},
				}})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, 1, len(r.Topics))
				test.Equals(t, "foo", r.Topics[0].Name)
				test.Equals(t, 1, len(r.Topics[0].Partitions))
				test.Equals(t, protocol.None, r.Topics[0].Partitions[0].ErrorCode)
				test.Equals(t, 0, len(r.Topics[0].Partitions[0].RecordSet.Records))
			},
		},
		{
			"empty with max wait time",
			func(t *testing.T, b *kafkatest.Broker) {
				start := time.Now()
				_, err := b.Client().Fetch(3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name:       "foo",
						Partitions: []fetch.RequestPartition{{}},
					},
				}, MaxWaitMs: 1000})
				end := time.Now()
				test.Ok(t, err)
				waitTime := end.Sub(start).Milliseconds()
				// fetch request waits for MaxWaitMs - 200ms
				test.Assert(t, waitTime < 100, "wait time should be 800ms but was %v", waitTime)
			},
		},
		{
			"empty with max wait time and min bytes",
			func(t *testing.T, b *kafkatest.Broker) {
				start := time.Now()
				_, err := b.Client().Fetch(3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name: "foo",
						Partitions: []fetch.RequestPartition{{
							Index:              0,
							CurrentLeaderEpoch: 0,
							FetchOffset:        0,
							LogStartOffset:     0,
							MaxBytes:           0,
						}},
					},
				}, MaxWaitMs: 1000, MinBytes: 1})
				end := time.Now()
				test.Ok(t, err)
				waitTime := end.Sub(start).Milliseconds()
				// fetch request waits for MaxWaitMs - 200ms
				test.Assert(t, waitTime > 800, "wait time should be 800ms but was %v", waitTime)
			},
		},
		{
			"fetch one record",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{Topics: []kafkatest.TopicConfig{{"foo", 1}}}))
				b.Store().Topic("foo").Partition(0).Write(protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Key:   protocol.NewBytes([]byte("foo")),
							Value: protocol.NewBytes([]byte("bar")),
						},
					},
				})
				r, err := b.Client().Fetch(3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name: "foo",
						Partitions: []fetch.RequestPartition{{
							MaxBytes: 1,
						}},
					},
				}})
				test.Ok(t, err)
				test.Equals(t, 1, len(r.Topics[0].Partitions[0].RecordSet.Records))
				test.Equals(t, int64(0), r.Topics[0].Partitions[0].HighWatermark)

				record := r.Topics[0].Partitions[0].RecordSet.Records[0]
				test.Equals(t, int64(0), record.Offset)
				test.Equals(t, "foo", kafkatest.BytesToString(record.Key))
				test.Equals(t, "bar", kafkatest.BytesToString(record.Value))
			},
		},
		{
			"fetch one record with MaxBytes 1",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{Topics: []kafkatest.TopicConfig{{"foo", 1}}}))
				b.Store().Topic("foo").Partition(0).Write(protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Key:   protocol.NewBytes([]byte("key-1")),
							Value: protocol.NewBytes([]byte("value-1")),
						},
						{
							Key:   protocol.NewBytes([]byte("key-2")),
							Value: protocol.NewBytes([]byte("value-2")),
						},
					},
				})
				r, err := b.Client().Fetch(3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name: "foo",
						Partitions: []fetch.RequestPartition{{
							MaxBytes: 1,
						}},
					},
				}})
				test.Ok(t, err)
				// only one record returned because of MaxBytes 1
				test.Equals(t, 1, len(r.Topics[0].Partitions[0].RecordSet.Records))
				test.Equals(t, int64(1), r.Topics[0].Partitions[0].HighWatermark)

				record := r.Topics[0].Partitions[0].RecordSet.Records[0]
				test.Equals(t, int64(0), record.Offset)
				test.Equals(t, "key-1", kafkatest.BytesToString(record.Key))
				test.Equals(t, "value-1", kafkatest.BytesToString(record.Value))
			},
		},
		{
			"fetch two records",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{Topics: []kafkatest.TopicConfig{{"foo", 1}}}))
				b.Store().Topic("foo").Partition(0).Write(protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Key:   protocol.NewBytes([]byte("key-1")),
							Value: protocol.NewBytes([]byte("value-1")),
						},
						{
							Key:   protocol.NewBytes([]byte("key-2")),
							Value: protocol.NewBytes([]byte("value-2")),
						},
					},
				})
				r, err := b.Client().Fetch(3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name: "foo",
						Partitions: []fetch.RequestPartition{{
							Index:       0,
							FetchOffset: 0,
							MaxBytes:    500,
						}},
					},
				}})
				test.Ok(t, err)
				test.Equals(t, 2, len(r.Topics[0].Partitions[0].RecordSet.Records))
				test.Equals(t, int64(1), r.Topics[0].Partitions[0].HighWatermark)

				record1 := r.Topics[0].Partitions[0].RecordSet.Records[0]
				test.Equals(t, int64(0), record1.Offset)
				test.Equals(t, "key-1", kafkatest.BytesToString(record1.Key))
				test.Equals(t, "value-1", kafkatest.BytesToString(record1.Value))

				record2 := r.Topics[0].Partitions[0].RecordSet.Records[1]
				test.Equals(t, int64(1), record2.Offset)
				test.Equals(t, "key-2", kafkatest.BytesToString(record2.Key))
				test.Equals(t, "value-2", kafkatest.BytesToString(record2.Value))
			},
		},
		{
			"wait fetch for MinBytes",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{Topics: []kafkatest.TopicConfig{{"foo", 1}}}))
				ch := make(chan *fetch.Response, 1)
				go func() {
					r, err := b.Client().Fetch(3, &fetch.Request{Topics: []fetch.Topic{
						{
							Name: "foo",
							Partitions: []fetch.RequestPartition{{
								Index:              0,
								CurrentLeaderEpoch: 0,
								FetchOffset:        0,
								LogStartOffset:     0,
								MaxBytes:           1,
							}},
						},
					}, MinBytes: 1, MaxWaitMs: 5000})
					if err != nil {
						panic(err)
					}
					ch <- r
				}()
				time.Sleep(300 * time.Millisecond)
				b.Store().Topic("foo").Partition(0).Write(protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Key:   protocol.NewBytes([]byte("foo")),
							Value: protocol.NewBytes([]byte("bar")),
						},
					},
				})
				r := <-ch

				test.Equals(t, 1, len(r.Topics[0].Partitions[0].RecordSet.Records))
				test.Equals(t, int64(0), r.Topics[0].Partitions[0].HighWatermark)

				record := r.Topics[0].Partitions[0].RecordSet.Records[0]
				test.Equals(t, int64(0), record.Offset)
				test.Equals(t, "foo", kafkatest.BytesToString(record.Key))
				test.Equals(t, "bar", kafkatest.BytesToString(record.Value))
			},
		},
		{
			"fetch offset out of range when empty",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{Topics: []kafkatest.TopicConfig{{"foo", 1}}}))
				r, err := b.Client().Fetch(3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name: "foo",
						Partitions: []fetch.RequestPartition{{
							Index:              0,
							CurrentLeaderEpoch: 0,
							FetchOffset:        1,
							LogStartOffset:     0,
							MaxBytes:           0,
						}},
					},
				}})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, protocol.None, r.Topics[0].Partitions[0].ErrorCode)
			},
		},
		{
			"fetch offset out of range",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{Topics: []kafkatest.TopicConfig{{"foo", 1}}}))
				b.Store().Topic("foo").Partition(0).Write(protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Key:   protocol.NewBytes([]byte("foo")),
							Value: protocol.NewBytes([]byte("bar")),
						},
					},
				})
				r, err := b.Client().Fetch(3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name: "foo",
						Partitions: []fetch.RequestPartition{{
							Index:              0,
							CurrentLeaderEpoch: 0,
							FetchOffset:        1,
							LogStartOffset:     0,
							MaxBytes:           0,
						}},
					},
				}})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, protocol.OffsetOutOfRange, r.Topics[0].Partitions[0].ErrorCode)
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

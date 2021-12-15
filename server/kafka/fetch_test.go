package kafka_test

import (
	"mokapi/server/kafka/kafkatest"
	"mokapi/server/kafka/memory"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/fetch"
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
			"topic not exists",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetCluster(memory.NewCluster(memory.Schema{Topics: []memory.TopicSchema{
					{Name: "foo"},
				}}))
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
				b.SetCluster(memory.NewCluster(memory.Schema{Topics: []memory.TopicSchema{
					{Name: "foo", Partitions: []memory.PartitionSchema{{Index: 0}}},
				}}))
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
				b.SetCluster(memory.NewCluster(memory.Schema{Topics: []memory.TopicSchema{
					{Name: "foo", Partitions: []memory.PartitionSchema{{Index: 0}}},
				}}))
				b.Cluster().Topic("foo").Partition(0).Write(protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Key:   []byte("foo"),
							Value: []byte("bar"),
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
				test.Equals(t, "foo", string(record.Key))
				test.Equals(t, "bar", string(record.Value))
			},
		},
		{
			"fetch one record with MaxBytes 1",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetCluster(memory.NewCluster(memory.Schema{Topics: []memory.TopicSchema{
					{Name: "foo", Partitions: []memory.PartitionSchema{{Index: 0}}},
				}}))
				b.Cluster().Topic("foo").Partition(0).Write(protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Key:   []byte("key-1"),
							Value: []byte("value-1"),
						},
						{
							Key:   []byte("key-2"),
							Value: []byte("value-2"),
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
				test.Equals(t, "key-1", string(record.Key))
				test.Equals(t, "value-1", string(record.Value))
			},
		},
		{
			"fetch two records",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetCluster(memory.NewCluster(memory.Schema{Topics: []memory.TopicSchema{
					{Name: "foo", Partitions: []memory.PartitionSchema{{Index: 0}}},
				}}))
				b.Cluster().Topic("foo").Partition(0).Write(protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Key:   []byte("key-1"),
							Value: []byte("value-1"),
						},
						{
							Key:   []byte("key-2"),
							Value: []byte("value-2"),
						},
					},
				})
				r, err := b.Client().Fetch(3, &fetch.Request{Topics: []fetch.Topic{
					{
						Name: "foo",
						Partitions: []fetch.RequestPartition{{
							Index:              0,
							CurrentLeaderEpoch: 0,
							FetchOffset:        0,
							LogStartOffset:     0,
							MaxBytes:           500,
						}},
					},
				}})
				test.Ok(t, err)
				test.Equals(t, 2, len(r.Topics[0].Partitions[0].RecordSet.Records))
				test.Equals(t, int64(1), r.Topics[0].Partitions[0].HighWatermark)

				record1 := r.Topics[0].Partitions[0].RecordSet.Records[0]
				test.Equals(t, int64(0), record1.Offset)
				test.Equals(t, "key-1", string(record1.Key))
				test.Equals(t, "value-1", string(record1.Value))

				record2 := r.Topics[0].Partitions[0].RecordSet.Records[1]
				test.Equals(t, int64(1), record2.Offset)
				test.Equals(t, "key-2", string(record2.Key))
				test.Equals(t, "value-2", string(record2.Value))
			},
		},
		{
			"wait fetch for MinBytes",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetCluster(memory.NewCluster(memory.Schema{Topics: []memory.TopicSchema{
					{Name: "foo", Partitions: []memory.PartitionSchema{{Index: 0}}},
				}}))
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
				b.Cluster().Topic("foo").Partition(0).Write(protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Key:   []byte("foo"),
							Value: []byte("bar"),
						},
					},
				})
				r := <-ch

				test.Equals(t, 1, len(r.Topics[0].Partitions[0].RecordSet.Records))
				test.Equals(t, int64(0), r.Topics[0].Partitions[0].HighWatermark)

				record := r.Topics[0].Partitions[0].RecordSet.Records[0]
				test.Equals(t, int64(0), record.Offset)
				test.Equals(t, "foo", string(record.Key))
				test.Equals(t, "bar", string(record.Value))
			},
		},
		{
			"fetch offset out of range when empty",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetCluster(memory.NewCluster(memory.Schema{Topics: []memory.TopicSchema{
					{Name: "foo", Partitions: []memory.PartitionSchema{{Index: 0}}},
				}}))
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
				b.SetCluster(memory.NewCluster(memory.Schema{Topics: []memory.TopicSchema{
					{Name: "foo", Partitions: []memory.PartitionSchema{{Index: 0}}},
				}}))
				b.Cluster().Topic("foo").Partition(0).Write(protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Key:   []byte("foo"),
							Value: []byte("bar"),
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

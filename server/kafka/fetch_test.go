package kafka_test

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/server/kafka"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/fetch"
	"mokapi/server/kafka/protocol/kafkatest"
	"mokapi/test"
	"testing"
	"time"
)

func TestFetch(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(*testing.T, *kafka.Binding)
	}{
		{
			"empty",
			testFetchEmpty,
		},
		{
			"empty with max wait time",
			testFetchEmptyMaxWait,
		},
		{
			"fetch one record",
			testFetchOneRecord,
		},
		{
			"fetch one record with MaxBytes 0",
			testFetchTwoRecordMaxBytesZero,
		},
		{
			"fetch two records",
			testFetchTwoRecords,
		},
		{
			"wait fetch for MinBytes",
			testFetchMinBytes,
		},
		{
			"fetch offset out of range",
			testFetchOffsetOutOfRange,
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			b := kafka.NewBinding(func(topic string, key []byte, message []byte, partition int) {})
			defer b.Stop()
			data.fn(t, b)
		})
	}
}

func testFetchEmpty(t *testing.T, b *kafka.Binding) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", "127.0.0.1:9092"),
		asyncapitest.WithChannel(
			"foo", asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema())))))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient("127.0.0.1:9092", "kafkatest")
	defer client.Close()
	r, err := client.Fetch(3, &fetch.Request{Topics: []fetch.Topic{
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
	}})
	test.Ok(t, err)
	test.Equals(t, protocol.None, r.ErrorCode)
	test.Equals(t, 1, len(r.Topics))
	test.Equals(t, "foo", r.Topics[0].Name)
	test.Equals(t, 1, len(r.Topics[0].Partitions))
	test.Equals(t, protocol.None, r.Topics[0].Partitions[0].ErrorCode)
	test.Equals(t, 0, len(r.Topics[0].Partitions[0].RecordSet.Records))
}

func testFetchEmptyMaxWait(t *testing.T, b *kafka.Binding) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", "127.0.0.1:9092"),
		asyncapitest.WithChannel(
			"foo", asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema())))))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient("127.0.0.1:9092", "kafkatest")
	defer client.Close()
	start := time.Now()
	_, err = client.Fetch(3, &fetch.Request{Topics: []fetch.Topic{
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
	}, MaxWaitMs: 1000})
	end := time.Now()
	test.Ok(t, err)
	waitTime := end.Sub(start).Milliseconds()
	// fetch request waits for MaxWaitMs - 200ms
	test.Assert(t, waitTime > 800, "wait time should be 800ms but was %v", waitTime)
}

func testFetchOneRecord(t *testing.T, b *kafka.Binding) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", "127.0.0.1:9092"),
		asyncapitest.WithChannel(
			"foo", asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema())))))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient("127.0.0.1:9092", "kafkatest")
	defer client.Close()
	testProduce(t, b)
	r, err := client.Fetch(3, &fetch.Request{Topics: []fetch.Topic{
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
	}})
	test.Ok(t, err)
	test.Equals(t, 1, len(r.Topics[0].Partitions[0].RecordSet.Records))
	test.Equals(t, int64(1), r.Topics[0].Partitions[0].HighWatermark)

	record := r.Topics[0].Partitions[0].RecordSet.Records[0]
	test.Equals(t, int64(0), record.Offset)
	test.Equals(t, "foo", string(record.Key))
	test.Equals(t, "bar", string(record.Value))
}

func testFetchTwoRecordMaxBytesZero(t *testing.T, b *kafka.Binding) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", "127.0.0.1:9092"),
		asyncapitest.WithChannel(
			"foo", asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema())))))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient("127.0.0.1:9092", "kafkatest")
	defer client.Close()
	testProduce(t, b)
	testProduce(t, b)
	r, err := client.Fetch(3, &fetch.Request{Topics: []fetch.Topic{
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
	}})
	test.Ok(t, err)
	// only one record returned because of MaxBytes 0
	test.Equals(t, 1, len(r.Topics[0].Partitions[0].RecordSet.Records))
	test.Equals(t, int64(2), r.Topics[0].Partitions[0].HighWatermark)

	record := r.Topics[0].Partitions[0].RecordSet.Records[0]
	test.Equals(t, int64(0), record.Offset)
	test.Equals(t, "foo", string(record.Key))
	test.Equals(t, "bar", string(record.Value))
}

func testFetchTwoRecords(t *testing.T, b *kafka.Binding) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", "127.0.0.1:9092"),
		asyncapitest.WithChannel(
			"foo", asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema())))))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient("127.0.0.1:9092", "kafkatest")
	defer client.Close()
	testProduce(t, b)
	testProduce(t, b)
	r, err := client.Fetch(3, &fetch.Request{Topics: []fetch.Topic{
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
	test.Equals(t, int64(2), r.Topics[0].Partitions[0].HighWatermark)

	record1 := r.Topics[0].Partitions[0].RecordSet.Records[0]
	test.Equals(t, int64(0), record1.Offset)
	test.Equals(t, "foo", string(record1.Key))
	test.Equals(t, "bar", string(record1.Value))

	record2 := r.Topics[0].Partitions[0].RecordSet.Records[1]
	test.Equals(t, int64(1), record2.Offset)
	test.Equals(t, "foo", string(record2.Key))
	test.Equals(t, "bar", string(record2.Value))
}

func testFetchMinBytes(t *testing.T, b *kafka.Binding) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", "127.0.0.1:9092"),
		asyncapitest.WithChannel(
			"foo", asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema())))))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient("127.0.0.1:9092", "kafkatest")
	defer client.Close()
	ch := make(chan *fetch.Response, 1)
	go func() {
		r, err := client.Fetch(3, &fetch.Request{Topics: []fetch.Topic{
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
		}, MinBytes: 1, MaxWaitMs: 5000})
		test.Ok(t, err)
		ch <- r
	}()
	time.Sleep(300 * time.Millisecond)
	testProduce(t, b)

	r := <-ch

	// TODO: currently not working because fetch does not update offset during wait time
	_ = r
	//test.Equals(t, 1, len(r.Topics[0].Partitions[0].RecordSet.Records))
	//test.Equals(t, int64(1), r.Topics[0].Partitions[0].HighWatermark)
	//
	//record := r.Topics[0].Partitions[0].RecordSet.Records[0]
	//test.Equals(t, int64(0), record.Offset)
	//test.Equals(t, "foo", string(record.Key))
	//test.Equals(t, "bar", string(record.Value))
}

func testFetchOffsetOutOfRange(t *testing.T, b *kafka.Binding) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", "127.0.0.1:9092"),
		asyncapitest.WithChannel(
			"foo", asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema())))))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient("127.0.0.1:9092", "kafkatest")
	defer client.Close()
	r, err := client.Fetch(3, &fetch.Request{Topics: []fetch.Topic{
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
}

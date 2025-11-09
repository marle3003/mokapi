package fetch_test

import (
	"bytes"
	"encoding/binary"
	"mokapi/kafka"
	"mokapi/kafka/fetch"
	"mokapi/kafka/kafkatest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.Fetch]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(12), reg.MaxVersion)
}

func TestRequest(t *testing.T) {
	kafkatest.TestRequest(t, 11, &fetch.Request{
		ReplicaId:      0,
		MaxWaitMs:      123,
		MinBytes:       456,
		MaxBytes:       789,
		IsolationLevel: 0,
		SessionId:      0,
		SessionEpoch:   0,
		Topics: []fetch.Topic{
			{
				Name: "foo",
				Partitions: []fetch.RequestPartition{
					{
						Index:              0,
						CurrentLeaderEpoch: 0,
						FetchOffset:        0,
						LastFetchedEpoch:   0,
						LogStartOffset:     0,
						MaxBytes:           123,
					},
				},
			},
		},
		ForgottenTopics: []fetch.Topic{},
		RackId:          "bar",
	})

	kafkatest.TestRequest(t, 12, &fetch.Request{
		ReplicaId:      0,
		MaxWaitMs:      123,
		MinBytes:       456,
		MaxBytes:       789,
		IsolationLevel: 0,
		SessionId:      0,
		SessionEpoch:   0,
		Topics: []fetch.Topic{
			{
				Name: "foo",
				Partitions: []fetch.RequestPartition{
					{
						Index:              0,
						CurrentLeaderEpoch: 0,
						FetchOffset:        0,
						LastFetchedEpoch:   0,
						LogStartOffset:     0,
						MaxBytes:           123,
					},
				},
			},
		},
		ForgottenTopics: []fetch.Topic{},
		RackId:          "bar",
	})

	b := kafkatest.WriteRequest(t, 12, 123, "me", &fetch.Request{
		ReplicaId:      0,
		MaxWaitMs:      123,
		MinBytes:       456,
		MaxBytes:       789,
		IsolationLevel: 0,
		SessionId:      0,
		SessionEpoch:   0,
		Topics: []fetch.Topic{
			{
				Name: "foo",
				Partitions: []fetch.RequestPartition{
					{
						Index:              0,
						CurrentLeaderEpoch: 0,
						FetchOffset:        0,
						LastFetchedEpoch:   0,
						LogStartOffset:     0,
						MaxBytes:           123,
					},
				},
			},
		},
		ForgottenTopics: []fetch.Topic{},
		RackId:          "bar",
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(84))          // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.Fetch)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(12))          // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))         // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))           // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))       // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))            // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(0))      // ReplicaId
	_ = binary.Write(expected, binary.BigEndian, int32(123))    // MaxWaitMs
	_ = binary.Write(expected, binary.BigEndian, int32(456))    // MinBytes
	_ = binary.Write(expected, binary.BigEndian, int32(789))    // MaxBytes
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // IsolationLevel
	_ = binary.Write(expected, binary.BigEndian, int32(0))      // SessionId
	_ = binary.Write(expected, binary.BigEndian, int32(0))      // SessionEpoch
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // Name
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Partitions length
	_ = binary.Write(expected, binary.BigEndian, int32(0))      // Index
	_ = binary.Write(expected, binary.BigEndian, int32(0))      // CurrentLeaderEpoch
	_ = binary.Write(expected, binary.BigEndian, int64(0))      // FetchOffset
	_ = binary.Write(expected, binary.BigEndian, int32(0))      // LastFetchedEpoch
	_ = binary.Write(expected, binary.BigEndian, int64(0))      // LogStartOffset
	_ = binary.Write(expected, binary.BigEndian, int32(123))    // MaxBytes
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // Partitions tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(1))       // ForgottenTopics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // RackId length
	_ = binary.Write(expected, binary.BigEndian, []byte("bar")) // Name
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	require.Equal(t, expected.Bytes(), b)
}

func TestResponse(t *testing.T) {
	kafkatest.TestResponse(t, 11, &fetch.Response{
		ThrottleTimeMs: 0,
		ErrorCode:      0,
		SessionId:      0,
		Topics: []fetch.ResponseTopic{
			{
				Name: "foo",
				Partitions: []fetch.ResponsePartition{
					{
						Index:                0,
						ErrorCode:            0,
						HighWatermark:        0,
						LastStableOffset:     0,
						LogStartOffset:       0,
						AbortedTransactions:  []fetch.AbortedTransaction{},
						PreferredReadReplica: 0,
						RecordSet: kafka.RecordBatch{Records: []*kafka.Record{
							{
								Offset:  0,
								Time:    kafka.ToTime(1657010762684),
								Key:     kafka.NewBytes([]byte("foo")),
								Value:   kafka.NewBytes([]byte("bar")),
								Headers: nil,
							},
						},
						},
					},
				},
			},
		},
	})

	kafkatest.TestResponse(t, 12, &fetch.Response{
		ThrottleTimeMs: 0,
		ErrorCode:      0,
		SessionId:      0,
		Topics: []fetch.ResponseTopic{
			{
				Name: "foo",
				Partitions: []fetch.ResponsePartition{
					{
						Index:                0,
						ErrorCode:            0,
						HighWatermark:        0,
						LastStableOffset:     0,
						LogStartOffset:       0,
						AbortedTransactions:  []fetch.AbortedTransaction{},
						PreferredReadReplica: 0,
						RecordSet: kafka.RecordBatch{Records: []*kafka.Record{
							{
								Offset:  0,
								Time:    kafka.ToTime(1657010762684),
								Key:     kafka.NewBytes([]byte("foo")),
								Value:   kafka.NewBytes([]byte("bar")),
								Headers: nil,
							},
						},
						},
					},
				},
			},
		},
	})

	b := kafkatest.WriteResponse(t, 11, 123, &fetch.Response{
		ThrottleTimeMs: 0,
		ErrorCode:      0,
		SessionId:      0,
		Topics: []fetch.ResponseTopic{
			{
				Name: "foo",
				Partitions: []fetch.ResponsePartition{
					{
						Index:                0,
						ErrorCode:            0,
						HighWatermark:        0,
						LastStableOffset:     0,
						LogStartOffset:       0,
						AbortedTransactions:  []fetch.AbortedTransaction{},
						PreferredReadReplica: 0,
						RecordSet: kafka.RecordBatch{Records: []*kafka.Record{
							{
								Offset:  0,
								Time:    kafka.ToTime(1657010762684),
								Key:     kafka.NewBytes([]byte("foo")),
								Value:   kafka.NewBytes([]byte("bar")),
								Headers: nil,
							},
						},
						},
					},
				},
			},
		},
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(143)) // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(0))                 // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int16(0))                 // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int32(0))                 // SessionId
	_ = binary.Write(expected, binary.BigEndian, int32(1))                 // Topics length
	_ = binary.Write(expected, binary.BigEndian, int16(3))                 // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo"))            // Name
	_ = binary.Write(expected, binary.BigEndian, int32(1))                 // Partitions length
	_ = binary.Write(expected, binary.BigEndian, int32(0))                 // Index
	_ = binary.Write(expected, binary.BigEndian, int16(0))                 // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int64(0))                 // HighWatermark
	_ = binary.Write(expected, binary.BigEndian, int64(0))                 // LastStableOffset
	_ = binary.Write(expected, binary.BigEndian, int64(0))                 // LogStartOffset
	_ = binary.Write(expected, binary.BigEndian, int32(0))                 // AbortedTransactions length
	_ = binary.Write(expected, binary.BigEndian, int32(0))                 // PreferredReadReplica
	_ = binary.Write(expected, binary.BigEndian, int32(74))                // Records length
	_ = binary.Write(expected, binary.BigEndian, int64(0))                 // base offset
	_ = binary.Write(expected, binary.BigEndian, int32(62))                // message size
	_ = binary.Write(expected, binary.BigEndian, int32(0))                 // leader epoch
	_ = binary.Write(expected, binary.BigEndian, int8(2))                  // magic
	_ = binary.Write(expected, binary.BigEndian, []byte{119, 89, 114, 22}) // crc32
	_ = binary.Write(expected, binary.BigEndian, int16(0))                 // attributes
	_ = binary.Write(expected, binary.BigEndian, int32(0))                 // last offset delta
	_ = binary.Write(expected, binary.BigEndian, int64(1657010762684))     // first timestamp
	_ = binary.Write(expected, binary.BigEndian, int64(1657010762684))     // max timestamp

	_ = binary.Write(expected, binary.BigEndian, []byte{255, 255, 255, 255, 255, 255, 255, 255}) // producer id
	_ = binary.Write(expected, binary.BigEndian, []byte{255, 255})                               // producer epoch
	_ = binary.Write(expected, binary.BigEndian, []byte{255, 255, 255, 255})                     // base sequence

	_ = binary.Write(expected, binary.BigEndian, int32(1))      // number of records
	_ = binary.Write(expected, binary.BigEndian, int8(24))      // record length
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // attributes
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // delta timestamp
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // delta offset
	_ = binary.Write(expected, binary.BigEndian, int8(6))       // key length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // key
	_ = binary.Write(expected, binary.BigEndian, int8(6))       // value length
	_ = binary.Write(expected, binary.BigEndian, []byte("bar")) // value
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // header length

	require.Equal(t, expected.Bytes(), b)

	b = kafkatest.WriteResponse(t, 12, 123, &fetch.Response{
		ThrottleTimeMs: 0,
		ErrorCode:      0,
		SessionId:      0,
		Topics: []fetch.ResponseTopic{
			{
				Name: "foo",
				Partitions: []fetch.ResponsePartition{
					{
						Index:                0,
						ErrorCode:            0,
						HighWatermark:        0,
						LastStableOffset:     0,
						LogStartOffset:       0,
						AbortedTransactions:  []fetch.AbortedTransaction{},
						PreferredReadReplica: 0,
						RecordSet: kafka.RecordBatch{
							Records: []*kafka.Record{
								{
									Offset:  0,
									Time:    kafka.ToTime(1657010762684),
									Key:     kafka.NewBytes([]byte("foo")),
									Value:   kafka.NewBytes([]byte("bar")),
									Headers: nil,
								},
							},
						},
					},
				},
			},
		},
	})
	expected = new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(134)) // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(0))                 // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int16(0))                 // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int32(0))                 // SessionId
	_ = binary.Write(expected, binary.BigEndian, int8(2))                  // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))                  // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo"))            // Name
	_ = binary.Write(expected, binary.BigEndian, int8(2))                  // Partitions length
	_ = binary.Write(expected, binary.BigEndian, int32(0))                 // Index
	_ = binary.Write(expected, binary.BigEndian, int16(0))                 // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int64(0))                 // HighWatermark
	_ = binary.Write(expected, binary.BigEndian, int64(0))                 // LastStableOffset
	_ = binary.Write(expected, binary.BigEndian, int64(0))                 // LogStartOffset
	_ = binary.Write(expected, binary.BigEndian, int8(1))                  // AbortedTransactions length
	_ = binary.Write(expected, binary.BigEndian, int32(0))                 // PreferredReadReplica
	_ = binary.Write(expected, binary.BigEndian, int8(75))                 // Records length
	_ = binary.Write(expected, binary.BigEndian, int64(0))                 // base offset
	_ = binary.Write(expected, binary.BigEndian, int32(62))                // message size
	_ = binary.Write(expected, binary.BigEndian, int32(0))                 // leader epoch
	_ = binary.Write(expected, binary.BigEndian, int8(2))                  // magic
	_ = binary.Write(expected, binary.BigEndian, []byte{119, 89, 114, 22}) // crc32
	_ = binary.Write(expected, binary.BigEndian, int16(0))                 // attributes
	_ = binary.Write(expected, binary.BigEndian, int32(0))                 // last offset delta
	_ = binary.Write(expected, binary.BigEndian, int64(1657010762684))     // first timestamp
	_ = binary.Write(expected, binary.BigEndian, int64(1657010762684))     // max timestamp

	_ = binary.Write(expected, binary.BigEndian, []byte{255, 255, 255, 255, 255, 255, 255, 255}) // producer id
	_ = binary.Write(expected, binary.BigEndian, []byte{255, 255})                               // producer epoch
	_ = binary.Write(expected, binary.BigEndian, []byte{255, 255, 255, 255})                     // base sequence

	_ = binary.Write(expected, binary.BigEndian, int32(1))      // number of records
	_ = binary.Write(expected, binary.BigEndian, int8(24))      // record length
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // attributes
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // delta timestamp
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // delta offset
	_ = binary.Write(expected, binary.BigEndian, int8(6))       // key length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // key
	_ = binary.Write(expected, binary.BigEndian, int8(6))       // value length
	_ = binary.Write(expected, binary.BigEndian, []byte("bar")) // value
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // header length

	_ = binary.Write(expected, binary.BigEndian, int8(0)) // Partitions tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0)) // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0)) // tag buffer

	require.Equal(t, expected.Bytes(), b)
}

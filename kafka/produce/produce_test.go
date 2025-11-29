package produce_test

import (
	"bytes"
	"encoding/binary"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/produce"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.Produce]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(9), reg.MaxVersion)
}

func TestRequest(t *testing.T) {
	kafkatest.TestRequest(t, 8, &produce.Request{
		TransactionalId: "",
		Acks:            0,
		TimeoutMs:       12,
		Topics: []produce.RequestTopic{
			{
				Name: "foo",
				Partitions: []produce.RequestPartition{
					{
						Index: 0,
						Record: kafka.RecordBatch{
							Records: []*kafka.Record{
								{
									Offset:         0,
									Time:           kafka.ToTime(1657010762684),
									Key:            kafka.NewBytes([]byte("foo")),
									Value:          kafka.NewBytes([]byte("bar")),
									ProducerId:     -1,
									ProducerEpoch:  -1,
									SequenceNumber: -1,
									Headers:        nil,
								},
							},
						},
					},
				},
			},
		},
	})

	kafkatest.TestRequest(t, 9, &produce.Request{
		TransactionalId: "",
		Acks:            0,
		TimeoutMs:       12,
		Topics: []produce.RequestTopic{
			{
				Name: "foo",
				Partitions: []produce.RequestPartition{
					{
						Index: 1,
						Record: kafka.RecordBatch{
							Records: []*kafka.Record{
								{
									Offset:         0,
									Time:           kafka.ToTime(1657010762684),
									Key:            kafka.NewBytes([]byte("foo")),
									Value:          kafka.NewBytes([]byte("bar")),
									ProducerId:     -1,
									ProducerEpoch:  -1,
									SequenceNumber: -1,
									Headers:        nil,
								},
							},
						},
					},
				},
			},
		},
	})

	b := kafkatest.WriteRequest(t, 9, 123, "me", &produce.Request{
		TransactionalId: "",
		Acks:            0,
		TimeoutMs:       12,
		Topics: []produce.RequestTopic{
			{
				Name: "foo",
				Partitions: []produce.RequestPartition{
					{
						Index: 1,
						Record: kafka.RecordBatch{
							Records: []*kafka.Record{
								{
									Offset:         0,
									Time:           kafka.ToTime(1657010762684),
									Key:            kafka.NewBytes([]byte("foo")),
									Value:          kafka.NewBytes([]byte("bar")),
									ProducerId:     -1,
									ProducerEpoch:  -1,
									SequenceNumber: -1,
									Headers:        nil,
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
	_ = binary.Write(expected, binary.BigEndian, int32(108))           // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.Produce)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(9))             // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))           // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))             // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))         // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))              // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // TransactionalId length
	_ = binary.Write(expected, binary.BigEndian, int16(0))      // Acks
	_ = binary.Write(expected, binary.BigEndian, int32(12))     // TimeoutMs
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // Name
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Partitions length
	_ = binary.Write(expected, binary.BigEndian, int32(1))      // Index

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

	_ = binary.Write(expected, binary.BigEndian, int8(0)) // Paritions tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0)) // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0)) // tag buffer
	require.Equal(t, expected.Bytes(), b)
}

func TestResponse(t *testing.T) {
	kafkatest.TestResponse(t, 8, &produce.Response{
		Topics: []produce.ResponseTopic{
			{
				Name: "foo",
				Partitions: []produce.ResponsePartition{
					{
						Index:          1,
						ErrorCode:      0,
						BaseOffset:     1,
						LogAppendTime:  0,
						LogStartOffset: 0,
						RecordErrors:   nil,
						ErrorMessage:   "",
					},
				},
			},
		},
		ThrottleTimeMs: 0,
	})

	kafkatest.TestResponse(t, 9, &produce.Response{
		Topics: []produce.ResponseTopic{
			{
				Name: "foo",
				Partitions: []produce.ResponsePartition{
					{
						Index:          1,
						ErrorCode:      0,
						BaseOffset:     1,
						LogAppendTime:  0,
						LogStartOffset: 0,
						RecordErrors:   nil,
						ErrorMessage:   "",
					},
				},
			},
		},
		ThrottleTimeMs: 0,
	})

	b := kafkatest.WriteResponse(t, 9, 123, &produce.Response{
		Topics: []produce.ResponseTopic{
			{
				Name: "foo",
				Partitions: []produce.ResponsePartition{
					{
						Index:          1,
						ErrorCode:      0,
						BaseOffset:     1,
						LogAppendTime:  0,
						LogStartOffset: 0,
						RecordErrors:   nil,
						ErrorMessage:   "",
					},
				},
			},
		},
		ThrottleTimeMs: 0,
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(50))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Topics length
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // Name length
	_ = binary.Write(expected, binary.BigEndian, []byte("foo")) // Name
	_ = binary.Write(expected, binary.BigEndian, int8(2))       // Partitions length
	_ = binary.Write(expected, binary.BigEndian, int32(1))      // Index
	_ = binary.Write(expected, binary.BigEndian, int16(0))      // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int64(1))      // BaseOffset
	_ = binary.Write(expected, binary.BigEndian, int64(0))      // LogAppendTime
	_ = binary.Write(expected, binary.BigEndian, int64(0))      // LogStartOffset
	_ = binary.Write(expected, binary.BigEndian, int8(1))       // RecordErrors length
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // ErrorMessage length
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // Partitions tag buffer
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // Topics tag buffer
	_ = binary.Write(expected, binary.BigEndian, int32(0))      // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	require.Equal(t, expected.Bytes(), b)
}

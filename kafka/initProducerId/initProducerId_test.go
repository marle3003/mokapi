package initProducerId_test

import (
	"bytes"
	"encoding/binary"
	"mokapi/kafka"
	"mokapi/kafka/initProducerId"
	"mokapi/kafka/kafkatest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	reg := kafka.ApiTypes[kafka.InitProducerId]
	require.Equal(t, int16(0), reg.MinVersion)
	require.Equal(t, int16(6), reg.MaxVersion)
}

func TestRequest(t *testing.T) {
	kafkatest.TestRequest(t, 2, &initProducerId.Request{
		TransactionalId:      "trx",
		TransactionTimeoutMs: 100,
	})

	kafkatest.TestRequest(t, 6, &initProducerId.Request{
		TransactionalId:      "trx",
		TransactionTimeoutMs: 100,
		ProducerId:           123,
		ProducerEpoch:        1,
		Enable2PC:            false,
	})

	b := kafkatest.WriteRequest(t, 6, 123, "me", &initProducerId.Request{
		TransactionalId:      "trx",
		TransactionTimeoutMs: 100,
		ProducerId:           123,
		ProducerEpoch:        1,
		Enable2PC:            false,
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(33))                   // length
	_ = binary.Write(expected, binary.BigEndian, int16(kafka.InitProducerId)) // ApiKey
	_ = binary.Write(expected, binary.BigEndian, int16(6))                    // ApiVersion
	_ = binary.Write(expected, binary.BigEndian, int32(123))                  // correlationId
	_ = binary.Write(expected, binary.BigEndian, int16(2))                    // ClientId length
	_ = binary.Write(expected, binary.BigEndian, []byte("me"))                // ClientId
	_ = binary.Write(expected, binary.BigEndian, int8(0))                     // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int8(4))       // TransactionalId length
	_ = binary.Write(expected, binary.BigEndian, []byte("trx")) // TransactionalId
	_ = binary.Write(expected, binary.BigEndian, int32(100))    // TransactionTimeoutMs
	_ = binary.Write(expected, binary.BigEndian, int64(123))    // ProducerId
	_ = binary.Write(expected, binary.BigEndian, int16(1))      // ProducerEpoch
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // Enable2PC
	_ = binary.Write(expected, binary.BigEndian, int8(0))       // tag buffer
	require.Equal(t, expected.Bytes(), b)
}

func TestResponse(t *testing.T) {
	kafkatest.TestResponse(t, 2, &initProducerId.Response{
		ThrottleTimeMs: 100,
		ErrorCode:      0,
		ProducerId:     123,
		ProducerEpoch:  1,
	})

	kafkatest.TestResponse(t, 6, &initProducerId.Response{
		ThrottleTimeMs:          100,
		ErrorCode:               0,
		ProducerId:              123,
		ProducerEpoch:           1,
		OngoingTxnProducerId:    124,
		OngoingTxnProducerEpoch: 2,
	})

	b := kafkatest.WriteResponse(t, 2, 123, &initProducerId.Response{
		ThrottleTimeMs:          100,
		ErrorCode:               0,
		ProducerId:              123,
		ProducerEpoch:           1,
		OngoingTxnProducerId:    124,
		OngoingTxnProducerEpoch: 2,
	})
	expected := new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(20))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(100)) // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int16(0))   // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int64(123)) // ProducerId
	_ = binary.Write(expected, binary.BigEndian, int16(1))   // ProducerEpoch

	require.Equal(t, expected.Bytes(), b)

	b = kafkatest.WriteResponse(t, 6, 123, &initProducerId.Response{
		ThrottleTimeMs:          100,
		ErrorCode:               0,
		ProducerId:              123,
		ProducerEpoch:           1,
		OngoingTxnProducerId:    124,
		OngoingTxnProducerEpoch: 2,
	})
	expected = new(bytes.Buffer)
	// header
	_ = binary.Write(expected, binary.BigEndian, int32(32))  // length
	_ = binary.Write(expected, binary.BigEndian, int32(123)) // correlationId
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer
	// message
	_ = binary.Write(expected, binary.BigEndian, int32(100)) // ThrottleTimeMs
	_ = binary.Write(expected, binary.BigEndian, int16(0))   // ErrorCode
	_ = binary.Write(expected, binary.BigEndian, int64(123)) // ProducerId
	_ = binary.Write(expected, binary.BigEndian, int16(1))   // ProducerEpoch
	_ = binary.Write(expected, binary.BigEndian, int64(124)) // OngoingTxnProducerId
	_ = binary.Write(expected, binary.BigEndian, int16(2))   // OngoingTxnProducerEpoch
	_ = binary.Write(expected, binary.BigEndian, int8(0))    // tag buffer

	require.Equal(t, expected.Bytes(), b)
}

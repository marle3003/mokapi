package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/apiVersion"
	"mokapi/kafka/kafkatest"
	"net"
	"testing"
	"time"
)

func TestApiVersion(t *testing.T) {
	s := store.New(asyncapitest.NewConfig(), enginetest.NewEngine())
	defer s.Close()

	rr := kafkatest.NewRecorder()
	s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &apiVersion.Request{
		ClientSwName:    "kafkatest",
		ClientSwVersion: "1.0",
	}))

	res, ok := rr.Message.(*apiVersion.Response)
	require.True(t, ok)
	require.Equal(t, kafka.None, res.ErrorCode)
	require.Equal(t, len(kafka.ApiTypes), len(res.ApiKeys))

	for _, a := range res.ApiKeys {
		match, ok := kafka.ApiTypes[a.ApiKey]
		require.True(t, ok, "api key is defined")
		require.Equal(t, match.MinVersion, a.MinVersion, "%v min version exp: %v, got: %v", a.ApiKey, match.MinVersion, a.MinVersion)
		require.Equal(t, match.MaxVersion, a.MaxVersion, "%v max version exp: %v, got: %v", a.ApiKey, match.MaxVersion, a.MaxVersion)
	}
}

func TestApiVersion_Raw(t *testing.T) {
	s := store.New(asyncapitest.NewConfig(), enginetest.NewEngine())
	defer s.Close()
	b := kafkatest.NewBroker(kafkatest.WithHandler(s))
	defer b.Close()

	var err error
	var conn net.Conn
	for i := 0; i < 10; i++ {
		d := net.Dialer{Timeout: time.Second * 10}
		conn, err = d.Dial("tcp", b.Addr)
		if err != nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}
	}
	require.NoError(t, err)
	defer conn.Close()

	r := &kafka.Request{
		Header: &kafka.Header{
			ApiKey:     kafka.ApiVersions,
			ApiVersion: 0,
		},
		Message: &apiVersion.Request{},
	}
	r.Write(conn)

	buf := make([]byte, 256)
	conn.Read(buf)

	// compare the first few bytes
	expect := []byte{
		0, 0, 0, 88, // length
		0, 0, 0, 0, // Correlation
		0, 0, // Error Code
		0, 0, 0, 13, // length of array

		0, 0, // Produce
		0, 0, // min
		0, 8, // max

		0, 1, // Fetch
		0, 0, // min
		0, 11, // max

		0, 2, // Offset
		0, 0, // min
		0, 7, // max
	}

	require.Equal(t, expect, buf[0:len(expect)])
}

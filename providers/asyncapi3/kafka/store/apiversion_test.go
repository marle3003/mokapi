package store_test

import (
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/apiVersion"
	"mokapi/kafka/kafkatest"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime/events/eventstest"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestApiVersion(t *testing.T) {
	s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngine(), &eventstest.Handler{})
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
	s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngine(), &eventstest.Handler{})
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
	defer func() { _ = conn.Close() }()

	r := &kafka.Request{
		Header: &kafka.Header{
			ApiKey:     kafka.ApiVersions,
			ApiVersion: 0,
		},
		Message: &apiVersion.Request{},
	}
	err = r.Write(conn)
	require.NoError(t, err)

	buf := make([]byte, 256)
	_, err = conn.Read(buf)
	require.NoError(t, err)

	// compare the first few bytes
	expect := []byte{
		0, 0, 0, 0x5e, // length
		0, 0, 0, 0, // Correlation
		0, 0, // Error Code
		0, 0, 0, 14, // length of array

		0, 0, // Produce
		0, 0, // min
		0, 9, // max

		0, 1, // Fetch
		0, 0, // min
		0, 12, // max

		0, 2, // Offset
		0, 0, // min
		0, 8, // max
	}

	require.Equal(t, expect, buf[0:len(expect)])
}

func TestApiVersion_Client_Is_Ahead(t *testing.T) {
	s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngine(), &eventstest.Handler{})
	defer s.Close()

	r := kafkatest.NewRequest("kafkatest", 30, &apiVersion.Request{
		ClientSwName:    "kafkatest",
		ClientSwVersion: "1.0",
	})
	rr := kafkatest.NewRecorder()
	s.ServeMessage(rr, r)

	// handler should change the version to zero
	require.Equal(t, int16(0), r.Header.ApiVersion)

	res, ok := rr.Message.(*apiVersion.Response)
	require.True(t, ok)
	require.Equal(t, kafka.UnsupportedVersion, res.ErrorCode)
	require.Equal(t, len(kafka.ApiTypes), len(res.ApiKeys))

	for _, a := range res.ApiKeys {
		match, ok := kafka.ApiTypes[a.ApiKey]
		require.True(t, ok, "api key is defined")
		require.Equal(t, match.MinVersion, a.MinVersion, "%v min version exp: %v, got: %v", a.ApiKey, match.MinVersion, a.MinVersion)
		require.Equal(t, match.MaxVersion, a.MaxVersion, "%v max version exp: %v, got: %v", a.ApiKey, match.MaxVersion, a.MaxVersion)
	}
}

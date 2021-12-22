package kafka_test

import (
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/apiVersion"
	"mokapi/test"
	"net"
	"testing"
)

func TestApiVersion(t *testing.T) {
	b := kafkatest.NewBroker()
	defer b.Close()

	r, err := b.Client().ApiVersion(3, &apiVersion.Request{
		ClientSwName:    "kafkatest",
		ClientSwVersion: "1.0",
	})
	test.Ok(t, err)
	test.Equals(t, protocol.None, r.ErrorCode)
	test.Equals(t, len(protocol.ApiTypes), len(r.ApiKeys))

	for _, a := range r.ApiKeys {
		match, ok := protocol.ApiTypes[a.ApiKey]
		test.Assert(t, ok, "api key is defined")
		test.Assert(t, match.MinVersion == a.MinVersion, "%v min version exp: %v, got: %v", a.ApiKey, match.MinVersion, a.MinVersion)
		test.Assert(t, match.MaxVersion == a.MaxVersion, "%v max version exp: %v, got: %v", a.ApiKey, match.MaxVersion, a.MaxVersion)
	}
}

func TestApiVersion_Raw(t *testing.T) {
	b := kafkatest.NewBroker()
	defer b.Close()

	d := net.Dialer{}
	conn, err := d.Dial("tcp", b.Listener.Addr().String())
	defer conn.Close()
	test.Ok(t, err)

	r := &protocol.Request{
		Header: &protocol.Header{
			ApiKey:     protocol.ApiVersions,
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
		0, 6, // max
	}

	test.Equals(t, expect, buf[0:len(expect)])
}

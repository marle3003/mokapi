package kafka_test

import (
	"mokapi/server/kafka/kafkatest"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/apiVersion"
	"mokapi/test"
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

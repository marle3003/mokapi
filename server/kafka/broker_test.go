package kafka_test

import (
	"mokapi/server/kafka/kafkatest"
	"mokapi/server/kafka/protocol"
	"mokapi/test"
	"testing"
	"time"
)

func TestApiKeys(t *testing.T) {
	t.Parallel()
	for k, at := range protocol.ApiTypes {
		key := k
		apiType := at
		t.Run(k.String(), func(t *testing.T) {
			t.Parallel()
			b := kafkatest.NewBroker()
			defer b.Close()
			c := b.Client()
			c.Timeout = 3 * time.Second

			r, err := c.Send(&protocol.Request{
				Header: &protocol.Header{
					ApiKey:     key,
					ApiVersion: apiType.MinVersion,
				},
				Message: kafkatest.GetRequest(key),
			})
			test.Ok(t, err)
			test.Equals(t, key, r.Header.ApiKey)
		})
	}
}

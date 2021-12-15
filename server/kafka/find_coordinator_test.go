package kafka_test

import (
	"mokapi/server/kafka/kafkatest"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/findCoordinator"
	"mokapi/test"
	"testing"
)

func TestFindCoordinator(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, b *kafkatest.Broker)
	}{{
		"find group",
		func(t *testing.T, b *kafkatest.Broker) {
			r, err := b.Client().FindCoordinator(3, &findCoordinator.Request{
				Key:     "foo",
				KeyType: findCoordinator.KeyTypeGroup,
			})
			test.Ok(t, err)
			test.Equals(t, protocol.CoordinatorNotAvailable, r.ErrorCode)
		},
	}}

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

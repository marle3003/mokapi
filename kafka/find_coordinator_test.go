package kafka_test

import (
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/findCoordinator"
	"mokapi/test"
	"testing"
)

func TestFindCoordinator(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, b *kafkatest.Broker)
	}{
		{
			"find group",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{
					Brokers: []string{b.Listener.Addr().String()}}))
				r, err := b.Client().FindCoordinator(3, &findCoordinator.Request{
					Key:     "foo",
					KeyType: findCoordinator.KeyTypeGroup,
				})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)

				host, port := b.HostPort()
				test.Equals(t, host, r.Host)
				test.Equals(t, int32(port), r.Port)
			},
		},
		{
			"unsupported key type",
			func(t *testing.T, b *kafkatest.Broker) {
				r, err := b.Client().FindCoordinator(3, &findCoordinator.Request{
					Key:     "foo",
					KeyType: 99,
				})
				test.Ok(t, err)
				test.Equals(t, protocol.Unknown, r.ErrorCode)
			},
		},
	}

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

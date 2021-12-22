package kafka_test

import (
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/listgroup"
	"mokapi/test"
	"testing"
)

func TestListGroup(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, b *kafkatest.Broker)
	}{
		{
			"empty",
			func(t *testing.T, b *kafkatest.Broker) {
				r, err := b.Client().Listgroup(3, &listgroup.Request{})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, 0, len(r.Groups))
			},
		},
	}

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

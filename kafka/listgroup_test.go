package kafka_test

import (
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/listgroup"
	"mokapi/kafka/store"
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
		{
			"with group state",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{Brokers: []string{b.Listener.Addr().String()}}))
				group := b.Store().GetOrCreateGroup("foo", 0)
				group.SetState(store.Joining)
				g := group.NewGeneration()
				g.Members[""] = &store.Member{}

				r, err := b.Client().Listgroup(4, &listgroup.Request{})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, 1, len(r.Groups))
				test.Equals(t, "PreparingRebalance", r.Groups[0].GroupState)
			},
		},
		{
			"filtering",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(kafkatest.NewStore(kafkatest.StoreConfig{Brokers: []string{b.Listener.Addr().String()}}))
				b.Store().GetOrCreateGroup("foo", 0)
				group := b.Store().GetOrCreateGroup("bar", 0)
				group.SetState(store.AwaitingSync)
				g := group.NewGeneration()
				g.Members[""] = &store.Member{}

				r, err := b.Client().Listgroup(4, &listgroup.Request{StatesFilter: []string{"Empty"}})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
				test.Equals(t, 1, len(r.Groups))
				test.Equals(t, "foo", r.Groups[0].GroupId)
				test.Equals(t, "Empty", r.Groups[0].GroupState)
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

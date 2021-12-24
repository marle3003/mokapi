package kafka_test

import (
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/heartbeat"
	"mokapi/kafka/protocol/joinGroup"
	"mokapi/kafka/schema"
	"mokapi/kafka/store"
	"mokapi/test"
	"testing"
)

func TestHeartbeat(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, b *kafkatest.Broker)
	}{
		{
			"not in group",
			func(t *testing.T, b *kafkatest.Broker) {
				r, err := b.Client().Heartbeat(3, &heartbeat.Request{
					GroupId:  "foo",
					MemberId: "bar",
				})
				test.Ok(t, err)
				test.Equals(t, protocol.UnknownMemberId, r.ErrorCode)
			},
		},
		{
			"group balancing",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(store.New(schema.Cluster{Topics: []schema.Topic{
					{Name: "foo"},
				}, Brokers: []schema.Broker{schema.NewBroker(0, b.Listener.Addr().String())}}))
				j, err := b.Client().JoinGroup(3, &joinGroup.Request{GroupId: "foo", MemberId: "bar"})
				test.Ok(t, err)
				test.Equals(t, protocol.None, j.ErrorCode)

				r, err := b.Client().Heartbeat(3, &heartbeat.Request{
					GroupId:  "foo",
					MemberId: "bar",
				})
				test.Ok(t, err)
				test.Equals(t, protocol.RebalanceInProgress, r.ErrorCode)
			},
		},
		{
			"ok",
			func(t *testing.T, b *kafkatest.Broker) {
				b.SetStore(store.New(schema.Cluster{Brokers: []schema.Broker{schema.NewBroker(0, b.Listener.Addr().String())}}))
				err := b.Client().JoinSyncGroup("foo", "TestGroup", 3, 3)
				test.Ok(t, err)
				r, err := b.Client().Heartbeat(3, &heartbeat.Request{
					GroupId:         "TestGroup",
					GenerationId:    0,
					MemberId:        "foo",
					GroupInstanceId: "",
				})
				test.Ok(t, err)
				test.Equals(t, protocol.None, r.ErrorCode)
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

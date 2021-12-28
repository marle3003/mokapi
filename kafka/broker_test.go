package kafka_test

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/apiVersion"
	"mokapi/kafka/store"
	"mokapi/test"
	"net"
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
			cfg := asyncapitest.NewConfig(asyncapitest.WithServer("test", "kafka", b.Listener.Addr().String()))
			b.SetStore(store.New(cfg))
			c := b.Client()
			// todo lower timeout when group.initial.rebalance.delay.ms is configurable
			c.Timeout = 5 * time.Second

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

func TestBroker_Disconnect(t *testing.T) {
	hook := test.NewNullLogger()
	b := kafkatest.NewBroker()
	d := net.Dialer{}
	conn, err := d.Dial("tcp", b.Listener.Addr().String())
	test.Ok(t, err)

	r := &protocol.Request{
		Header: &protocol.Header{
			ApiKey:     protocol.ApiVersions,
			ApiVersion: 0,
		},
		Message: &apiVersion.Request{},
	}

	r.Write(conn)
	conn.Close()
	time.Sleep(1000 * time.Millisecond)
	// should not log any panic message
	test.Equals(t, nil, hook.LastEntry())
}

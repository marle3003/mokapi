package store_test

import (
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"testing"

	"github.com/stretchr/testify/require"
)

func newStoreWithServerHost(host string) *store.Store {
	return newStoreFromConfig(asyncapi3test.NewConfig(asyncapi3test.WithServer("a", "kafka", host)))
}

func TestBroker_Addr(t *testing.T) {
	s := newStoreWithServerHost("bar:9092")
	defer s.Close()

	require.Equal(t, "bar:9092", s.Brokers()[0].Addr())
}

func TestBroker_HostPort(t *testing.T) {
	t.Run("from config", func(t *testing.T) {
		s := newStoreWithServerHost("bar:9092")
		defer s.Close()

		host, port := s.Brokers()[0].HostPort()
		require.Equal(t, "bar", host)
		require.Equal(t, 9092, port)
	})

	t.Run("host and port change with config update", func(t *testing.T) {
		s := newStoreWithServerHost("bar:9092")
		defer s.Close()
		b := s.Brokers()[0]

		s.Update(asyncapi3test.NewConfig(asyncapi3test.WithServer("a", "kafka", "foo:9093")))

		host, port := b.HostPort()
		require.Equal(t, "foo", host)
		require.Equal(t, 9093, port)
		require.Equal(t, "foo:9093", b.Addr())
		require.Same(t, b, s.BrokerByPort("foo:9093"))
		require.Nil(t, s.BrokerByPort("foo:9092"))
	})

	t.Run("zero value", func(t *testing.T) {
		b := &store.Broker{}

		host, port := b.HostPort()
		require.Equal(t, "", host)
		require.Equal(t, 0, port)
	})
}

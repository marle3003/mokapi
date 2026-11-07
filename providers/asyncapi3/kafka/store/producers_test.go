package store

import (
	"mokapi/kafka"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore_Producers(t *testing.T) {
	t.Run("register assigns increasing ids starting with epoch 0", func(t *testing.T) {
		s := NewEmpty(nil, nil, nil)

		first := s.registerProducer()
		second := s.registerProducer()

		require.Equal(t, int64(1), first.ProducerId)
		require.Equal(t, int16(0), first.ProducerEpoch)
		require.Equal(t, int64(2), second.ProducerId)
	})

	t.Run("registered producer can be looked up", func(t *testing.T) {
		s := NewEmpty(nil, nil, nil)
		registered := s.registerProducer()

		ps, ok := s.producer(registered.ProducerId)

		require.True(t, ok)
		require.Equal(t, registered, ps)
	})

	t.Run("unknown producer", func(t *testing.T) {
		s := NewEmpty(nil, nil, nil)

		_, ok := s.producer(42)

		require.False(t, ok)
	})

	t.Run("renew increments epoch", func(t *testing.T) {
		s := NewEmpty(nil, nil, nil)
		id := s.registerProducer().ProducerId

		ps, code := s.renewProducer(id, 0)

		require.Equal(t, kafka.None, code)
		require.Equal(t, ProducerState{ProducerId: id, ProducerEpoch: 1}, ps)
		stored, _ := s.producer(id)
		require.Equal(t, int16(1), stored.ProducerEpoch)
	})

	t.Run("renew unknown producer", func(t *testing.T) {
		s := NewEmpty(nil, nil, nil)

		_, code := s.renewProducer(42, 0)

		require.Equal(t, kafka.UnknownProducerId, code)
	})

	t.Run("renew with lower epoch is fenced", func(t *testing.T) {
		s := NewEmpty(nil, nil, nil)
		id := s.registerProducer().ProducerId
		_, _ = s.renewProducer(id, 0)

		_, code := s.renewProducer(id, 0)

		require.Equal(t, kafka.ProducerFenced, code)
		stored, _ := s.producer(id)
		require.Equal(t, int16(1), stored.ProducerEpoch)
	})

	t.Run("producer returns a copy", func(t *testing.T) {
		s := NewEmpty(nil, nil, nil)
		id := s.registerProducer().ProducerId

		ps, _ := s.producer(id)
		ps.ProducerEpoch = 7

		stored, _ := s.producer(id)
		require.Equal(t, int16(0), stored.ProducerEpoch)
	})
}

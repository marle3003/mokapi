package store

import (
	"mokapi/kafka"
	"sync/atomic"
)

// registerProducer creates a producer with a new id and epoch 0.
func (s *Store) registerProducer() ProducerState {
	ps := ProducerState{ProducerId: atomic.AddInt64(&s.nextPID, 1)}

	s.producersM.Lock()
	defer s.producersM.Unlock()
	s.producers[ps.ProducerId] = &ps
	return ps
}

// renewProducer increments the epoch of a known producer. It returns the updated state or
// an error code if the producer is unknown or the given epoch is outdated.
func (s *Store) renewProducer(id int64, epoch int16) (ProducerState, kafka.ErrorCode) {
	s.producersM.Lock()
	defer s.producersM.Unlock()

	ps, ok := s.producers[id]
	if !ok {
		return ProducerState{}, kafka.UnknownProducerId
	}
	if epoch < ps.ProducerEpoch {
		return ProducerState{}, kafka.ProducerFenced
	}
	ps.ProducerEpoch++
	return *ps, kafka.None
}

// producer returns a copy of the producer state, so callers can use it without holding producersM.
func (s *Store) producer(id int64) (ProducerState, bool) {
	s.producersM.RLock()
	defer s.producersM.RUnlock()

	ps, ok := s.producers[id]
	if !ok {
		return ProducerState{}, false
	}
	return *ps, true
}

package store

// NumSegments and IsSegmentClosed let tests in package store_test observe segments
// while the log cleaner is running.
func (p *Partition) NumSegments() int {
	p.m.RLock()
	defer p.m.RUnlock()
	return len(p.Segments)
}

func (p *Partition) IsSegmentClosed(head int64) bool {
	p.m.RLock()
	defer p.m.RUnlock()
	s, ok := p.Segments[head]
	return ok && !s.Closed.IsZero()
}

// BrokerByPort lets tests look up a broker the way request handlers do.
func (s *Store) BrokerByPort(addr string) *Broker {
	return s.getBrokerByPort(addr)
}

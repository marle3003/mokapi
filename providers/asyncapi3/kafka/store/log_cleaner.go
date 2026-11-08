package store

import (
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

func (s *Store) cleanLog(b *Broker) {
	brokerRetentionMs := b.kafkaConfig.LogRetentionMs
	if brokerRetentionMs == 0 {
		brokerRetentionMs = 604800000 // 7 days
	}
	brokerRetentionBytes := b.kafkaConfig.LogRetentionBytes
	if brokerRetentionBytes == 0 {
		brokerRetentionBytes = -1
	}
	brokerRollingMs := b.kafkaConfig.LogRollMs
	if brokerRollingMs == 0 {
		brokerRollingMs = 604800000 // 7 days
	}
	now := time.Now()

	for _, topic := range s.Topics() {
		channel, partitions := topic.channelAndPartitions()

		retentionTime := time.Duration(brokerRetentionMs) * time.Millisecond
		retentionBytes := brokerRetentionBytes
		rollingTime := time.Duration(brokerRollingMs) * time.Millisecond

		if channel.Bindings != nil {
			if channel.Bindings.Kafka.RetentionMs > 0 {
				retentionTime = time.Duration(channel.Bindings.Kafka.RetentionMs) * time.Millisecond
			}
			if channel.Bindings.Kafka.RetentionBytes > 0 {
				retentionBytes = channel.Bindings.Kafka.RetentionBytes
			}
			if channel.Bindings.Kafka.SegmentMs > 0 {
				rollingTime = time.Duration(channel.Bindings.Kafka.SegmentMs) * time.Millisecond
			}
		}

		for _, p := range partitions {
			if p.leader.Id != b.Id {
				continue
			}
			cleanPartition(p, now, rollingTime, retentionTime, retentionBytes)
		}
	}
}

// cleanPartition holds only the partition lock, never the store lock.
func cleanPartition(p *Partition, now time.Time, rollingTime, retentionTime time.Duration, retentionBytes int64) {
	p.m.Lock()
	defer p.m.Unlock()

	partitionSize := int64(0)
	for _, segment := range p.Segments {
		partitionSize += int64(segment.Size)

		// check rolling
		if segment.Closed.IsZero() && now.After(segment.Opened.Add(rollingTime)) {
			// Close the current segment.
			// Segment creation is lazy — it happens when a write comes in.
			segment.Closed = now
		}

		// check retention
		if segment.Size > 0 && !segment.Closed.IsZero() && now.After(segment.Closed.Add(retentionTime)) {
			log.Infof(
				"kafka: deleting segment with offset [%v:%v] from partition %v topic '%s'",
				segment.Head, segment.Tail, p.Index, p.Topic.Name)
			p.removeSegmentLocked(segment)
		}
	}

	if retentionBytes > 0 && partitionSize >= retentionBytes {
		deleteSegmentsUntilLogIsWithinLimit(p)
	}
}

// deleteSegmentsUntilLogIsWithinLimit requires the partition lock.
func deleteSegmentsUntilLogIsWithinLimit(p *Partition) {
	// Build a slice of segments
	var segments []*Segment
	for _, s := range p.Segments {
		if !s.Closed.IsZero() {
			segments = append(segments, s)
		}
	}

	// Sort by oldest first (e.g., Head offset)
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].Head < segments[j].Head
	})

	for _, s := range segments {
		if !s.Closed.IsZero() {
			log.Infof("kafka: maximum partition size reached. deleting segment [%v:%v] from partition %v of topic '%s'",
				s.Head, s.Tail, p.Index, p.Topic.Name)
			p.removeSegmentLocked(s)
		}
	}
}

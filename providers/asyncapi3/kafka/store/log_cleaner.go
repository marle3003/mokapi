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

	for _, topic := range s.topics {
		retentionTime := time.Duration(brokerRetentionMs) * time.Millisecond
		retentionBytes := brokerRetentionBytes
		rollingTime := time.Duration(brokerRollingMs) * time.Millisecond

		if topic.Config.Bindings.Kafka.RetentionMs > 0 {
			retentionTime = time.Duration(topic.Config.Bindings.Kafka.RetentionMs) * time.Millisecond
		}
		if topic.Config.Bindings.Kafka.RetentionBytes > 0 {
			retentionBytes = topic.Config.Bindings.Kafka.RetentionBytes
		}
		if topic.Config.Bindings.Kafka.SegmentMs > 0 {
			rollingTime = time.Duration(topic.Config.Bindings.Kafka.SegmentMs) * time.Millisecond
		}

		for _, p := range topic.Partitions {
			if p.leader.Id != b.Id {
				continue
			}

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
						segment.Head, segment.Tail, p.Index, topic.Name)
					p.removeSegment(segment)
				}
			}

			if retentionBytes > 0 && partitionSize >= retentionBytes {
				deleteSegmentsUntilLogIsWithinLimit(p)
			}
		}
	}
}

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
			p.removeSegment(s)
		}
	}
}

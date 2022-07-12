package store

import (
	log "github.com/sirupsen/logrus"
	"time"
)

func (s *Store) cleanLog(b *Broker) {
	brokerRetentionTime := time.Duration(b.kafkaConfig.LogRetentionMs()) * time.Millisecond
	brokerRetentionBytes := b.kafkaConfig.LogRetentionBytes()
	brokerRollingTime := time.Duration(b.kafkaConfig.LogRollMs()) * time.Millisecond
	now := time.Now()

	for _, topic := range s.topics {
		retentionTime := brokerRetentionTime
		retentionBytes := brokerRetentionBytes
		rollingTime := brokerRollingTime

		if ms, ok := topic.kafkaConfig.RetentionMs(); ok {
			retentionTime = time.Duration(ms) * time.Millisecond
		}
		if bytes, ok := topic.kafkaConfig.RetentionBytes(); ok {
			retentionBytes = bytes
		}
		if ms, ok := topic.kafkaConfig.SegmentMs(); ok {
			rollingTime = time.Duration(ms) * time.Millisecond
		}

		for _, p := range topic.Partitions {
			if p.Leader != b.Id {
				continue
			}

			partitionSize := int64(0)
			for _, segment := range p.Segments {
				partitionSize += int64(segment.Size)

				// check rolling
				if segment.Closed.IsZero() && now.After(segment.Opened.Add(rollingTime)) {
					p.addSegment()
				}

				// check retention
				if segment.Size > 0 && !segment.Closed.IsZero() && now.After(segment.Closed.Add(retentionTime)) {
					log.Infof(
						"kafka: deleting segment with offset [%v:%v] from partition %v topic %q",
						segment.Head, segment.Tail, p.Index, topic.Name)
					p.removeSegment(segment)
				}
			}

			if retentionBytes > 0 && partitionSize >= retentionBytes {
				log.Infof("kafka: maximum partition size reached. cleanup partition %v from topic %q",
					p.Index, topic.Name)
				p.removeClosedSegments()
			}
		}
	}
}

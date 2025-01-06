package store

import (
	log "github.com/sirupsen/logrus"
	"time"
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

		if topic.channel.Bindings.Kafka.RetentionMs > 0 {
			retentionTime = time.Duration(topic.channel.Bindings.Kafka.RetentionMs) * time.Millisecond
		}
		if topic.channel.Bindings.Kafka.RetentionBytes > 0 {
			retentionBytes = topic.channel.Bindings.Kafka.RetentionBytes
		}
		if topic.channel.Bindings.Kafka.SegmentMs > 0 {
			rollingTime = time.Duration(topic.channel.Bindings.Kafka.SegmentMs) * time.Millisecond
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

func valueOrDefault(i int64, d int64) int64 {
	if i == 0 {
		return d
	}
	return i
}

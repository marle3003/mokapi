package models

import (
	"crypto/rand"
	"io"
	"runtime"
	"time"
)

type Metrics struct {
	Start             time.Time
	TotalRequests     int64
	RequestsWithError int64
	LastRequests      []*RequestMetric
	LastErrorRequests []*RequestMetric
	Memory            int64
	Kafka             *KafkaMetric
	OpenApi           map[string]*ServiceMetric

	TotalMails int64
	LastMails  []*Mail
}

type KafkaMetric struct {
	Topics map[string]KafkaTopic
}

type KafkaTopic struct {
	Name       string         `json:"name"`
	Count      int64          `json:"count"`
	Size       int64          `json:"size"`
	LastRecord time.Time      `json:"lastRecord"`
	Partitions int            `json:"partitions"`
	Segments   int            `json:"segments"`
	Messages   []KafkaMessage `json:"messages"`
	//Groups []string `json:"groups"`
}

type KafkaMessage struct {
	Key   string `json:"name"`
	Value string `json:"value"`
}

func newMetrics() *Metrics {
	return &Metrics{LastRequests: make([]*RequestMetric, 0), Start: time.Now(), Kafka: &KafkaMetric{Topics: make(map[string]KafkaTopic)}, OpenApi: make(map[string]*ServiceMetric)}
}

func (m *Metrics) AddMessage(topic string, key []byte, value []byte) {
	msg := KafkaMessage{Key: string(key), Value: string(value)}
	if _, ok := m.Kafka.Topics[topic]; !ok {
		m.Kafka.Topics[topic] = KafkaTopic{Name: topic}
	}
	t := m.Kafka.Topics[topic]
	t.Messages = append(t.Messages, msg)
	if len(t.Messages) > 10 {
		t.Messages = t.Messages[1:]
	}
}

func (m *Metrics) Update() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	m.Memory = int64(stats.Alloc)
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

func newId(length int) string {
	b := make([]byte, length)
	n, err := io.ReadAtLeast(rand.Reader, b, length)
	if n != length {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

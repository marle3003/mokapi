package api

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"mokapi/models"
	"net/http"
	"strings"
)

type Topic struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Messages    []models.KafkaMessage `json:"messages"`
	Partitions  []partition           `json:"partitions"`
	Groups      []topicGroup          `json:"groups"`
	Count       int64                 `json:"count"`
}

func (b *Binding) handleKafka(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")

	if len(segments) < 5 {
		w.WriteHeader(404)
		return
	}

	kafka := segments[4]

	// /api/dashboard/kafka/:name/topics/:topic
	if len(segments) >= 6 && segments[5] == "topics" {
		b.getKafkaTopic(kafka, segments[6], w, r)
		return
	}
}

func (b *Binding) getKafkaTopic(kafka string, topicName string, w http.ResponseWriter, r *http.Request) {
	s, ok := b.runtime.AsyncApi[kafka]
	if !ok {
		w.WriteHeader(404)
		return
	}

	topic, ok := s.Channels["/"+topicName]
	if !ok {
		w.WriteHeader(404)
		return
	} else if topic.Value == nil {
		w.WriteHeader(500)
		return
	}

	data := Topic{Name: topicName, Description: topic.Value.Description}

	m, ok := b.runtime.Metrics.Kafka.Topics[topicName]
	if ok {
		data.Messages = m.Messages
		data.Count = m.Count
		for _, p := range m.Partitions {
			data.Partitions = append(data.Partitions, newPartition(p))
		}
		for name, g := range m.Groups {
			data.Groups = append(data.Groups, newTopicGroup(name, g))
		}
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Errorf("Error in writing service response: %v", err.Error())
	}
}

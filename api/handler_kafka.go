package api

import (
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"net/http"
	"strings"
)

type kafkaSummary struct {
	Name        string   `json:"name"`
	Topics      []string `json:"topics"`
	LastMessage int64    `json:"lastMessage"`
	Messages    int64    `json:"messages"`
	Errors      int64    `json:"errors"`
}

func (h *handler) getKafkaServices(w http.ResponseWriter, _ *http.Request) {
	result := getKafkaServices(h.app.Kafka, h.app.Monitor.Kafka)
	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func (h *handler) getKafkaService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	if s, ok := h.app.Kafka[name]; ok {
		w.Header().Set("Content-Type", "application/json")
		writeJsonBody(w, s)
	} else {
		w.WriteHeader(404)
	}
}

func getKafkaServices(services map[string]*runtime.KafkaInfo, m *monitor.Kafka) []*kafkaSummary {
	result := make([]*kafkaSummary, 0, len(services))
	for _, hs := range services {
		k := &kafkaSummary{
			Name:     hs.Info.Name,
			Messages: int64(m.Messages.WithLabel(hs.Info.Name).Value()),
		}

		for name := range hs.Channels {
			k.Topics = append(k.Topics, name)
		}
		result = append(result, k)
	}
	return result
}

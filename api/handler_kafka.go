package api

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"net/http"
	"strings"
	"time"
)

type kafkaSummary struct {
	service
	Topics []string `json:"topics"`
}

type kafka struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Version     string           `json:"version"`
	Contact     *kafkaContact    `json:"contact"`
	Servers     []kafkaServer    `json:"servers,omitempty"`
	Topics      []topic          `json:"topics"`
	Groups      []group          `json:"groups"`
	Metrics     []metrics.Metric `json:"metrics"`
}

type kafkaServer struct {
	Name        string `json:"name"`
	Url         string `json:"url"`
	Description string `json:"description"`
}

type group struct {
	Name               string   `json:"name"`
	Members            []member `json:"members"`
	Coordinator        string   `json:"coordinator"`
	Leader             string   `json:"leader"`
	State              string   `json:"state"`
	AssignmentStrategy string   `json:"protocol"`
	Topics             []string `json:"topics"`
}

type member struct {
	Name                  string    `json:"name"`
	Addr                  string    `json:"addr"`
	ClientSoftwareName    string    `json:"clientSoftwareName"`
	ClientSoftwareVersion string    `json:"clientSoftwareVersion"`
	Heartbeat             time.Time `json:"heartbeat"`
	Partitions            []int     `json:"partitions"`
}

type kafkaContact struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Email string `json:"email"`
}

type topic struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Partitions  []partition  `json:"partitions"`
	Configs     *topicConfig `json:"configs"`
}

type partition struct {
	Id          int    `json:"id"`
	StartOffset int64  `json:"startOffset"`
	Offset      int64  `json:"offset"`
	Leader      broker `json:"leader"`
	Segments    int    `json:"segments"`
}

type broker struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
}

type topicConfig struct {
	Key         *schemaInfo `json:"key"`
	Message     *schemaInfo `json:"message"`
	Header      *schemaInfo `json:header`
	MessageType string      `json:messageType`
}

func (h *handler) getKafkaServices(w http.ResponseWriter, _ *http.Request) {
	result := getKafkaServices(h.app.Kafka, h.app.Monitor)
	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func (h *handler) getKafkaService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	if s, ok := h.app.Kafka[name]; ok {
		k := getKafka(s)
		k.Metrics = h.app.Monitor.FindAll(metrics.ByNamespace("kafka"), metrics.ByLabel("service", name))

		w.Header().Set("Content-Type", "application/json")
		writeJsonBody(w, k)
	} else {
		w.WriteHeader(404)
	}
}

func getKafkaServices(services map[string]*runtime.KafkaInfo, m *monitor.Monitor) []interface{} {
	result := make([]interface{}, 0, len(services))
	for _, hs := range services {
		k := &kafkaSummary{
			service: service{
				Name:        hs.Info.Name,
				Description: hs.Info.Description,
				Version:     hs.Info.Version,
				Type:        ServiceKafka,
				Metrics:     m.FindAll(metrics.ByNamespace("kafka"), metrics.ByLabel("service", hs.Info.Name)),
			},
		}

		for name := range hs.Channels {
			k.Topics = append(k.Topics, name)
		}
		result = append(result, k)
	}
	return result
}

func getKafka(info *runtime.KafkaInfo) kafka {
	k := kafka{
		Name:        info.Config.Info.Name,
		Description: info.Config.Info.Description,
		Version:     info.Config.Info.Version,
		Groups:      make([]group, 0),
	}

	if info.Config.Info.Contact != nil {
		k.Contact = &kafkaContact{
			Name:  info.Config.Info.Contact.Name,
			Url:   info.Config.Info.Contact.Url,
			Email: info.Config.Info.Contact.Email,
		}
	}

	for name, server := range info.Servers {
		k.Servers = append(k.Servers, kafkaServer{
			Name:        name,
			Url:         server.Url,
			Description: server.Description,
		})
	}

	for name, ch := range info.Config.Channels {
		if ch.Value == nil {
			continue
		}
		t := info.Store.Topic(name)
		k.Topics = append(k.Topics, newTopic(info.Store, t, ch.Value))
	}

	for _, g := range info.Store.Groups() {
		k.Groups = append(k.Groups, newGroup(g))
	}

	return k
}

func newTopic(s *store.Store, t *store.Topic, config *asyncApi.Channel) topic {
	var partitions []partition
	for _, p := range t.Partitions {
		partitions = append(partitions, newPartition(s, p))
	}
	result := topic{
		Name:        t.Name,
		Description: config.Description,
		Partitions:  partitions,
	}

	if config.Publish.Message.Value != nil {
		result.Configs = &topicConfig{
			Key:         getSchema(config.Publish.Message.Value.Bindings.Kafka.Key),
			Message:     getSchema(config.Publish.Message.Value.Payload),
			Header:      nil,
			MessageType: config.Publish.Message.Value.ContentType,
		}
	}

	return result
}

func newGroup(g *store.Group) group {
	grp := group{
		Name:        g.Name,
		State:       g.State.String(),
		Coordinator: g.Coordinator.Addr(),
	}
	if g.Generation != nil {
		grp.Leader = g.Generation.LeaderId
		grp.AssignmentStrategy = g.Generation.Protocol

		for id, m := range g.Generation.Members {
			var partitions []int
			for _, p := range m.Partitions {
				partitions = append(partitions, p.Index)
			}
			grp.Members = append(grp.Members, member{
				Name:                  id,
				Addr:                  m.Client.Addr,
				ClientSoftwareName:    m.Client.ClientSoftwareName,
				ClientSoftwareVersion: m.Client.ClientSoftwareVersion,
				Heartbeat:             m.Client.Heartbeat,
				Partitions:            partitions,
			})
		}
	}
	for topicName := range g.Commits {
		grp.Topics = append(grp.Topics, topicName)
	}

	return grp
}

func newPartition(s *store.Store, p *store.Partition) partition {
	leader, _ := s.Broker(p.Leader)
	return partition{
		Id:          p.Index,
		StartOffset: p.StartOffset(),
		Offset:      p.Offset(),
		Leader:      newBroker(leader),
		Segments:    len(p.Segments),
	}
}

func newBroker(b *store.Broker) broker {
	if b == nil {
		return broker{}
	}
	return broker{
		Name: b.Name,
		Addr: b.Addr(),
	}
}

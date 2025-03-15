package api

import (
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"net/http"
	"sort"
	"strings"
	"time"
)

type kafkaSummary struct {
	service
}

type kafka struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Version     string           `json:"version"`
	Contact     *kafkaContact    `json:"contact,omitempty"`
	Servers     []kafkaServer    `json:"servers,omitempty"`
	Topics      []topic          `json:"topics,omitempty"`
	Groups      []group          `json:"groups,omitempty"`
	Metrics     []metrics.Metric `json:"metrics,omitempty"`
	Configs     []config         `json:"configs,omitempty"`
}

type kafkaServer struct {
	Name        string           `json:"name"`
	Host        string           `json:"host"`
	Description string           `json:"description"`
	Tags        []kafkaServerTag `json:"tags,omitempty"`
}

type kafkaServerTag struct {
	Name        string `json:"name"`
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
	Name                  string           `json:"name"`
	Addr                  string           `json:"addr"`
	ClientSoftwareName    string           `json:"clientSoftwareName"`
	ClientSoftwareVersion string           `json:"clientSoftwareVersion"`
	Heartbeat             time.Time        `json:"heartbeat"`
	Partitions            map[string][]int `json:"partitions"`
}

type kafkaContact struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Email string `json:"email"`
}

type topic struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Partitions  []partition              `json:"partitions"`
	Messages    map[string]messageConfig `json:"messages,omitempty"`
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

type messageConfig struct {
	Name        string      `json:"name,omitempty"`
	Title       string      `json:"title,omitempty"`
	Summary     string      `json:"summary,omitempty"`
	Description string      `json:"description,omitempty"`
	Key         *schemaInfo `json:"key,omitempty"`
	Payload     *schemaInfo `json:"payload"`
	Header      *schemaInfo `json:"header,omitempty"`
	ContentType string      `json:"contentType"`
}

func getKafkaServices(store *runtime.KafkaStore, m *monitor.Monitor) []interface{} {
	list := store.List()
	result := make([]interface{}, 0, len(list))
	for _, hs := range list {
		s := service{
			Name:        hs.Info.Name,
			Description: hs.Info.Description,
			Version:     hs.Info.Version,
			Type:        ServiceKafka,
			Metrics:     m.FindAll(metrics.ByNamespace("kafka"), metrics.ByLabel("service", hs.Info.Name)),
		}

		if hs.Info.Contact != nil {
			c := hs.Info.Contact
			s.Contact = &contact{
				Name:  c.Name,
				Url:   c.Url,
				Email: c.Email,
			}
		}

		result = append(result, kafkaSummary{service: s})
	}
	return result
}

func (h *handler) getKafkaService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	if s := h.app.Kafka.Get(name); s != nil {
		k := getKafka(s)
		k.Metrics = h.app.Monitor.FindAll(metrics.ByNamespace("kafka"), metrics.ByLabel("service", name))

		w.Header().Set("Content-Type", "application/json")
		writeJsonBody(w, k)
	} else {
		w.WriteHeader(404)
	}
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

	for name, s := range info.Servers {
		if s == nil || s.Value == nil {
			continue
		}

		ks := kafkaServer{
			Name:        name,
			Host:        s.Value.Host,
			Description: s.Value.Description,
		}
		for _, r := range s.Value.Tags {
			if r.Value == nil {
				continue
			}
			t := r.Value
			ks.Tags = append(ks.Tags, kafkaServerTag{
				Name:        t.Name,
				Description: t.Description,
			})
		}
		k.Servers = append(k.Servers, ks)
	}
	sort.Slice(k.Servers, func(i, j int) bool {
		return strings.Compare(k.Servers[i].Name, k.Servers[j].Name) < 0
	})

	for name, ch := range info.Config.Channels {
		if ch.Value == nil {
			continue
		}
		t := info.Store.Topic(name)
		k.Topics = append(k.Topics, newTopic(info.Store, t, ch.Value, info.DefaultContentType))
	}
	sort.Slice(k.Topics, func(i, j int) bool {
		return strings.Compare(k.Topics[i].Name, k.Topics[j].Name) < 0
	})

	for _, g := range info.Store.Groups() {
		k.Groups = append(k.Groups, newGroup(g))
	}
	sort.Slice(k.Groups, func(i, j int) bool {
		return strings.Compare(k.Groups[i].Name, k.Groups[j].Name) < 0
	})

	k.Configs = getConfigs(info.Configs())

	return k
}

func newTopic(s *store.Store, t *store.Topic, config *asyncapi3.Channel, defaultContentType string) topic {
	var partitions []partition
	for _, p := range t.Partitions {
		partitions = append(partitions, newPartition(s, p))
	}
	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].Id < partitions[j].Id
	})

	result := topic{
		Name:        t.Name,
		Description: config.Description,
		Partitions:  partitions,
	}

	for messageId, ref := range config.Messages {
		if ref.Value == nil {
			continue
		}
		msg := ref.Value

		m := messageConfig{
			Name:        msg.Name,
			Title:       msg.Title,
			Summary:     msg.Summary,
			Description: msg.Description,
			ContentType: msg.ContentType,
		}

		if msg.Payload != nil && msg.Payload.Value != nil {
			m.Payload = &schemaInfo{Schema: msg.Payload.Value.Schema, Format: msg.Payload.Value.Format}
		}
		if msg.Headers != nil && msg.Headers.Value != nil {
			m.Header = &schemaInfo{Schema: msg.Headers.Value.Schema, Format: msg.Headers.Value.Format}
		}

		if m.ContentType == "" {
			m.ContentType = defaultContentType
		}

		if msg.Bindings.Kafka.Key != nil {
			m.Key = &schemaInfo{Schema: msg.Bindings.Kafka.Key.Value.Schema}
		}
		if result.Messages == nil {
			result.Messages = map[string]messageConfig{}
		}
		result.Messages[messageId] = m
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
			grp.Members = append(grp.Members, member{
				Name:                  id,
				Addr:                  m.Client.Addr,
				ClientSoftwareName:    m.Client.ClientSoftwareName,
				ClientSoftwareVersion: m.Client.ClientSoftwareVersion,
				Heartbeat:             m.Client.Heartbeat,
				Partitions:            m.Partitions,
			})
		}
		sort.Slice(grp.Members, func(i, j int) bool {
			return strings.Compare(grp.Members[i].Name, grp.Members[j].Name) < 0
		})
	}
	for topicName := range g.Commits {
		grp.Topics = append(grp.Topics, topicName)
	}
	sort.Slice(grp.Topics, func(i, j int) bool {
		return strings.Compare(grp.Topics[i], grp.Topics[j]) < 0
	})

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

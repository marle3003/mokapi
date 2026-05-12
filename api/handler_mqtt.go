package api

import (
	"mokapi/providers/asyncapi3"
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"net/http"
	"slices"
	"sort"
	"strings"

	"github.com/gorilla/mux"
)

type mqttInfo struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Version     string       `json:"version"`
	Contact     *contact     `json:"contact,omitempty"`
	Servers     []mqttServer `json:"servers,omitempty"`
	Topics      []mqttTopic  `json:"topics,omitempty"`
	Configs     []config     `json:"configs,omitempty"`
	Clients     []mqttClient `json:"clients,omitempty"`
}

type mqttServer struct {
	Name        string         `json:"name"`
	Host        string         `json:"host"`
	Protocol    string         `json:"protocol"`
	Title       string         `json:"title"`
	Summary     string         `json:"summary"`
	Description string         `json:"description"`
	Configs     map[string]any `json:"configs,omitempty"`
	Tags        []kafkaTag     `json:"tags,omitempty"`
}

type mqttTopic struct {
	Name        string                   `json:"name"`
	Title       string                   `json:"title,omitempty"`
	Summary     string                   `json:"summary,omitempty"`
	Description string                   `json:"description,omitempty"`
	Messages    map[string]messageConfig `json:"messages,omitempty"`
	Tags        []kafkaTag               `json:"tags,omitempty"`
	Instances   []mqttTopicInstance      `json:"instances,omitempty"`
	Metrics     mqttTopicMetrics         `json:"metrics,omitempty"`
}

type mqttTopicInstance struct {
	Name       string            `json:"name"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

type mqttClient struct {
	ClientId        string `json:"clientId"`
	Address         string `json:"address"`
	BrokerAddress   string `json:"brokerAddress"`
	ProtocolVersion byte   `json:"protocolVersion"`
}

type mqttTopicMetrics struct {
	NumMessages     float64 `json:"mqtt_messages_total"`
	LastMessageTime float64 `json:"mqtt_message_timestamp"`
}

func getMqttServices(store *runtime.MqttStore, m *monitor.Monitor) []service {
	list := store.List()
	result := make([]service, 0, len(list))
	for _, mi := range list {
		s := service{
			Name:        mi.Info.Name,
			Description: mi.Info.Description,
			Version:     mi.Info.Version,
			Type:        ServiceMqtt,
		}

		if mi.Info.Contact != nil {
			c := mi.Info.Contact
			s.Contact = &contact{
				Name:  c.Name,
				Url:   c.Url,
				Email: c.Email,
			}
		}

		s.Metrics = mqttTopicMetrics{
			NumMessages:     m.Mqtt.Messages.Sum(metrics.NewQuery(metrics.ByLabel("service", mi.Info.Name))),
			LastMessageTime: m.Mqtt.LastMessage.Max(metrics.NewQuery(metrics.ByLabel("service", mi.Info.Name))),
		}

		result = append(result, s)
	}
	return result
}

func (h *handler) setupMqtt() {
	r := h.router.PathPrefix("/api/services/mqtt").Subrouter()

	r.HandleFunc("", h.getMqttClusters).Methods(http.MethodGet)
	r.HandleFunc("/{cluster}", h.getMqttInfo).Methods(http.MethodGet)
	r.HandleFunc("/{cluster}/topics", h.getMqttTopics).Methods(http.MethodGet)
}

func (h *handler) getMqttClusters(w http.ResponseWriter, _ *http.Request) {
	services := getMqttServices(h.app.Mqtt, h.app.Monitor)
	write(w, services)
}

func (h *handler) getMqttInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	mi := h.app.Mqtt.Get(vars["cluster"])
	if mi == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	m := mqttInfo{
		Name:        mi.Config.Info.Name,
		Description: mi.Config.Info.Description,
		Version:     mi.Config.Info.Version,
	}

	if mi.Config.Info.Contact != nil {
		m.Contact = &contact{
			Name:  mi.Config.Info.Contact.Name,
			Url:   mi.Config.Info.Contact.Url,
			Email: mi.Config.Info.Contact.Email,
		}
	}

	for it := mi.Servers.Iter(); it.Next(); {
		name := it.Key()
		s := it.Value()
		if s == nil || s.Value == nil || strings.ToLower(s.Value.Protocol) != "mqtt" {
			continue
		}

		ms := mqttServer{
			Name:        name,
			Host:        s.Value.Host,
			Title:       s.Value.Title,
			Summary:     s.Value.Summary,
			Description: s.Value.Description,
			Protocol:    s.Value.Protocol,
		}

		for _, r := range s.Value.Tags {
			if r.Value == nil {
				continue
			}
			t := r.Value
			ms.Tags = append(ms.Tags, kafkaTag{
				Name:        t.Name,
				Description: t.Description,
			})
		}
		m.Servers = append(m.Servers, ms)
	}
	sort.Slice(m.Servers, func(i, j int) bool {
		return strings.Compare(m.Servers[i].Name, m.Servers[j].Name) < 0
	})

	m.Topics = getMqttTopics(mi, h.app.Monitor.Mqtt)

	for _, client := range mi.Store.Clients() {
		c := mqttClient{
			ClientId:        client.Id,
			Address:         client.Addr(),
			BrokerAddress:   client.ServerAddress(),
			ProtocolVersion: client.ProtocolVersion(),
		}
		m.Clients = append(m.Clients, c)
	}

	m.Configs = getConfigs(mi.Configs())

	write(w, m)
}

func (h *handler) getMqttTopics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	ki := h.app.Mqtt.Get(vars["cluster"])
	if ki == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	write(w, getMqttTopics(ki, h.app.Monitor.Mqtt))
}

func getMqttTopics(mi *runtime.MqttInfo, m *monitor.Mqtt) []mqttTopic {
	topics := make([]mqttTopic, 0, len(mi.Config.Channels))
	for name, ch := range mi.Config.Channels {
		if ch.Value == nil {
			continue
		}
		if !ch.Value.IsChannelAvailable("kafka") {
			continue
		}
		addr := name
		if ch.Value.Address != "" {
			addr = ch.Value.Address
		}

		data := newMqttTopic(addr, ch.Value, mi.Config)
		if len(ch.Value.Parameters) > 0 {
			for _, t := range mi.Topics {
				if err := ch.Value.IsNameValid(t.Name); err == nil {
					params, _ := ch.Value.ExtractParams(t.Name)
					data.Instances = append(data.Instances, mqttTopicInstance{
						Name:       t.Name,
						Parameters: params,
					})
				}
			}
		}

		data.Metrics = mqttTopicMetrics{
			NumMessages:     m.Messages.Sum(metrics.NewQuery(metrics.ByLabel("service", mi.Info.Name))),
			LastMessageTime: m.LastMessage.Max(metrics.NewQuery(metrics.ByLabel("service", mi.Info.Name))),
		}

		topics = append(topics, data)
	}
	slices.SortFunc(topics, func(a, b mqttTopic) int {
		return strings.Compare(a.Name, b.Name)
	})
	return topics
}

func newMqttTopic(name string, ch *asyncapi3.Channel, cfg *asyncapi3.Config) mqttTopic {
	result := mqttTopic{
		Name:        name,
		Title:       ch.Title,
		Summary:     ch.Summary,
		Description: ch.Description,
		Messages:    getMessageConfigs(ch, cfg),
	}

	for _, tRef := range ch.Tags {
		if tRef.Value == nil {
			continue
		}
		result.Tags = append(result.Tags, kafkaTag{
			Name:        tRef.Value.Name,
			Description: tRef.Value.Description,
		})
	}

	return result
}

package api

import (
	"mokapi/providers/asyncapi3"
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"net/http"
	"sort"
	"strings"
)

type mqttInfo struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Version     string           `json:"version"`
	Contact     *contact         `json:"contact,omitempty"`
	Servers     []mqttServer     `json:"servers,omitempty"`
	Topics      []mqttTopic      `json:"topics,omitempty"`
	Groups      []group          `json:"groups,omitempty"`
	Metrics     []metrics.Metric `json:"metrics,omitempty"`
	Configs     []config         `json:"configs,omitempty"`
	Clients     []mqttClient     `json:"clients,omitempty"`
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
	Description string                   `json:"description"`
	Messages    map[string]messageConfig `json:"messages"`
	Tags        []kafkaTag               `json:"tags,omitempty"`
	Instances   []mqttTopicInstance      `json:"instances,omitempty"`
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

func getMqttServices(store *runtime.MqttStore, m *monitor.Monitor) []service {
	list := store.List()
	result := make([]service, 0, len(list))
	for _, hs := range list {
		s := service{
			Name:        hs.Info.Name,
			Description: hs.Info.Description,
			Version:     hs.Info.Version,
			Type:        ServiceMqtt,
			Metrics:     m.FindAll(metrics.ByNamespace("mqtt"), metrics.ByLabel("service", hs.Info.Name)),
		}

		if hs.Info.Contact != nil {
			c := hs.Info.Contact
			s.Contact = &contact{
				Name:  c.Name,
				Url:   c.Url,
				Email: c.Email,
			}
		}

		result = append(result, s)
	}
	return result
}

func (h *handler) handleMqtt(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	switch {
	// /api/services/mqtt
	case len(segments) == 3:
		w.Header().Set("Content-Type", "application/json")
		writeJsonBody(w, getMqttClusters(h.app))
		return
	// /api/services/kafka/{cluster}
	case len(segments) == 4:
		name := segments[3]
		if s := h.app.Mqtt.Get(name); s != nil {
			m := getMqtt(s)
			m.Metrics = h.app.Monitor.FindAll(metrics.ByNamespace("mqtt"), metrics.ByLabel("service", name))

			w.Header().Set("Content-Type", "application/json")
			writeJsonBody(w, m)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
		return
	// /api/services/mqtt/{cluster}/topics
	case len(segments) == 5 && segments[4] == "topics":
		m := h.app.Mqtt.Get(segments[3])
		if m == nil {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.Header().Set("Content-Type", "application/json")
			writeJsonBody(w, getMqttTopics(m))
		}
		return
	}
}

func getMqttClusters(app *runtime.App) []cluster {
	var clusters []cluster
	for _, k := range app.Mqtt.List() {
		var c *contact
		if k.Info.Contact != nil {
			c = &contact{
				Name:  k.Info.Contact.Name,
				Url:   k.Info.Contact.Url,
				Email: k.Info.Contact.Email,
			}
		}
		clusters = append(clusters, cluster{
			Name:        k.Info.Name,
			Description: k.Info.Description,
			Contact:     c,
			Version:     k.Info.Version,
		})
	}
	return clusters
}

func getMqtt(info *runtime.MqttInfo) mqttInfo {
	m := mqttInfo{
		Name:        info.Config.Info.Name,
		Description: info.Config.Info.Description,
		Version:     info.Config.Info.Version,
	}

	if info.Config.Info.Contact != nil {
		m.Contact = &contact{
			Name:  info.Config.Info.Contact.Name,
			Url:   info.Config.Info.Contact.Url,
			Email: info.Config.Info.Contact.Email,
		}
	}

	for it := info.Servers.Iter(); it.Next(); {
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

	m.Topics = getMqttTopics(info)

	for _, client := range info.Store.Clients() {
		c := mqttClient{
			ClientId:        client.Id,
			Address:         client.Addr(),
			BrokerAddress:   client.ServerAddress(),
			ProtocolVersion: client.ProtocolVersion(),
		}
		m.Clients = append(m.Clients, c)
	}

	m.Configs = getConfigs(info.Configs())

	return m
}

func getMqttTopics(info *runtime.MqttInfo) []mqttTopic {
	topics := make([]mqttTopic, 0, len(info.Config.Channels))
	for name, ch := range info.Config.Channels {
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

		data := newMqttTopic(addr, ch.Value, info.Config)
		if len(ch.Value.Parameters) > 0 {
			for _, t := range info.Topics {
				if err := ch.Value.IsNameValid(t.Name); err == nil {
					params, _ := ch.Value.ExtractParams(t.Name)
					data.Instances = append(data.Instances, mqttTopicInstance{
						Name:       t.Name,
						Parameters: params,
					})
				}
			}
		}

		topics = append(topics, data)
	}
	sort.Slice(topics, func(i, j int) bool {
		return strings.Compare(topics[i].Name, topics[j].Name) < 0
	})
	return topics
}

func newMqttTopic(name string, ch *asyncapi3.Channel, cfg *asyncapi3.Config) mqttTopic {
	result := mqttTopic{
		Name:        name,
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

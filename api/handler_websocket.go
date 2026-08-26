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

type websocketInfo struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Version     string             `json:"version"`
	Contact     *contact           `json:"contact,omitempty"`
	Servers     []websocketServer  `json:"servers,omitempty"`
	Channels    []websocketChannel `json:"channels,omitempty"`
	Configs     []config           `json:"configs,omitempty"`
	Clients     []websocketClient  `json:"clients,omitempty"`
}

type websocketServer struct {
	Name        string     `json:"name"`
	Host        string     `json:"host"`
	Protocol    string     `json:"protocol"`
	Title       string     `json:"title"`
	Summary     string     `json:"summary"`
	Description string     `json:"description"`
	Tags        []kafkaTag `json:"tags,omitempty"`
}

type websocketChannel struct {
	Name        string                     `json:"name"`
	Title       string                     `json:"title,omitempty"`
	Summary     string                     `json:"summary,omitempty"`
	Description string                     `json:"description,omitempty"`
	Messages    map[string]messageConfig   `json:"messages,omitempty"`
	Tags        []kafkaTag                 `json:"tags,omitempty"`
	Instances   []websocketChannelInstance `json:"instances,omitempty"`
	Metrics     websocketChannelMetrics    `json:"metrics,omitempty"`
}

type websocketChannelInstance struct {
	Name       string            `json:"name"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

type websocketClient struct {
	Id            string `json:"id"`
	Address       string `json:"address"`
	ServerAddress string `json:"serverAddress"`
}

type websocketChannelMetrics struct {
	NumMessages     float64 `json:"websocket_messages_total"`
	LastMessageTime float64 `json:"websocket_message_timestamp"`
}

func getWebsocketServices(store *runtime.WebsocketStore, m *monitor.Monitor) []service {
	list := store.List()
	result := make([]service, 0, len(list))
	for _, mi := range list {
		s := service{
			Name:        mi.Info.Name,
			Description: mi.Info.Description,
			Version:     mi.Info.Version,
			Type:        ServiceWebsocket,
		}

		if mi.Info.Contact != nil {
			c := mi.Info.Contact
			s.Contact = &contact{
				Name:  c.Name,
				Url:   c.Url,
				Email: c.Email,
			}
		}

		s.Metrics = websocketChannelMetrics{
			NumMessages:     m.Websocket.Messages.Sum(metrics.NewQuery(metrics.ByLabel("service", mi.Info.Name))),
			LastMessageTime: m.Websocket.LastMessage.Max(metrics.NewQuery(metrics.ByLabel("service", mi.Info.Name))),
		}

		result = append(result, s)
	}
	return result
}

func (h *handler) setupWebsocket() {
	r := h.router.PathPrefix("/api/services/websocket").Subrouter()

	r.HandleFunc("", h.getWebsocketServices).Methods(http.MethodGet)
	r.HandleFunc("/{service}", h.getWebsocketInfo).Methods(http.MethodGet)
	r.HandleFunc("/{cluster}/asyncapi.{ext}", h.exportAsyncApi).Methods(http.MethodGet)
	r.HandleFunc("/{service}/channels", h.getWebsocketChannels).Methods(http.MethodGet)
}

func (h *handler) getWebsocketServices(w http.ResponseWriter, _ *http.Request) {
	services := getWebsocketServices(h.app.Websocket, h.app.Monitor)
	write(w, services)
}

func (h *handler) getWebsocketInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	wi := h.app.Websocket.Get(vars["service"])
	if wi == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	m := websocketInfo{
		Name:        wi.Config.Info.Name,
		Description: wi.Config.Info.Description,
		Version:     wi.Config.Info.Version,
	}

	if wi.Config.Info.Contact != nil {
		m.Contact = &contact{
			Name:  wi.Config.Info.Contact.Name,
			Url:   wi.Config.Info.Contact.Url,
			Email: wi.Config.Info.Contact.Email,
		}
	}

	for it := wi.Servers.Iter(); it.Next(); {
		name := it.Key()
		s := it.Value()
		if s == nil || s.Value == nil || strings.ToLower(s.Value.Protocol) != "ws" {
			continue
		}

		ws := websocketServer{
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
			ws.Tags = append(ws.Tags, kafkaTag{
				Name:        t.Name,
				Description: t.Description,
			})
		}
		m.Servers = append(m.Servers, ws)
	}
	sort.Slice(m.Servers, func(i, j int) bool {
		return strings.Compare(m.Servers[i].Name, m.Servers[j].Name) < 0
	})

	m.Channels = getWebsocketChannels(wi, h.app.Monitor.Websocket)

	for _, client := range wi.Store.Clients() {
		c := websocketClient{
			Id:            client.Id,
			Address:       client.RemoteAddr,
			ServerAddress: client.ServerAddr,
		}
		m.Clients = append(m.Clients, c)
	}

	m.Configs = getConfigs(wi.Configs())

	write(w, m)
}

func (h *handler) getWebsocketChannels(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	wi := h.app.Websocket.Get(vars["service"])
	if wi == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	write(w, getWebsocketChannels(wi, h.app.Monitor.Websocket))
}

func getWebsocketChannels(wi *runtime.WebsocketInfo, m *monitor.Websocket) []websocketChannel {
	channels := make([]websocketChannel, 0, len(wi.Config.Channels))
	for name, ch := range wi.Config.Channels {
		if ch.Value == nil {
			continue
		}
		if !ch.Value.IsChannelAvailable("ws") {
			continue
		}
		addr := name
		if ch.Value.Address != "" {
			addr = ch.Value.Address
		}

		data := newWebsocketChannel(addr, ch.Value, wi.Config)
		if len(ch.Value.Parameters) > 0 {
			for _, c := range wi.Store.Channels {
				if err := ch.Value.IsNameValid(c.Name); err == nil {
					params, _ := ch.Value.ExtractParams(c.Name)
					data.Instances = append(data.Instances, websocketChannelInstance{
						Name:       c.Name,
						Parameters: params,
					})
				}
			}
		}

		data.Metrics = websocketChannelMetrics{
			NumMessages:     m.Messages.Sum(metrics.NewQuery(metrics.ByLabel("service", wi.Info.Name))),
			LastMessageTime: m.LastMessage.Max(metrics.NewQuery(metrics.ByLabel("service", wi.Info.Name))),
		}

		channels = append(channels, data)
	}
	slices.SortFunc(channels, func(a, b websocketChannel) int {
		return strings.Compare(a.Name, b.Name)
	})
	return channels
}

func newWebsocketChannel(name string, ch *asyncapi3.Channel, cfg *asyncapi3.Config) websocketChannel {
	result := websocketChannel{
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

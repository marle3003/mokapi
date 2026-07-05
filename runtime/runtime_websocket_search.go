package runtime

import (
	"fmt"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/search"
	"mokapi/schema/json/schema"
	"strings"

	log "github.com/sirupsen/logrus"
)

type websocketSearchIndexData struct {
	Type          string                      `json:"type"`
	Discriminator string                      `json:"discriminator"`
	Api           string                      `json:"api"`
	Name          string                      `json:"name"`
	Version       string                      `json:"version"`
	Description   string                      `json:"description"`
	Contact       *asyncapi3.Contact          `json:"contact"`
	Servers       []websocketServerSearchData `json:"servers"`
	Meta          map[string]string           `json:"meta"`
}

type websocketServerSearchData struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

type websocketTopicSearchIndexData struct {
	Type          string                            `json:"type"`
	Discriminator string                            `json:"discriminator"`
	Api           string                            `json:"api"`
	ChannelId     string                            `json:"channelId"`
	Name          string                            `json:"name"`
	Title         string                            `json:"title"`
	Address       string                            `json:"address"`
	Summary       string                            `json:"summary"`
	Description   string                            `json:"description"`
	Messages      []websocketMessageSearchIndexData `json:"messages"`
}

type websocketMessageSearchIndexData struct {
	MessageId   string            `json:"messageId"`
	Name        string            `json:"name"`
	Title       string            `json:"title"`
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	Payload     *schema.IndexData `json:"payload"`
}

func (s *WebsocketStore) addToIndex(cfg *asyncapi3.Config) {
	if cfg == nil || cfg.Info.Name == "" {
		return
	}

	c := websocketSearchIndexData{
		Type:          "websocket",
		Discriminator: "websocket",
		Api:           cfg.Info.Name,
		Name:          cfg.Info.Name,
		Version:       cfg.Info.Version,
		Description:   cfg.Info.Description,
		Contact:       cfg.Info.Contact,
		Meta: map[string]string{
			"channels": fmt.Sprintf("%d", len(cfg.Channels)),
		},
	}
	for it := cfg.Servers.Iter(); it.Next(); {
		name := it.Key()
		server := it.Value()
		if server == nil || server.Value == nil {
			continue
		}
		c.Servers = append(c.Servers, websocketServerSearchData{
			Name:        name,
			Host:        server.Value.Host,
			Title:       server.Value.Title,
			Summary:     server.Value.Summary,
			Description: server.Value.Description,
		})
	}

	s.index.Add(fmt.Sprintf("websocket_%s", cfg.Info.Name), c)

	for name, ch := range cfg.Channels {
		if ch == nil || ch.Value == nil {
			continue
		}
		chName := name
		if ch.Value.Name != "" {
			chName = ch.Value.Name
		}
		if ch.Value.Address != "" {
			chName = ch.Value.Address
		}

		t := websocketTopicSearchIndexData{
			Type:          "websocket",
			Discriminator: "websocket_channel",
			Api:           cfg.Info.Name,
			ChannelId:     name,
			Name:          chName,
			Title:         ch.Value.Title,
			Address:       ch.Value.Address,
			Summary:       ch.Value.Summary,
			Description:   ch.Value.Description,
		}

		for messageId, message := range ch.Value.Messages {
			if message == nil || message.Value == nil {
				continue
			}
			p, err := getSchema(message.Value.Headers)
			if err != nil {
				log.Errorf("indexing message for topic '%v' failed for payload: %v", ch.Value.Name, err)
			}

			t.Messages = append(t.Messages, websocketMessageSearchIndexData{
				MessageId:   messageId,
				Name:        message.Value.Name,
				Title:       message.Value.Title,
				Summary:     message.Value.Summary,
				Description: message.Value.Description,
				Payload:     p,
			})
		}
		id := fmt.Sprintf("websocket_%s_%s", cfg.Info.Name, name)
		s.index.Add(id, t)
	}
}

func getWebsocketSearchResult(fields map[string]string, discriminator []string) (search.ResultItem, error) {
	result := search.ResultItem{
		Type: "Websocket",
	}

	if len(discriminator) == 1 {
		result.Title = fields["name"]
		result.Params = map[string]string{
			"type":    strings.ToLower(result.Type),
			"service": result.Title,
		}
		for k, v := range fields {
			if strings.HasPrefix(k, "meta.") {
				k = strings.Replace(k, "meta.", "", 1)
				result.Params[k] = v
			}
		}
		return result, nil
	}

	switch discriminator[1] {
	case "channel":
		title := fields["channelId"]
		if len(fields["name"]) > 0 {
			title = fields["name"]
		} else if len(fields["title"]) > 0 {
			title = fields["title"]
		}
		result.Description = BuildDescription(150, fields["summary"], fields["description"])
		result.Domain = fields["api"]
		result.Title = fmt.Sprintf("Channel %s", title)
		result.Params = map[string]string{
			"type":    strings.ToLower(result.Type),
			"service": result.Domain,
			"channel": fields["name"],
		}
	default:
		return result, fmt.Errorf("unsupported search result: %s", strings.Join(discriminator, "_"))
	}
	return result, nil
}

func (s *WebsocketStore) removeFromIndex(cfg *asyncapi3.Config) {
	s.index.Delete(fmt.Sprintf("websocket_%s", cfg.Info.Name))

	for name := range cfg.Channels {
		s.index.Delete(fmt.Sprintf("websocket_%s_%s", cfg.Info.Name, name))
	}
}

package runtime

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/search"
	"mokapi/schema/json/schema"
	"strings"
)

type kafkaSearchIndexData struct {
	Type          string                  `json:"type"`
	Discriminator string                  `json:"discriminator"`
	Api           string                  `json:"api"`
	Name          string                  `json:"name"`
	Version       string                  `json:"version"`
	Description   string                  `json:"description"`
	Contact       *asyncapi3.Contact      `json:"contact"`
	Servers       []kafkaServerSearchData `json:"servers"`
}

type kafkaServerSearchData struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

type kafkaTopicSearchIndexData struct {
	Type          string                        `json:"type"`
	Discriminator string                        `json:"discriminator"`
	Api           string                        `json:"api"`
	ChannelId     string                        `json:"channelId"`
	Name          string                        `json:"name"`
	Title         string                        `json:"title"`
	Address       string                        `json:"address"`
	Summary       string                        `json:"summary"`
	Description   string                        `json:"description"`
	Messages      []kafkaMessageSearchIndexData `json:"messages"`
}

type kafkaMessageSearchIndexData struct {
	MessageId   string            `json:"messageId"`
	Name        string            `json:"name"`
	Title       string            `json:"title"`
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	Headers     *schema.IndexData `json:"headers"`
	Payload     *schema.IndexData `json:"payload"`
}

func (s *KafkaStore) addToIndex(cfg *asyncapi3.Config) {
	if cfg == nil || cfg.Info.Name == "" {
		return
	}

	c := kafkaSearchIndexData{
		Type:          "kafka",
		Discriminator: "kafka",
		Api:           cfg.Info.Name,
		Name:          cfg.Info.Name,
		Version:       cfg.Info.Version,
		Description:   cfg.Info.Description,
		Contact:       cfg.Info.Contact,
	}
	for name, server := range cfg.Servers {
		if server == nil || server.Value == nil {
			continue
		}
		c.Servers = append(c.Servers, kafkaServerSearchData{
			Name:        name,
			Host:        server.Value.Host,
			Title:       server.Value.Title,
			Summary:     server.Value.Summary,
			Description: server.Value.Description,
		})
	}

	add(s.index, fmt.Sprintf("kafka_%s", cfg.Info.Name), c)

	for name, topic := range cfg.Channels {
		if topic == nil || topic.Value == nil {
			continue
		}

		t := kafkaTopicSearchIndexData{
			Type:          "kafka",
			Discriminator: "kafka_topic",
			Api:           cfg.Info.Name,
			ChannelId:     name,
			Name:          topic.Value.Name,
			Title:         topic.Value.Title,
			Address:       topic.Value.Address,
			Summary:       topic.Value.Summary,
			Description:   topic.Value.Description,
		}

		for messageId, message := range topic.Value.Messages {
			if message == nil || message.Value == nil {
				continue
			}
			h, err := getSchema(message.Value.Headers)
			if err != nil {
				log.Errorf("indexing message for topic '%v' failed for headers: %v", topic.Value.Name, err)
			}
			p, err := getSchema(message.Value.Headers)
			if err != nil {
				log.Errorf("indexing message for topic '%v' failed for payload: %v", topic.Value.Name, err)
			}

			t.Messages = append(t.Messages, kafkaMessageSearchIndexData{
				MessageId:   messageId,
				Name:        message.Value.Name,
				Title:       message.Value.Title,
				Summary:     message.Value.Summary,
				Description: message.Value.Description,
				Headers:     h,
				Payload:     p,
			})
		}
		id := fmt.Sprintf("kafka_%s_%s", cfg.Info.Name, name)
		add(s.index, id, t)
	}
}

func getKafkaSearchResult(fields map[string]string, discriminator []string) (search.ResultItem, error) {
	result := search.ResultItem{
		Type: "Kafka",
	}

	if len(discriminator) == 1 {
		result.Title = fields["name"]
		result.Params = map[string]string{
			"type":    strings.ToLower(result.Type),
			"service": result.Title,
		}
		return result, nil
	}

	switch discriminator[1] {
	case "topic":
		title := fields["channelId"]
		if len(fields["name"]) > 0 {
			title = fields["name"]
		} else if len(fields["title"]) > 0 {
			title = fields["title"]
		}
		result.Domain = fields["api"]
		result.Title = fmt.Sprintf("Topic %s", title)
		result.Params = map[string]string{
			"type":    strings.ToLower(result.Type),
			"service": result.Domain,
			"topic":   fields["name"],
		}
	default:
		return result, fmt.Errorf("unsupported search result: %s", strings.Join(discriminator, "_"))
	}
	return result, nil
}

func getSchema(s *asyncapi3.SchemaRef) (*schema.IndexData, error) {
	if s == nil || s.Value == nil {
		return nil, nil
	}
	switch v := s.Value.Schema.(type) {
	case *schema.Schema:
		return schema.NewIndexData(v), nil
	default:
		return nil, fmt.Errorf("unsupported schema type: %T", v)
	}
}

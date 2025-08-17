package api

import (
	"encoding/json"
	"fmt"
	"io"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type kafkaSummary struct {
	service
}

type cluster struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Contact     *contact `json:"contact,omitempty"`
	Version     string   `json:"version,omitempty"`
}

type kafkaInfo struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Version     string           `json:"version"`
	Contact     *contact         `json:"contact,omitempty"`
	Servers     []kafkaServer    `json:"servers,omitempty"`
	Topics      []topic          `json:"topics,omitempty"`
	Groups      []group          `json:"groups,omitempty"`
	Metrics     []metrics.Metric `json:"metrics,omitempty"`
	Configs     []config         `json:"configs,omitempty"`
}

type kafkaServer struct {
	Name        string           `json:"name"`
	Host        string           `json:"host"`
	Protocol    string           `json:"protocol"`
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

type topic struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Partitions  []partition              `json:"partitions"`
	Messages    map[string]messageConfig `json:"messages,omitempty"`
	Bindings    bindings                 `json:"bindings,omitempty"`
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

type bindings struct {
	Partitions            int   `json:"partitions,omitempty"`
	RetentionBytes        int64 `json:"retentionBytes,omitempty"`
	RetentionMs           int64 `json:"retentionMs,omitempty"`
	SegmentBytes          int64 `json:"segmentBytes,omitempty"`
	SegmentMs             int64 `json:"segmentMs,omitempty"`
	ValueSchemaValidation bool  `json:"valueSchemaValidation,omitempty"`
	KeySchemaValidation   bool  `json:"keySchemaValidation,omitempty"`
}

type produceRequest struct {
	Records []store.Record `json:"records"`
}

type produceResponse struct {
	Offsets []recordResult `json:"offsets"`
}

type recordResult struct {
	Partition int
	Offset    int64
	Error     string
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

func (h *handler) handleKafka(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	switch {
	// /api/services/kafka
	case len(segments) == 3:
		w.Header().Set("Content-Type", "application/json")
		writeJsonBody(w, getKafkaClusters(h.app))
		return
	// /api/services/kafka/{cluster}
	case len(segments) == 4:
		name := segments[3]
		if s := h.app.Kafka.Get(name); s != nil {
			k := getKafka(s)
			k.Metrics = h.app.Monitor.FindAll(metrics.ByNamespace("kafka"), metrics.ByLabel("service", name))

			w.Header().Set("Content-Type", "application/json")
			writeJsonBody(w, k)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
		return
	// /api/services/kafka/{cluster}/topics
	case len(segments) == 5 && segments[4] == "topics":
		k := h.app.Kafka.Get(segments[3])
		if k == nil {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.Header().Set("Content-Type", "application/json")
			writeJsonBody(w, getTopics(k))
		}
		return
	// /api/services/kafka/{cluster}/topics/{topic}
	case len(segments) == 6 && segments[4] == "topics":
		k := h.app.Kafka.Get(segments[3])
		if k == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		topicName := segments[5]

		if r.Method == "GET" {
			t := getTopic(k, topicName)
			if t == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			} else {
				w.Header().Set("Content-Type", "application/json")
				writeJsonBody(w, t)
				return
			}
		} else if r.Method == "POST" {
			records, err := getProduceRecords(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			c := store.NewClient(k.Store, h.app.Monitor.Kafka)
			ct := media.ParseContentType(r.Header.Get("Content-Type"))
			result, err := c.Write(topicName, records, &ct)
			if err != nil {
				if errors.Is(err, store.TopicNotFound) || errors.Is(err, store.PartitionNotFound) {
					http.Error(w, err.Error(), http.StatusNotFound)
				} else {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
			}
			res := produceResponse{}
			for _, rec := range result {
				res.Offsets = append(res.Offsets, recordResult{
					Partition: rec.Partition,
					Offset:    rec.Offset,
					Error:     rec.Error,
				})
			}
			w.Header().Set("Content-Type", "application/json")
			writeJsonBody(w, res)
			return
		}
	// /api/services/kafka/{cluster}/topics/{topic}/partitions
	case len(segments) == 7 && segments[6] == "partitions":
		k := h.app.Kafka.Get(segments[3])
		if k == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		topicName := segments[5]
		t := k.Store.Topic(topicName)
		if t == nil {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.Header().Set("Content-Type", "application/json")
			writeJsonBody(w, getPartitions(k, t))
		}
		return
	// /api/services/kafka/{cluster}/topics/{topic}/partitions/{id}
	case len(segments) == 8 && segments[6] == "partitions":
		k := h.app.Kafka.Get(segments[3])
		if k == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		topicName := segments[5]
		t := k.Store.Topic(topicName)
		if t == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		idValue := segments[7]
		id, err := strconv.Atoi(idValue)
		if err != nil {
			http.Error(w, "error partition ID is not an integer", http.StatusBadRequest)
			return
		}
		p := t.Partition(id)
		if p == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			writeJsonBody(w, newPartition(k.Store, p))
		} else {
			records, err := getProduceRecords(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			for _, record := range records {
				record.Partition = id
			}
			c := store.NewClient(k.Store, h.app.Monitor.Kafka)
			ct := media.ParseContentType(r.Header.Get("Content-Type"))
			result, err := c.Write(topicName, records, &ct)
			if err != nil {
				if errors.Is(err, store.TopicNotFound) || errors.Is(err, store.PartitionNotFound) {
					http.Error(w, err.Error(), http.StatusNotFound)
				} else {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
			}
			res := produceResponse{}
			for _, rec := range result {
				res.Offsets = append(res.Offsets, recordResult{
					Partition: rec.Partition,
					Offset:    rec.Offset,
					Error:     rec.Error,
				})
			}
			w.Header().Set("Content-Type", "application/json")
			writeJsonBody(w, res)
			return
		}
		return
	// /api/services/kafka/{cluster}/topics/{topic}/partitions/{id}/offsets
	case len(segments) == 9 && segments[8] == "offsets":
		k := h.app.Kafka.Get(segments[3])
		if k == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		topicName := segments[5]
		idValue := segments[7]
		id, err := strconv.Atoi(idValue)
		if err != nil {
			http.Error(w, fmt.Errorf("error partition ID is not an integer").Error(), http.StatusBadRequest)
			return
		}
		offsetValue := r.URL.Query().Get("offset")
		offset := -1
		if offsetValue != "" {
			offset, err = strconv.Atoi(offsetValue)
			if err != nil {
				http.Error(w, fmt.Errorf("error offset is not an integer").Error(), http.StatusBadRequest)
				return
			}
		}

		c := store.NewClient(k.Store, h.app.Monitor.Kafka)
		ct := media.ParseContentType(r.Header.Get("Accept"))
		records, err := c.Read(topicName, id, int64(offset), &ct)
		if err != nil {
			if errors.Is(err, store.TopicNotFound) || errors.Is(err, store.PartitionNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJsonBody(w, records)
		return
	// /api/services/kafka/{cluster}/topics/{topic}/partitions/{id}/offsets/0
	case len(segments) == 10 && segments[8] == "offsets":
		k := h.app.Kafka.Get(segments[3])
		if k == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		topicName := segments[5]
		idValue := segments[7]
		id, err := strconv.Atoi(idValue)
		if err != nil {
			http.Error(w, fmt.Errorf("error partition ID is not an integer").Error(), http.StatusBadRequest)
			return
		}
		offsetValue := segments[9]
		offset, err := strconv.Atoi(offsetValue)
		if err != nil {
			http.Error(w, fmt.Errorf("error offset is not an integer").Error(), http.StatusBadRequest)
			return
		}

		c := store.NewClient(k.Store, h.app.Monitor.Kafka)
		ct := media.ParseContentType(r.Header.Get("Accept"))
		records, err := c.Read(topicName, id, int64(offset), &ct)
		if err != nil {
			if errors.Is(err, store.TopicNotFound) || errors.Is(err, store.PartitionNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}
		if len(records) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJsonBody(w, records[0])
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func getKafka(info *runtime.KafkaInfo) kafkaInfo {
	k := kafkaInfo{
		Name:        info.Config.Info.Name,
		Description: info.Config.Info.Description,
		Version:     info.Config.Info.Version,
		Groups:      make([]group, 0),
	}

	if info.Config.Info.Contact != nil {
		k.Contact = &contact{
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
			Protocol:    s.Value.Protocol,
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

	k.Topics = getTopics(info)

	for _, g := range info.Store.Groups() {
		k.Groups = append(k.Groups, newGroup(g))
	}
	sort.Slice(k.Groups, func(i, j int) bool {
		return strings.Compare(k.Groups[i].Name, k.Groups[j].Name) < 0
	})

	k.Configs = getConfigs(info.Configs())

	return k
}

func getTopics(info *runtime.KafkaInfo) []topic {
	topics := make([]topic, 0, len(info.Config.Channels))
	for name, ch := range info.Config.Channels {
		if ch.Value == nil {
			continue
		}
		addr := ch.Value.Address
		if addr == "" {
			addr = name
		}
		t := info.Store.Topic(addr)
		topics = append(topics, newTopic(info.Store, t, ch.Value, info.Config))
	}
	sort.Slice(topics, func(i, j int) bool {
		return strings.Compare(topics[i].Name, topics[j].Name) < 0
	})
	return topics
}

func getTopic(info *runtime.KafkaInfo, name string) *topic {
	for n, ch := range info.Config.Channels {
		if ch.Value == nil {
			continue
		}
		addr := ch.Value.Address
		if addr == "" {
			addr = n
		}
		if addr == name {
			t := info.Store.Topic(addr)
			r := newTopic(info.Store, t, ch.Value, info.Config)
			return &r
		}

	}
	return nil
}

func newTopic(s *store.Store, t *store.Topic, ch *asyncapi3.Channel, cfg *asyncapi3.Config) topic {
	var partitions []partition
	for _, p := range t.Partitions {
		partitions = append(partitions, newPartition(s, p))
	}
	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].Id < partitions[j].Id
	})

	result := topic{
		Name:        t.Name,
		Description: ch.Description,
		Partitions:  partitions,
		Bindings: bindings{
			Partitions:            t.Config.Bindings.Kafka.Partitions,
			RetentionBytes:        t.Config.Bindings.Kafka.RetentionBytes,
			RetentionMs:           t.Config.Bindings.Kafka.RetentionMs,
			SegmentBytes:          t.Config.Bindings.Kafka.SegmentBytes,
			SegmentMs:             t.Config.Bindings.Kafka.SegmentMs,
			ValueSchemaValidation: t.Config.Bindings.Kafka.ValueSchemaValidation,
			KeySchemaValidation:   t.Config.Bindings.Kafka.KeySchemaValidation,
		},
	}

	for messageId, ref := range ch.Messages {
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
		if m.Name == "" {
			m.Name = messageId
		}

		if msg.Payload != nil && msg.Payload.Value != nil {
			m.Payload = &schemaInfo{Schema: msg.Payload.Value.Schema, Format: msg.Payload.Value.Format}
		}
		if msg.Headers != nil && msg.Headers.Value != nil {
			m.Header = &schemaInfo{Schema: msg.Headers.Value.Schema, Format: msg.Headers.Value.Format}
		}

		if m.ContentType == "" {
			m.ContentType = cfg.DefaultContentType
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

func getPartitions(info *runtime.KafkaInfo, t *store.Topic) []partition {
	var partitions []partition
	for _, p := range t.Partitions {
		partitions = append(partitions, newPartition(info.Store, p))
	}
	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].Id < partitions[j].Id
	})
	return partitions
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

func getKafkaClusters(app *runtime.App) []cluster {
	var clusters []cluster
	for _, k := range app.Kafka.List() {
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

func getProduceRecords(r *http.Request) ([]store.Record, error) {
	var pr produceRequest
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body")
	}
	err = json.Unmarshal(b, &pr)
	if err != nil {
		return nil, fmt.Errorf("error parsing body")
	}
	return pr.Records, nil
}

func (r *recordResult) MarshalJSON() ([]byte, error) {
	aux := &struct {
		Partition int     `json:"partition"`
		Offset    int64   `json:"offset"`
		Error     *string `json:"error,omitempty"`
	}{
		Partition: r.Partition,
		Offset:    r.Offset,
	}
	if r.Error != "" {
		aux.Error = &r.Error
	}
	return json.Marshal(aux)
}

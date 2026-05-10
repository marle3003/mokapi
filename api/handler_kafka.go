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
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type kafkaInfo struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Version     string            `json:"version"`
	Contact     *contact          `json:"contact,omitempty"`
	Servers     []kafkaServer     `json:"servers,omitempty"`
	Topics      []kafkaTopicInfo  `json:"topics,omitempty"`
	Groups      []kafkaGroupInfo  `json:"groups,omitempty"`
	Configs     []configInfo      `json:"configs,omitempty"`
	Clients     []kafkaClientInfo `json:"clients,omitempty"`
}

type kafkaServer struct {
	Name        string         `json:"name"`
	Host        string         `json:"host"`
	Protocol    string         `json:"protocol"`
	Title       string         `json:"title,omitempty"`
	Summary     string         `json:"summary,omitempty"`
	Description string         `json:"description,omitempty"`
	Configs     map[string]any `json:"configs,omitempty"`
	Tags        []kafkaTag     `json:"tags,omitempty"`
}

type kafkaTag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type kafkaGroupInfo struct {
	Name       string            `json:"name"`
	Generation int               `json:"generation"`
	State      string            `json:"state"`
	Protocol   string            `json:"protocol"`
	Members    int               `json:"members"`
	Metrics    kafkaGroupMetrics `json:"metrics"`
}

type group struct {
	Name       string            `json:"name"`
	Generation int               `json:"generation"`
	State      string            `json:"state"`
	Protocol   string            `json:"protocol"`
	Members    []kafkaMember     `json:"members"`
	Leader     string            `json:"leader"`
	Topics     []string          `json:"topics"`
	Metrics    kafkaGroupMetrics `json:"metrics"`
}

type kafkaClientInfo struct {
	ClientId              string `json:"clientId"`
	Address               string `json:"address"`
	ClientSoftwareName    string `json:"clientSoftwareName"`
	ClientSoftwareVersion string `json:"clientSoftwareVersion"`
}

type kafkaClient struct {
	kafkaClientInfo
	BrokerAddress string              `json:"brokerAddress"`
	Groups        []clientGroupMember `json:"groups"`
}

type clientGroupMember struct {
	MemberId string `json:"memberId"`
	Group    string `json:"group"`
}

type kafkaMember struct {
	Name                  string           `json:"name"`
	ClientId              string           `json:"clientId"`
	Addr                  string           `json:"addr"`
	ClientSoftwareName    string           `json:"clientSoftwareName"`
	ClientSoftwareVersion string           `json:"clientSoftwareVersion"`
	Heartbeat             time.Time        `json:"heartbeat"`
	Partitions            map[string][]int `json:"partitions"`
}

type kafkaTopicInfo struct {
	Name    string           `json:"name"`
	Summary string           `json:"summary,omitempty"`
	Tags    []kafkaTag       `json:"tags,omitempty"`
	Metrics kafkaTopicMetric `json:"metrics"`
}

type kafkaTopic struct {
	Name        string                   `json:"name"`
	Title       string                   `json:"title,omitempty"`
	Summary     string                   `json:"summary,omitempty"`
	Description string                   `json:"description,omitempty"`
	Partitions  []kafkaPartition         `json:"partitions"`
	Messages    map[string]messageConfig `json:"messages,omitempty"`
	Bindings    kafkaBindings            `json:"bindings,omitempty"`
	Tags        []kafkaTag               `json:"tags,omitempty"`
	Groups      []kafkaGroupInfo         `json:"groups,omitempty"`
}

type kafkaPartition struct {
	Id          int   `json:"id"`
	StartOffset int64 `json:"startOffset"`
	Offset      int64 `json:"offset"`
	Segments    int   `json:"segments"`
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

type kafkaBindings struct {
	Partitions            int   `json:"partitions,omitempty"`
	RetentionBytes        int64 `json:"retentionBytes,omitempty"`
	RetentionMs           int64 `json:"retentionMs,omitempty"`
	SegmentBytes          int64 `json:"segmentBytes,omitempty"`
	SegmentMs             int64 `json:"segmentMs,omitempty"`
	ValueSchemaValidation bool  `json:"valueSchemaValidation,omitempty"`
	KeySchemaValidation   bool  `json:"keySchemaValidation,omitempty"`
}

type kafkaProduceRequest struct {
	Records []store.Record `json:"records"`
}

type KafkaProduceResponse struct {
	Offsets []KafkaRecordResult `json:"offsets"`
}

type KafkaRecordResult struct {
	Partition int
	Offset    int64
	Error     string
}

type kafkaTopicMetric struct {
	NumMessages     float64 `json:"kafka_messages_total"`
	LastMessageTime float64 `json:"kafka_message_timestamp"`
}

type kafkaGroupMetrics struct {
	LastRebalancing float64                             `json:"kafka_rebalance_timestamp"`
	Topics          map[string][]kafkaTopicGroupMetrics `json:"topics,omitempty"`
}

type kafkaTopicGroupMetrics struct {
	Partition int     `json:"partition"`
	Lags      float64 `json:"kafka_consumer_group_lag"`
	Commit    float64 `json:"kafka_consumer_group_commit"`
}

func getKafkaServices(store *runtime.KafkaStore, m *monitor.Monitor) []service {
	list := store.List()
	result := make([]service, 0, len(list))
	for _, ki := range list {
		s := service{
			Name:        ki.Info.Name,
			Description: ki.Info.Description,
			Version:     ki.Info.Version,
			Type:        ServiceKafka,
		}

		s.Metrics = kafkaTopicMetric{
			NumMessages:     m.Kafka.Messages.Sum(metrics.NewQuery(metrics.ByLabel("service", ki.Info.Name))),
			LastMessageTime: m.Kafka.LastMessage.Max(metrics.NewQuery(metrics.ByLabel("service", ki.Info.Name))),
		}

		if ki.Info.Contact != nil {
			c := ki.Info.Contact
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

func (h *handler) setupKafka() {
	r := h.router.PathPrefix("/api/services/kafka").Subrouter()

	r.HandleFunc("", h.getKafkaClusters).Methods(http.MethodGet)
	r.HandleFunc("/{cluster}", h.getKafkaInfo).Methods(http.MethodGet)
	r.HandleFunc("/{cluster}/topics", h.getKafkaTopics).Methods(http.MethodGet)
	r.HandleFunc("/{cluster}/topics/{topic}", h.getKafkaTopic).Methods(http.MethodGet)
	r.HandleFunc("/{cluster}/topics/{topic}", h.produceKafkaMessage).Methods(http.MethodPost)
	r.HandleFunc("/{cluster}/topics/{topic}/partitions", h.getKafkaPartitions).Methods(http.MethodGet)
	r.HandleFunc("/{cluster}/topics/{topic}/partitions/{partition}", h.getKafkaPartition).Methods(http.MethodGet)
	r.HandleFunc("/{cluster}/topics/{topic}/partitions/{partition}", h.produceKafkaMessage).Methods(http.MethodPost)
	r.HandleFunc("/{cluster}/topics/{topic}/partitions/{partition}/offsets", h.getKafkaMessages).Methods(http.MethodGet)
	r.HandleFunc("/{cluster}/topics/{topic}/partitions/{partition}/offsets/{offset}", h.getKafkaMessage).Methods(http.MethodGet)
	r.HandleFunc("/{cluster}/groups/{group}", h.getKafkaGroup).Methods(http.MethodGet)
}

func (h *handler) getKafkaClusters(w http.ResponseWriter, _ *http.Request) {
	services := getKafkaServices(h.app.Kafka, h.app.Monitor)
	write(w, services)
}

func (h *handler) getKafkaInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	ki := h.app.Kafka.Get(vars["cluster"])
	if ki == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	k := kafkaInfo{
		Name:        ki.Config.Info.Name,
		Description: ki.Config.Info.Description,
		Version:     ki.Config.Info.Version,
	}

	if ki.Config.Info.Contact != nil {
		k.Contact = &contact{
			Name:  ki.Config.Info.Contact.Name,
			Url:   ki.Config.Info.Contact.Url,
			Email: ki.Config.Info.Contact.Email,
		}
	}

	for it := ki.Servers.Iter(); it.Next(); {
		name := it.Key()
		s := it.Value()
		if s == nil || s.Value == nil || strings.ToLower(s.Value.Protocol) != "kafka" {
			continue
		}

		ks := kafkaServer{
			Name:        name,
			Host:        s.Value.Host,
			Title:       s.Value.Title,
			Summary:     s.Value.Summary,
			Description: s.Value.Description,
			Protocol:    s.Value.Protocol,
			Configs:     s.Value.Bindings.Kafka.Configs(),
		}
		for _, r := range s.Value.Tags {
			if r.Value == nil {
				continue
			}
			t := r.Value
			ks.Tags = append(ks.Tags, kafkaTag{
				Name:        t.Name,
				Description: t.Description,
			})
		}
		k.Servers = append(k.Servers, ks)
	}
	sort.Slice(k.Servers, func(i, j int) bool {
		return strings.Compare(k.Servers[i].Name, k.Servers[j].Name) < 0
	})

	k.Topics = getTopicInfos(ki, h.app.Monitor.Kafka)
	k.Groups = getGroupInfos(ki, "", h.app.Monitor.Kafka)
	k.Configs = getConfigInfos(ki.Configs())

	for _, ctx := range ki.Store.Clients() {
		c := kafkaClientInfo{
			ClientId:              ctx.ClientId,
			Address:               ctx.Addr,
			ClientSoftwareName:    ctx.ClientSoftwareName,
			ClientSoftwareVersion: ctx.ClientSoftwareVersion,
		}

		k.Clients = append(k.Clients, c)
	}

	write(w, k)
}

func (h *handler) getKafkaTopics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	ki := h.app.Kafka.Get(vars["cluster"])
	if ki == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	topics := getTopicInfos(ki, h.app.Monitor.Kafka)
	write(w, topics)
}

func (h *handler) getKafkaTopic(w http.ResponseWriter, r *http.Request) {
	ki, t, ok := h.resolveKafkaTopic(w, r)
	if !ok {
		return
	}

	for n, ch := range ki.Channels {
		if ch.Value == nil {
			continue
		}
		addr := ch.Value.Address
		if addr == "" {
			addr = n
		}
		if addr == t.Name {
			r := newTopic(t, ki, ch.Value, h.app.Monitor.Kafka)
			write(w, r)
			return
		}

	}
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("kafka topic not found"))
}

func (h *handler) getKafkaPartitions(w http.ResponseWriter, r *http.Request) {
	_, t, ok := h.resolveKafkaTopic(w, r)
	if !ok {
		return
	}

	var partitions []kafkaPartition
	for _, p := range t.Partitions {
		partitions = append(partitions, newPartition(p))
	}
	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].Id < partitions[j].Id
	})
	write(w, partitions)
}

func (h *handler) getKafkaPartition(w http.ResponseWriter, r *http.Request) {
	_, _, p, ok := h.resolveKafkaPartition(w, r)
	if !ok {
		return
	}

	write(w, newPartition(p))
}

func (h *handler) getKafkaMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ki, t, p, ok := h.resolveKafkaPartition(w, r)
	if !ok {
		return
	}

	startOffset := -1
	if startOffsetValue, ok := vars["startOffset"]; ok {
		var err error
		startOffset, err = strconv.Atoi(startOffsetValue)
		if err != nil {
			writeError(w, fmt.Errorf("error startOffset is not an integer"), http.StatusBadRequest)
			return
		}
	}

	c := store.NewClient(ki.Store, h.app.Monitor.Kafka)
	ct := media.ParseContentType(r.Header.Get("Accept"))
	records, err := c.Read(t.Name, p.Index, int64(startOffset), &ct)
	if err != nil {
		if errors.Is(err, store.TopicNotFound) || errors.Is(err, store.PartitionNotFound) {
			writeError(w, err, http.StatusNotFound)
		} else {
			writeError(w, err, http.StatusInternalServerError)
		}
		return
	}
	write(w, records)
}

func (h *handler) getKafkaMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ki, t, p, ok := h.resolveKafkaPartition(w, r)
	if !ok {
		return
	}

	offset := -1
	offsetValue := vars["offset"]
	var err error
	offset, err = strconv.Atoi(offsetValue)
	if err != nil {
		writeError(w, fmt.Errorf("error offset is not an integer"), http.StatusBadRequest)
		return
	}

	c := store.NewClient(ki.Store, h.app.Monitor.Kafka)
	ct := media.ParseContentType(r.Header.Get("Accept"))
	record, err := c.Offset(t.Name, p.Index, int64(offset), &ct)
	if err != nil {
		if errors.Is(err, store.TopicNotFound) || errors.Is(err, store.PartitionNotFound) || errors.Is(err, store.OffsetOutOfRange) {
			writeError(w, err, http.StatusNotFound)
		} else {
			writeError(w, err, http.StatusInternalServerError)
		}
		return
	}
	write(w, record)
}

func (h *handler) getKafkaGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	ki := h.app.Kafka.Get(vars["cluster"])
	if ki == nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("kafka cluster not found"))
		return
	}

	g, ok := ki.Group(vars["group"])
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("kafka group not found"))
		return
	}

	grp := group{
		Name:    g.Name,
		State:   g.State.String(),
		Metrics: getGroupMetrics(g.Name, ki, h.app.Monitor.Kafka),
	}

	if g.Generation != nil {
		grp.Generation = g.Generation.Id
		grp.Leader = g.Generation.LeaderId
		grp.Protocol = g.Generation.Protocol

		for id, m := range g.Generation.Members {
			grp.Members = append(grp.Members, kafkaMember{
				Name:                  id,
				ClientId:              m.Client.ClientId,
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
	} else {
		grp.Generation = -1
	}
	for topicName := range g.Commits {
		grp.Topics = append(grp.Topics, topicName)
	}
	sort.Slice(grp.Topics, func(i, j int) bool {
		return strings.Compare(grp.Topics[i], grp.Topics[j]) < 0
	})

	write(w, grp)
}

func (h *handler) produceKafkaMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	ki := h.app.Kafka.Get(vars["cluster"])
	if ki == nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("kafka cluster not found"))
		return
	}

	if t := ki.Topic(vars["topic"]); t == nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("kafka topic not found"))
		return
	}

	records, err := getProduceRecords(r)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	if idValue, ok := vars["partition"]; ok {
		var id int
		id, err = strconv.Atoi(idValue)
		if err != nil {
			writeError(w, fmt.Errorf("error partition ID is not an integer"), http.StatusBadRequest)
			return
		}
		for _, record := range records {
			record.Partition = id
		}
	}

	c := store.NewClient(ki.Store, h.app.Monitor.Kafka)
	ct := media.ParseContentType(r.Header.Get("Content-Type"))
	result, err := c.Write(vars["topic"], records, ct)
	if err != nil {
		if errors.Is(err, store.TopicNotFound) || errors.Is(err, store.PartitionNotFound) {
			writeError(w, err, http.StatusNotFound)
		} else {
			writeError(w, err, http.StatusBadRequest)
		}
	}
	res := KafkaProduceResponse{}
	for _, rec := range result {
		res.Offsets = append(res.Offsets, KafkaRecordResult{
			Partition: rec.Partition,
			Offset:    rec.Offset,
			Error:     rec.Error,
		})
	}
	write(w, res)
}

func getTopicInfos(ki *runtime.KafkaInfo, m *monitor.Kafka) []kafkaTopicInfo {
	topics := make([]kafkaTopicInfo, 0, len(ki.Config.Channels))
	for name, ch := range ki.Config.Channels {
		if ch.Value == nil {
			continue
		}
		if !ch.Value.IsChannelAvailable("kafka") {
			continue
		}
		addr := ch.Value.Address
		if addr == "" {
			addr = name
		}

		ti := kafkaTopicInfo{
			Name:    addr,
			Summary: ch.Value.Summary,
			Tags:    getKafkaTags(ch.Value),
		}

		ti.Metrics = kafkaTopicMetric{
			NumMessages:     m.Messages.Sum(metrics.NewQuery(metrics.ByLabel("service", ki.Info.Name))),
			LastMessageTime: m.LastMessage.Max(metrics.NewQuery(metrics.ByLabel("service", ki.Info.Name))),
		}

		topics = append(topics, ti)
	}
	sort.Slice(topics, func(i, j int) bool {
		return strings.Compare(topics[i].Name, topics[j].Name) < 0
	})
	return topics
}

func newTopic(t *store.Topic, ki *runtime.KafkaInfo, ch *asyncapi3.Channel, m *monitor.Kafka) kafkaTopic {
	var partitions []kafkaPartition
	for _, p := range t.Partitions {
		partitions = append(partitions, newPartition(p))
	}
	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].Id < partitions[j].Id
	})

	result := kafkaTopic{
		Name:        t.Name,
		Title:       ch.Title,
		Summary:     ch.Summary,
		Description: ch.Description,
		Partitions:  partitions,
		Bindings: kafkaBindings{
			Partitions:            t.Config.Bindings.Kafka.Partitions,
			RetentionBytes:        t.Config.Bindings.Kafka.RetentionBytes,
			RetentionMs:           t.Config.Bindings.Kafka.RetentionMs,
			SegmentBytes:          t.Config.Bindings.Kafka.SegmentBytes,
			SegmentMs:             t.Config.Bindings.Kafka.SegmentMs,
			ValueSchemaValidation: t.Config.Bindings.Kafka.ValueSchemaValidation,
			KeySchemaValidation:   t.Config.Bindings.Kafka.KeySchemaValidation,
		},
		Tags:   getKafkaTags(ch),
		Groups: getGroupInfos(ki, t.Name, m),
	}

	result.Messages = getMessageConfigs(ch, ki.Config)

	return result
}

func getPartitions(t *store.Topic) []kafkaPartition {
	var partitions []kafkaPartition
	for _, p := range t.Partitions {
		partitions = append(partitions, newPartition(p))
	}
	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].Id < partitions[j].Id
	})
	return partitions
}

func getGroupInfos(ki *runtime.KafkaInfo, topic string, m *monitor.Kafka) []kafkaGroupInfo {
	groups := ki.Groups()
	var result []kafkaGroupInfo
Groups:
	for _, g := range groups {
		if topic != "" {
			for topicName := range g.Commits {
				if topicName != topic {
					continue Groups
				}
			}
		}

		gi := kafkaGroupInfo{
			Name:    g.Name,
			Metrics: getGroupMetrics(g.Name, ki, m),
		}
		if g.Generation != nil {
			gi.Generation = g.Generation.Id
			gi.State = g.State.String()
			gi.Protocol = g.Generation.Protocol
			gi.Members = len(g.Generation.Members)
		}
		result = append(result, gi)
	}
	slices.SortFunc(result, func(a, b kafkaGroupInfo) int {
		return strings.Compare(a.Name, b.Name)
	})
	return result
}

func getGroupMetrics(groupName string, ki *runtime.KafkaInfo, m *monitor.Kafka) kafkaGroupMetrics {
	result := kafkaGroupMetrics{
		LastRebalancing: m.LastRebalancing.Value(metrics.NewQuery(metrics.ByLabel("service", ki.Info.Name))),
	}
	lags := m.Lags.FindAll(
		metrics.NewQuery(
			metrics.ByLabel("service", ki.Info.Name),
			metrics.ByLabel("group", groupName),
		),
	)
	if len(lags) > 0 {
		result.Topics = map[string][]kafkaTopicGroupMetrics{}
		for _, lag := range lags {
			topic := lag.Info().GetLabel("topic")
			if _, ok := result.Topics[topic]; !ok {
				result.Topics[topic] = []kafkaTopicGroupMetrics{}
			}
			partition := lag.Info().GetLabel("partition")
			id, _ := strconv.Atoi(partition)
			mCommit, _ := m.Commits.FindOne(metrics.NewQuery(
				metrics.ByLabel("service", ki.Info.Name),
				metrics.ByLabel("group", groupName),
				metrics.ByLabel("topic", topic),
				metrics.ByLabel("partition", partition),
			))
			commit := float64(0)
			if mCommit != nil {
				commit = mCommit.Value()
			}

			result.Topics[topic] = append(result.Topics[topic], kafkaTopicGroupMetrics{
				Partition: id,
				Lags:      lag.Value(),
				Commit:    commit,
			})
		}
	}
	return result
}

func newGroup(g *store.Group) group {
	grp := group{
		Name:  g.Name,
		State: g.State.String(),
	}
	if g.Generation != nil {
		grp.Generation = g.Generation.Id
		grp.Leader = g.Generation.LeaderId
		grp.Protocol = g.Generation.Protocol

		for id, m := range g.Generation.Members {
			grp.Members = append(grp.Members, kafkaMember{
				Name:                  id,
				ClientId:              m.Client.ClientId,
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
	} else {
		grp.Generation = -1
	}
	for topicName := range g.Commits {
		grp.Topics = append(grp.Topics, topicName)
	}
	sort.Slice(grp.Topics, func(i, j int) bool {
		return strings.Compare(grp.Topics[i], grp.Topics[j]) < 0
	})

	return grp
}

func newPartition(p *store.Partition) kafkaPartition {
	return kafkaPartition{
		Id:          p.Index,
		StartOffset: p.StartOffset(),
		Offset:      p.Offset(),
		Segments:    len(p.Segments),
	}
}

func getProduceRecords(r *http.Request) ([]store.Record, error) {
	var pr kafkaProduceRequest
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

func (r *KafkaRecordResult) MarshalJSON() ([]byte, error) {
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

func getMessageConfigs(ch *asyncapi3.Channel, cfg *asyncapi3.Config) map[string]messageConfig {
	var result map[string]messageConfig
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
			format := ""
			if msf, ok := msg.Payload.Value.(*asyncapi3.MultiSchemaFormat); ok {
				format = msf.Format
			}
			s, err := msg.Payload.GetSchema()
			if err != nil {
				log.Errorf("failed to get schema for message in topic '%s': %v", ch.Name, err)
			}
			m.Payload = &schemaInfo{Schema: s, Format: format}
		}
		if msg.Headers != nil && msg.Headers.Value != nil {
			format := ""
			if msf, ok := msg.Headers.Value.(*asyncapi3.MultiSchemaFormat); ok {
				format = msf.Format
			}
			s, err := msg.Headers.GetSchema()
			if err != nil {
				log.Errorf("failed to get schema for headers in topic '%s': %v", ch.Name, err)
			}
			m.Header = &schemaInfo{Schema: s, Format: format}
		}

		if m.ContentType == "" {
			m.ContentType = cfg.DefaultContentType
		}

		if msg.Bindings.Kafka.Key != nil {
			s, err := msg.Bindings.Kafka.Key.GetSchema()
			if err != nil {
				log.Errorf("failed to get schema for key in topic '%s': %v", ch.Name, err)
			}
			m.Key = &schemaInfo{Schema: s}
		}
		if result == nil {
			result = map[string]messageConfig{}
		}
		result[messageId] = m
	}
	return result
}

func getKafkaTags(ch *asyncapi3.Channel) []kafkaTag {
	var result []kafkaTag
	for _, tRef := range ch.Tags {
		if tRef.Value == nil {
			continue
		}
		result = append(result, kafkaTag{
			Name:        tRef.Value.Name,
			Description: tRef.Value.Description,
		})
	}
	return result
}

func (h *handler) resolveKafkaTopic(w http.ResponseWriter, r *http.Request) (*runtime.KafkaInfo, *store.Topic, bool) {
	vars := mux.Vars(r)

	ki := h.app.Kafka.Get(vars["cluster"])
	if ki == nil {
		http.Error(w, "kafka cluster not found", http.StatusNotFound)
		return nil, nil, false
	}

	t := ki.Topic(vars["topic"])
	if t == nil {
		http.Error(w, "kafka topic not found", http.StatusNotFound)
		return nil, nil, false
	}

	return ki, t, true
}

func (h *handler) resolveKafkaPartition(w http.ResponseWriter, r *http.Request) (*runtime.KafkaInfo, *store.Topic, *store.Partition, bool) {
	vars := mux.Vars(r)

	ki, t, ok := h.resolveKafkaTopic(w, r)
	if !ok {
		return nil, nil, nil, false
	}

	id, err := strconv.Atoi(vars["partition"])
	if err != nil {
		writeError(w, fmt.Errorf("partition ID is not an integer"), http.StatusBadRequest)
		return nil, nil, nil, false
	}

	p := t.Partition(id)
	if p == nil {
		http.Error(w, "kafka partition not found", http.StatusNotFound)
		return nil, nil, nil, false
	}

	return ki, t, p, true
}

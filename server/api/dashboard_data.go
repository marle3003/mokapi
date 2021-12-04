package api

import (
	"mokapi/engine"
	"mokapi/models"
	"time"
)

type dashboard struct {
	ServerUptime      time.Time              `json:"serverUptime"`
	TotalRequests     int64                  `json:"totalRequests"`
	RequestsWithError int64                  `json:"requestsWithError"`
	LastErrors        []requestSummary       `json:"lastErrors"`
	LastRequests      []requestSummary       `json:"lastRequests"`
	MemoryUsage       int64                  `json:"memoryUsage"`
	Services          []models.ServiceMetric `json:"services"`
	Kafka             kafka                  `json:"kafka"`
	LastMails         []mailSummary          `json:"lastMails"`
	TotalMails        int64                  `json:"totalMails"`

	HttpEnabled  bool `json:"httpEnabled"`
	KafkaEnabled bool `json:"kafkaEnabled"`
	LdapEnabled  bool `json:"ldapEnabled"`
	SmtpEnabled  bool `json:"smtpEnabled"`
}

type requestSummary struct {
	Id           string        `json:"id"`
	Method       string        `json:"method"`
	Url          string        `json:"url"`
	HttpStatus   int           `json:"httpStatus"`
	IsError      bool          `json:"isError"`
	Time         time.Time     `json:"time"`
	ResponseTime time.Duration `json:"responseTime"`
}

type request struct {
	requestSummary
	Parameters   []requestParameter `json:"parameters"`
	ContentType  string             `json:"contentType"`
	ResponseBody string             `json:"responseBody"`
	eventSummary []eventSummary     `json:"eventSummary"`
}

type requestParameter struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	Raw   string `json:"raw"`
}

type eventSummary struct {
	Duration time.Duration     `json:"duration"`
	Tags     map[string]string `json:"tags"`
}

type kafka struct {
	Topics []topic `json:"topics"`
	Groups []group `json:"groups"`
}

type topic struct {
	Service    string      `json:"service"`
	Name       string      `json:"name"`
	LastRecord time.Time   `json:"lastRecord"`
	Partitions []partition `json:"partitions"`
	Count      int64       `json:"count"`
}

type partition struct {
	Id          int    `json:"id"`
	StartOffset int64  `json:"startOffset"`
	Offset      int64  `json:"offset"`
	Size        int64  `json:"size"`
	Leader      string `json:"leader"`
	Segments    int    `json:"segments"`
}

type group struct {
	Name               string   `json:"name"`
	Members            []string `json:"members"`
	Coordinator        string   `json:"coordinator"`
	Leader             string   `json:"leader"`
	State              string   `json:"state"`
	AssignmentStrategy string   `json:"assignmentStrategy"`
}

type topicGroup struct {
	group
	Lag int64 `json:"lag"`
}

func newDashboard(runtime *models.Runtime) dashboard {
	dashboard := dashboard{LastErrors: make([]requestSummary, 0), Services: make([]models.ServiceMetric, 0), Kafka: kafka{}}
	dashboard.ServerUptime = runtime.Metrics.Start
	dashboard.TotalRequests = runtime.Metrics.TotalRequests
	dashboard.RequestsWithError = runtime.Metrics.RequestsWithError
	dashboard.LastRequests = make([]requestSummary, 0)
	dashboard.LastMails = make([]mailSummary, 0)
	dashboard.TotalMails = runtime.Metrics.TotalMails
	dashboard.MemoryUsage = runtime.Metrics.Memory

	for _, s := range runtime.Metrics.OpenApi {
		dashboard.Services = append(dashboard.Services, *s)
	}

	for _, t := range runtime.Metrics.Kafka.Topics {
		if len(t.Service) > 0 {
			dashboard.Kafka.Topics = append(dashboard.Kafka.Topics, newTopic(t))
		}
	}

	for name, g := range runtime.Metrics.Kafka.Groups {
		dashboard.Kafka.Groups = append(dashboard.Kafka.Groups, newGroup(name, g))
	}

	for _, r := range runtime.Metrics.LastErrorRequests {
		dashboard.LastErrors = append(dashboard.LastErrors, newRequestSummary(r))
	}

	for _, r := range runtime.Metrics.LastRequests {
		dashboard.LastRequests = append(dashboard.LastRequests, newRequestSummary(r))
	}

	for _, m := range runtime.Metrics.LastMails {
		dashboard.LastMails = append(dashboard.LastMails, newMailSummary(m.Mail))
	}

	dashboard.HttpEnabled = len(runtime.OpenApi) > 0
	dashboard.KafkaEnabled = len(runtime.AsyncApi) > 0
	dashboard.LdapEnabled = len(runtime.Ldap) > 0
	dashboard.SmtpEnabled = len(runtime.Smtp) > 0

	return dashboard
}

func newRequestSummary(r *models.RequestMetric) requestSummary {
	return requestSummary{
		Id:           r.Id,
		Method:       r.Method,
		Url:          r.Url,
		HttpStatus:   r.HttpStatus,
		Time:         r.Time,
		ResponseTime: r.ResponseTime,
		IsError:      r.IsError,
	}
}

func newRequest(r *models.RequestMetric) request {
	result := request{
		requestSummary: requestSummary{
			Id:           r.Id,
			Method:       r.Method,
			IsError:      r.IsError,
			Url:          r.Url,
			HttpStatus:   r.HttpStatus,
			Time:         r.Time,
			ResponseTime: r.ResponseTime,
		},
		ContentType:  r.ContentType,
		ResponseBody: r.ResponseBody,
	}
	for _, p := range r.Parameters {
		result.Parameters = append(result.Parameters, requestParameter{
			Name:  p.Name,
			Type:  p.Type,
			Value: p.Value,
			Raw:   p.Raw,
		})
	}
	for _, e := range r.EventSummary {
		result.eventSummary = append(result.eventSummary, newEventSummary(e))
	}
	return result
}

func newEventSummary(s *engine.Summary) eventSummary {
	result := eventSummary{
		Duration: s.Duration,
		Tags:     s.Tags,
	}

	return result
}

func newTopic(t *models.KafkaTopic) topic {
	result := topic{
		Service:    t.Service,
		Name:       t.Name,
		LastRecord: t.LastRecord,
		Count:      t.Count,
	}
	for _, p := range t.Partitions {
		result.Partitions = append(result.Partitions, newPartition(p))
	}
	return result
}

func newPartition(p *models.KafkaPartition) partition {
	return partition{
		Id:          p.Index,
		StartOffset: p.StartOffset,
		Offset:      p.Offset,
		Size:        p.Size,
		Leader:      p.Leader,
		Segments:    p.Segments,
	}
}

func newGroup(name string, g *models.KafkaGroup) group {
	return group{
		Name:               name,
		Coordinator:        g.Coordinator,
		Leader:             g.Leader,
		State:              g.State,
		Members:            g.Members,
		AssignmentStrategy: g.AssignmentStrategy,
	}
}

func newTopicGroup(name string, g *models.KafkaTopicGroup) topicGroup {
	return topicGroup{
		group: newGroup(name, &g.KafkaGroup),
		Lag:   g.Lag,
	}
}

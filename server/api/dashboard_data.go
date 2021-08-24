package api

import (
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
	KafkaTopics       []*models.KafkaTopic   `json:"kafkaTopics"`
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
	Workflows    []workflow         `json:"workflows"`
}

type requestParameter struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	Raw   string `json:"raw"`
}

type workflow struct {
	Name     string        `json:"name"`
	Logs     []string      `json:"logs"`
	Duration time.Duration `json:"duration"`
	Status   string        `json:"status"`
}

func newDashboard(runtime *models.Runtime) dashboard {
	dashboard := dashboard{LastErrors: make([]requestSummary, 0), Services: make([]models.ServiceMetric, 0)}
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
		dashboard.KafkaTopics = append(dashboard.KafkaTopics, t)
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
	for _, w := range r.WorkflowLogs {
		result.Workflows = append(result.Workflows, newWorkflow(w))
	}
	return result
}

func newWorkflow(w *models.WorkflowLog) workflow {
	result := workflow{
		Name:     w.Name,
		Duration: w.Duration,
		Logs:     w.Logs,
	}

	//switch s.Status {
	//case runtime2.Error:
	//	result.Status = "error"
	//case runtime2.Successful:
	//	result.Status = "successful"
	//}

	return result
}

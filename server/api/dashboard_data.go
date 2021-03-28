package api

import (
	"mokapi/models"
	"time"
)

type dashboard struct {
	ServerUptime      time.Time              `json:"serverUptime"`
	TotalRequests     int64                  `json:"totalRequests"`
	RequestsWithError int64                  `json:"requestsWithError"`
	LastErrors        []request              `json:"lastErrors"`
	LastRequests      []request              `json:"lastRequests"`
	MemoeryUsage      int64                  `json:"memoryUsage"`
	Services          []models.ServiceMetric `json:"services"`
	KafkaTopics       []models.KafkaTopic    `json:"kafkaTopics"`
}

type serviceStatus struct {
	Total  int `json:"total"`
	Errors int `json:"errors"`
}

type request struct {
	Method       string        `json:"method"`
	Url          string        `json:"url"`
	HttpStatus   int           `json:"httpStatus"`
	Error        string        `json:"error"`
	Time         time.Time     `json:"time"`
	ResponseTime time.Duration `json:"responseTime"`
}

func newDashboard(runtime *models.Runtime) dashboard {
	dashboard := dashboard{LastErrors: make([]request, 0), Services: make([]models.ServiceMetric, 0)}
	dashboard.ServerUptime = runtime.Metrics.Start
	dashboard.TotalRequests = runtime.Metrics.TotalRequests
	dashboard.RequestsWithError = runtime.Metrics.RequestsWithError
	dashboard.LastRequests = make([]request, 0)
	dashboard.MemoeryUsage = runtime.Metrics.Memory

	for _, s := range runtime.Metrics.OpenApi {
		dashboard.Services = append(dashboard.Services, *s)
	}

	for _, t := range runtime.Metrics.Kafka.Topics {
		dashboard.KafkaTopics = append(dashboard.KafkaTopics, t)
	}

	//for _, s := range app.OpenApi {
	//	if len(s.Errors) > 0 {
	//		dashboard.ServiceStatus.Errors++
	//	}
	//}

	for _, r := range runtime.Metrics.LastErrorRequests {
		dashboard.LastErrors = append(dashboard.LastErrors, newRequest(r))
	}

	for _, r := range runtime.Metrics.LastRequests {
		dashboard.LastRequests = append(dashboard.LastRequests, newRequest(r))
	}

	return dashboard
}

func newRequest(r *models.RequestMetric) request {
	return request{Method: r.Method, Error: r.Error, Url: r.Url, HttpStatus: r.HttpStatus, Time: r.Time, ResponseTime: r.ResponseTime}
}

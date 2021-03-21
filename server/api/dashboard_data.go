package api

import (
	"mokapi/models"
	"time"
)

type dashboard struct {
	ServerUptime      time.Time     `json:"serverUptime"`
	TotalRequests     int64         `json:"totalRequests"`
	RequestsWithError int64         `json:"requestsWithError"`
	ServiceStatus     serviceStatus `json:"serviceStatus"`
	LastErrors        []request     `json:"lastErrors"`
	LastRequests      []request     `json:"lastRequests"`
	MemoeryUsage      int64         `json:"memoryUsage"`
	TotalServices     int           `json:"totalServices"`
	Kafka             kafkaSummary  `json:"kafka"`
}

type kafkaSummary struct {
	TotalClusters   int              `json:"totalClusters"`
	TotalTopics     int              `json:"totalTopics"`
	TotalPartitions int              `json:"totalPartitions"`
	TotalSegments   int              `json:"totalSegments"`
	TotalMessages   int64            `json:"totalMessages"`
	TopicSizes      map[string]int64 `json:"topicSizes"`
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
	dashboard := dashboard{ServiceStatus: serviceStatus{}, LastErrors: make([]request, 0)}
	dashboard.ServerUptime = runtime.Metrics.Start
	dashboard.TotalRequests = runtime.Metrics.TotalRequests
	dashboard.RequestsWithError = runtime.Metrics.RequestsWithError
	dashboard.LastRequests = make([]request, 0)
	dashboard.MemoeryUsage = runtime.Metrics.Memory
	dashboard.TotalServices = len(runtime.OpenApi)

	dashboard.ServiceStatus.Total = len(runtime.OpenApi)
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

	dashboard.Kafka = kafkaSummary{}
	dashboard.Kafka.TotalClusters = len(runtime.Metrics.Kafka)
	for _, k := range runtime.Metrics.Kafka {
		dashboard.Kafka.TotalTopics += k.Topics
		dashboard.Kafka.TotalPartitions += k.Partitions
		dashboard.Kafka.TotalSegments += k.Segments
		dashboard.Kafka.TotalMessages += k.Messages
		dashboard.Kafka.TopicSizes = k.TopicSizes
	}

	return dashboard
}

func newRequest(r *models.RequestMetric) request {
	return request{Method: r.Method, Error: r.Error, Url: r.Url, HttpStatus: r.HttpStatus, Time: r.Time, ResponseTime: r.ResponseTime}
}

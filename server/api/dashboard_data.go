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

func newDashboard(app *models.Application) dashboard {
	dashboard := dashboard{ServiceStatus: serviceStatus{}, LastErrors: make([]request, 0)}
	dashboard.ServerUptime = app.Metrics.Start
	dashboard.TotalRequests = app.Metrics.TotalRequests
	dashboard.RequestsWithError = app.Metrics.RequestsWithError

	dashboard.ServiceStatus.Total = len(app.WebServices)
	for _, s := range app.WebServices {
		if len(s.Errors) > 0 {
			dashboard.ServiceStatus.Errors++
		}
	}

	for _, r := range app.Metrics.LastErrorRequests {
		dashboard.LastErrors = append(dashboard.LastErrors, newRequest(r))
	}

	for _, r := range app.Metrics.LastRequests {
		dashboard.LastRequests = append(dashboard.LastRequests, newRequest(r))
	}

	return dashboard
}

func newRequest(r *models.RequestMetric) request {
	return request{Method: r.Method, Error: r.Error, Url: r.Url, HttpStatus: r.HttpStatus, Time: r.Time, ResponseTime: r.ResponseTime}
}

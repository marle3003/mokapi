package api

import (
	"mokapi/models"
	"time"
)

type dashboard struct {
	ServerUptime      time.Time       `json:"serverUptime"`
	TotalRequests     int64           `json:"totalRequests"`
	RequestsWithError int64           `json:"requestsWithError"`
	ServiceStatus     serviceStatus   `json:"serviceStatus"`
	LastErrors        []requestErrors `json:"lastErrors"`
}

type serviceStatus struct {
	Total  int `json:"total"`
	Errors int `json:"errors"`
}

type requestErrors struct {
	Method     string    `json:"method"`
	Url        string    `json:"url"`
	HttpStatus int       `json:"httpStatus"`
	Error      string    `json:"error"`
	Time       time.Time `json:"time"`
}

func newDashboard(app *models.Application) dashboard {
	dashboard := dashboard{ServiceStatus: serviceStatus{}, LastErrors: make([]requestErrors, 0)}
	dashboard.ServerUptime = app.Metrics.Start
	dashboard.TotalRequests = app.Metrics.TotalRequests
	dashboard.RequestsWithError = app.Metrics.RequestsWithError

	dashboard.ServiceStatus.Total = len(app.Services)
	for _, s := range app.Services {
		if len(s.Errors) > 0 {
			dashboard.ServiceStatus.Errors++
		}
	}

	for _, r := range app.Metrics.LastRequests {
		if len(r.Error) > 0 {
			dashboard.LastErrors = append(dashboard.LastErrors, newRequestError(r))
		}
	}

	return dashboard
}

func newRequestError(r *models.RequestMetric) requestErrors {
	return requestErrors{Method: r.Method, Error: r.Error, Url: r.Url, HttpStatus: r.HttpStatus, Time: r.Time}
}

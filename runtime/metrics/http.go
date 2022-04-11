package metrics

type HttpMetrics struct {
	RequestCounter      *CounterMap
	RequestErrorCounter *CounterMap
}

func NewHttp() *HttpMetrics {
	return &HttpMetrics{
		RequestCounter:      NewCounterMap("http_requests_total"),
		RequestErrorCounter: NewCounterMap("http_requests_total.errors"),
	}
}

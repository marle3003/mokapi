package metrics

type HttpMetrics struct {
	RequestCounter      *Counter
	RequestErrorCounter *Counter
}

func NewHttp() *HttpMetrics {
	return &HttpMetrics{
		RequestCounter:      NewCounter("http.requests.total"),
		RequestErrorCounter: NewCounter("http.requests.total.errors"),
	}
}

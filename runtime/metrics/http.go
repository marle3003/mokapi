package metrics

type Http struct {
	RequestCounter      *CounterMap
	RequestErrorCounter *CounterMap
	LastRequest         *GaugeMap
}

func NewHttp() *Http {
	return &Http{
		RequestCounter:      NewCounterMap("http_requests_total"),
		RequestErrorCounter: NewCounterMap("http_requests_total.errors"),
		LastRequest:         NewGaugeMap("http_request_time"),
	}
}

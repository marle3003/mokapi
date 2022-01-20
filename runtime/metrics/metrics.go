package metrics

type Counter struct {
	label string
	value float64
}

func NewCounter(label string) *Counter {
	return &Counter{label: label}
}

func (c *Counter) Add(v float64) {
	c.value += v
}

func (c *Counter) Value() float64 {
	return c.value
}

type Metrics struct {
	Http *HttpMetrics
}

func New() *Metrics {
	return &Metrics{
		Http: NewHttp(),
	}
}

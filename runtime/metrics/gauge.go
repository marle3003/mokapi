package metrics

type Gauge struct {
	Name  string
	value float64
}

func NewGauge(name string) *Gauge {
	return &Gauge{
		Name: name,
	}
}

func (g *Gauge) Set(v float64) {
	g.value = v
}

func (g *Gauge) Value() float64 {
	return g.value
}

type GaugeMap struct {
	Name   string
	gauges map[string]*Gauge
}

func NewGaugeMap(name string) *GaugeMap {
	return &GaugeMap{Name: name, gauges: make(map[string]*Gauge)}
}

func (m *GaugeMap) WithLabel(label string) *Gauge {
	c, ok := m.gauges[label]
	if !ok {
		c = NewGauge(label)
		m.gauges[label] = c
	}
	return c
}

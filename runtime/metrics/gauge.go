package metrics

import (
	"encoding/json"
	"fmt"
	"sync"
)

type Gauge struct {
	info  *Info
	value float64
}

func NewGauge(opt ...Options) *Gauge {
	o := &options{}
	for _, opt := range opt {
		opt(o)
	}
	return &Gauge{
		info: &Info{Namespace: o.namespace, Name: o.name, labels: o.labels},
	}
}

func (g *Gauge) Set(v float64) {
	g.value = v
}

func (g *Gauge) Value() float64 {
	return g.value
}

func (g *Gauge) Info() *Info {
	return g.info
}

func (g *Gauge) MarshalJSON() ([]byte, error) {
	aux := &struct {
		Name  string  `json:"name"`
		Value float64 `json:"value"`
	}{
		Name:  g.Info().String(),
		Value: g.Value(),
	}
	return json.Marshal(aux)
}

func (g *Gauge) Collect(ch chan<- Metric) {
	ch <- g
}

type GaugeMap struct {
	info     *Info
	gauges   map[uint32]*Gauge
	newGauge func(values []string) *Gauge
	m        sync.Mutex
}

func NewGaugeMap(opts ...Options) *GaugeMap {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	info := &Info{Namespace: o.namespace, Name: o.name}

	return &GaugeMap{
		info:   info,
		gauges: make(map[uint32]*Gauge),
		newGauge: func(values []string) *Gauge {
			if len(o.labelNames) != len(values) {
				panic(fmt.Sprintf("invalid labels values for %v", info.String()))
			}
			labels := make([]*Label, 0)
			for i, v := range values {
				labels = append(labels, &Label{
					Name:  o.labelNames[i],
					Value: v,
				})
			}

			return &Gauge{
				info: &Info{
					Namespace: info.Namespace,
					Name:      info.Name,
					labels:    labels,
				},
				value: 0,
			}
		},
	}
}

func (m *GaugeMap) Info() *Info {
	return m.info
}

func (m *GaugeMap) WithLabel(values ...string) *Gauge {
	m.m.Lock()
	defer m.m.Unlock()

	key := hash(values)
	c, ok := m.gauges[key]
	if !ok {
		c = m.newGauge(values)
		m.gauges[key] = c
	}
	return c
}

func (m *GaugeMap) Value(query *Query) float64 {
	m.m.Lock()
	defer m.m.Unlock()

	var v float64
	for _, c := range m.gauges {
		if c.Info().Match(query) {
			v += c.Value()
		}
	}
	return v
}

func (m *GaugeMap) FindAll(query *Query) []Metric {
	m.m.Lock()
	defer m.m.Unlock()

	result := make([]Metric, 0)
	for _, c := range m.gauges {
		if c.Info().Match(query) {
			result = append(result, c)
		}
	}
	return result
}

func (m *GaugeMap) Collect(ch chan<- Metric) {
	m.m.Lock()
	defer m.m.Unlock()

	for _, g := range m.gauges {
		ch <- g
	}
}

func (m *GaugeMap) Reset() {
	m.m.Lock()
	defer m.m.Unlock()

	for _, g := range m.gauges {
		g.Set(0)
	}
}

func (m *GaugeMap) MarshalJSON() ([]byte, error) {
	m.m.Lock()
	defer m.m.Unlock()

	if len(m.gauges) == 0 {
		return json.Marshal(&struct {
			Name  string  `json:"name"`
			Value float64 `json:"value"`
		}{
			Name:  m.info.String(),
			Value: 0,
		})
	}
	result := make([]*Gauge, 0, len(m.gauges))
	for _, g := range m.gauges {
		result = append(result, g)
	}
	return json.Marshal(result)
}

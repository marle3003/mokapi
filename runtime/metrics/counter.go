package metrics

import (
	"encoding/json"
	"fmt"
	"sync"
)

type Counter struct {
	info  *Info
	value float64
}

func NewCounter(opt ...Options) *Counter {
	o := &options{}
	for _, opt := range opt {
		opt(o)
	}
	return &Counter{info: &Info{Namespace: o.namespace, Name: o.name, labels: o.labels}}
}

func (c *Counter) Add(v float64) {
	c.value += v
}

func (c *Counter) Value() float64 {
	return c.value
}

func (c *Counter) Info() *Info {
	return c.info
}

func (c *Counter) Collect(ch chan<- Metric) {
	ch <- c
}

func (c *Counter) Reset() {
	c.value = 0
}

func (c *Counter) MarshalJSON() ([]byte, error) {
	aux := &struct {
		Name  string  `json:"name"`
		Value float64 `json:"value"`
	}{
		Name:  c.Info().String(),
		Value: c.Value(),
	}
	return json.Marshal(aux)
}

type CounterMap struct {
	info       *Info
	counters   map[uint32]*Counter
	newCounter func(values []string) *Counter
	m          sync.Mutex
}

func NewCounterMap(opts ...Options) *CounterMap {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	info := &Info{Namespace: o.namespace, Name: o.name}

	return &CounterMap{
		info:     info,
		counters: make(map[uint32]*Counter),
		newCounter: func(values []string) *Counter {
			if len(o.labelNames) != len(values) {
				panic(fmt.Sprintf("invalid labels values for %v", info))
			}
			labels := make([]*Label, 0)
			for i, v := range values {
				labels = append(labels, &Label{
					Name:  o.labelNames[i],
					Value: v,
				})
			}

			return &Counter{
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

func (m *CounterMap) Info() *Info {
	return m.info
}

func (m *CounterMap) WithLabel(values ...string) *Counter {
	m.m.Lock()
	defer m.m.Unlock()

	key := hash(values)
	c, ok := m.counters[key]
	if !ok {
		c = m.newCounter(values)
		m.counters[key] = c
	}
	return c
}

func (m *CounterMap) Value(query *Query) float64 {
	m.m.Lock()
	defer m.m.Unlock()

	var v float64
	for _, c := range m.counters {
		if c.Info().Match(query) {
			v += c.Value()
		}
	}
	return v
}

func (m *CounterMap) Sum() float64 {
	m.m.Lock()
	defer m.m.Unlock()

	var v float64
	for _, c := range m.counters {
		v += c.Value()
	}
	return v
}

func (m *CounterMap) Collect(ch chan<- Metric) {
	m.m.Lock()
	defer m.m.Unlock()

	for _, c := range m.counters {
		ch <- c
	}
}

func (m *CounterMap) Reset() {
	m.m.Lock()
	defer m.m.Unlock()

	for _, c := range m.counters {
		c.Reset()
	}
}

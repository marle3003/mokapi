package metrics

import (
	"bytes"
	"encoding/json"
)

type Counter struct {
	Name  string
	value float64
}

func NewCounter(name string) *Counter {
	return &Counter{Name: name}
}

func (c *Counter) Add(v float64) {
	c.value += v
}

func (c *Counter) Value() float64 {
	return c.value
}

type CounterMap struct {
	Name     string
	counters map[string]*Counter
}

func NewCounterMap(name string) *CounterMap {
	return &CounterMap{Name: name, counters: make(map[string]*Counter)}
}

func (m *CounterMap) WithLabel(label string) *Counter {
	c, ok := m.counters[label]
	if !ok {
		c = NewCounter(label)
		m.counters[label] = c
	}
	return c
}

func (m *CounterMap) Value() float64 {
	var v float64
	for _, c := range m.counters {
		v += c.Value()
	}
	return v
}

func (c *Counter) MarshalJSON() ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	name, err := json.Marshal(c.Name)
	if err != nil {
		return nil, err
	}
	buf.Write(name)
	buf.WriteRune(':')

	value, err := json.Marshal(c.value)
	if err != nil {
		return nil, err
	}
	buf.Write(value)

	return buf.Bytes(), nil
}

func (m *CounterMap) MarshalJSON() ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)

	name, err := json.Marshal(m.Name)
	if err != nil {
		return nil, err
	}
	buf.Write(name)
	buf.WriteRune(':')

	value, err := json.Marshal(m.Value())
	if err != nil {
		return nil, err
	}
	buf.Write(value)

	for _, c := range m.counters {
		buf.WriteRune(',')
		j, err := json.Marshal(c.Name)
		if err != nil {
			return nil, err
		}
		buf.Write(j)
	}

	return buf.Bytes(), nil
}

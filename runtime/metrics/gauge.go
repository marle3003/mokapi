package metrics

import (
	"bytes"
	"encoding/json"
)

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

func (g *Gauge) writeJSON(buf *bytes.Buffer) error {
	name, err := json.Marshal(g.Name)
	if err != nil {
		return err
	}
	buf.Write(name)
	buf.WriteRune(':')

	value, err := json.Marshal(g.value)
	if err != nil {
		return err
	}
	buf.Write(value)

	return nil
}

package metrics

import (
	"bytes"
)

type Http struct {
	RequestCounter      *CounterMap
	RequestErrorCounter *CounterMap
}

func NewHttp() *Http {
	return &Http{
		RequestCounter:      NewCounterMap("http_requests_total"),
		RequestErrorCounter: NewCounterMap("http_requests_total.errors"),
	}
}

func (hm *Http) MarshalJSON() ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	buf.WriteRune('{')

	if err := hm.RequestCounter.writeJSON(buf); err != nil {
		return nil, err
	}
	buf.WriteRune(',')
	if err := hm.RequestErrorCounter.writeJSON(buf); err != nil {
		return nil, err
	}

	buf.WriteRune('}')

	return buf.Bytes(), nil
}

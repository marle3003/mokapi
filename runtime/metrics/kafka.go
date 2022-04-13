package metrics

import "bytes"

type Kafka struct {
	Messages *CounterMap
}

func NewKafka() *Kafka {
	return &Kafka{
		Messages: NewCounterMap("kafka_messages_total"),
	}
}

func (hm *Kafka) MarshalJSON() ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	buf.WriteRune('{')

	if err := hm.Messages.writeJSON(buf); err != nil {
		return nil, err
	}

	buf.WriteRune('}')

	return buf.Bytes(), nil
}

package asyncapi3

import (
	"sort"
	"strings"
)

type ProtocolUsage struct {
	Servers  []string // Server having these protocols
	Channels []string // Channels having these protocols
}

type Analysis struct {
	HasServers  bool
	HasChannels bool
	Protocols   map[string]*ProtocolUsage // "kafka", "mqtt", "ws"
}

func (a *Analysis) Usage(proto string) *ProtocolUsage {
	if u, ok := a.Protocols[proto]; ok {
		return u
	}
	return &ProtocolUsage{}
}

func normalizeProtocol(p string) string {
	return strings.ToLower(p)
}

func (c *Config) Analyze() *Analysis {
	a := &Analysis{Protocols: map[string]*ProtocolUsage{}}
	get := func(p string) *ProtocolUsage {
		if a.Protocols[p] == nil {
			a.Protocols[p] = &ProtocolUsage{}
		}
		return a.Protocols[p]
	}

	for it := c.Servers.Iter(); it.Next(); {
		ref := it.Value()
		if ref.Value == nil {
			continue
		}
		a.HasServers = true
		if p := normalizeProtocol(ref.Value.Protocol); p != "" {
			u := get(p)
			u.Servers = append(u.Servers, it.Key())
		}
	}

	opBindings := map[*Channel][]*OperationBindings{}
	for _, op := range c.Operations {
		if op.Value != nil && op.Value.Channel.Value != nil {
			ch := op.Value.Channel.Value
			opBindings[ch] = append(opBindings[ch], op.Value.Bindings)
		}
	}

	for name, ch := range c.Channels {
		if ch.Value == nil {
			continue
		}
		a.HasChannels = true
		for _, p := range channelProtocols(ch.Value, opBindings[ch.Value]) {
			u := get(p)
			u.Channels = append(u.Channels, name)
		}
	}

	for _, u := range a.Protocols {
		sort.Strings(u.Servers)
		sort.Strings(u.Channels)
	}
	return a
}

func channelProtocols(ch *Channel, ops []*OperationBindings) []string {
	s := protocolSet{}

	ch.Bindings.collect(s)

	for _, msg := range ch.Messages {
		if msg != nil && msg.Value != nil {
			msg.Value.Bindings.collect(s)
		}
	}

	for _, b := range ops {
		b.collect(s)
	}

	return s.sorted()
}

type protocolSet map[string]struct{}

func (c *ChannelBindings) collect(s protocolSet) {
	if c == nil {
		return
	}
	s.add(c.Kafka.IsSet(), c.Mqtt.IsSet(), c.Websocket.IsSet())
}

func (o *OperationBindings) collect(s protocolSet) {
	if o == nil {
		return
	}
	s.add(o.Kafka.IsSet(), o.Mqtt.IsSet(), o.Websocket.IsSet())
}

func (m *MessageBinding) collect(s protocolSet) {
	if m == nil {
		return
	}
	s.add(m.Kafka.IsSet(), m.Mqtt.IsSet(), m.Websocket.IsSet())
}

func (s protocolSet) add(kafka, mqtt, ws bool) {
	if kafka {
		s["kafka"] = struct{}{}
	}
	if mqtt {
		s["mqtt"] = struct{}{}
	}
	if ws {
		s["ws"] = struct{}{}
	}
}

func (s protocolSet) sorted() []string {
	result := make([]string, 0, len(s))
	for p := range s {
		result = append(result, p)
	}
	sort.Strings(result)
	return result
}

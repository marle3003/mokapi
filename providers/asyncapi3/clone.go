package asyncapi3

import (
	"maps"
	openapi "mokapi/providers/openapi/schema"
	jsonSchema "mokapi/schema/json/schema"
	"mokapi/sortedmap"
	"slices"
)

// copies maps the values that were already copied to their copy. A server, a channel
// or a message is referenced from several places of a config (a channel by its
// operations, a server by the channels assigned to it), and the stores find them
// by pointer. Copying such a value once keeps these references intact.
// A copies is used for one Patch call.
type copies map[any]any

// once returns the copy of p. It creates the copy with fill when p is seen for
// the first time; the copy is registered before fill runs, so cycles end.
func once[T any](c copies, p *T, fill func(dst *T)) *T {
	if p == nil {
		return nil
	}
	if v, ok := c[p]; ok {
		return v.(*T)
	}
	dst := new(T)
	*dst = *p
	c[p] = dst
	fill(dst)
	return dst
}

// cloneMap copies a map and its values. A nil map stays nil.
func cloneMap[V any](m map[string]V, c copies, clone func(V, copies) V) map[string]V {
	if m == nil {
		return nil
	}
	r := make(map[string]V, len(m))
	for k, v := range m {
		r[k] = clone(v, c)
	}
	return r
}

// cloneSlice copies a slice and its elements. A nil slice stays nil.
func cloneSlice[V any](s []V, c copies, clone func(V, copies) V) []V {
	if s == nil {
		return nil
	}
	r := make([]V, len(s))
	for i, v := range s {
		r[i] = clone(v, c)
	}
	return r
}

// clonePointer copies the value p points to. A nil pointer stays nil.
func clonePointer[T any](p *T) *T {
	if p == nil {
		return nil
	}
	r := *p
	return &r
}

func cloneServers(m *sortedmap.LinkedHashMap[string, *ServerRef], c copies) *sortedmap.LinkedHashMap[string, *ServerRef] {
	if m == nil {
		return nil
	}
	r := m.CloneWith(func(v *ServerRef) *ServerRef { return v.clone(c) })
	return &r
}

func (r *ServerRef) clone(c copies) *ServerRef {
	if r == nil {
		return nil
	}
	// the copy keeps Reference (incl. its origin) and Value: keywords next to $ref are allowed
	d := *r
	d.Value = r.Value.clone(c)
	return &d
}

func (s *Server) clone(c copies) *Server {
	return once(c, s, func(d *Server) {
		d.Variables = cloneMap(s.Variables, c, (*ServerVariableRef).clone)
		d.Tags = cloneSlice(s.Tags, c, (*TagRef).clone)
		d.Bindings = s.Bindings.clone()
		d.ExternalDocs = cloneExternalDocRefs(s.ExternalDocs, c)
	})
}

func (r *ServerVariableRef) clone(c copies) *ServerVariableRef {
	if r == nil {
		return nil
	}
	d := *r
	d.Value = r.Value.clone()
	return &d
}

func (v *ServerVariable) clone() *ServerVariable {
	if v == nil {
		return nil
	}
	d := *v
	d.Enum = slices.Clone(v.Enum)
	d.Examples = slices.Clone(v.Examples)
	return &d
}

func (r *ChannelRef) clone(c copies) *ChannelRef {
	if r == nil {
		return nil
	}
	d := *r
	d.Value = r.Value.clone(c)
	return &d
}

// clone does not follow Config, the back pointer to the config that contains the channel.
func (ch *Channel) clone(c copies) *Channel {
	return once(c, ch, func(d *Channel) {
		d.Servers = cloneSlice(ch.Servers, c, (*ServerRef).clone)
		d.Messages = cloneMap(ch.Messages, c, (*MessageRef).clone)
		d.Parameters = cloneMap(ch.Parameters, c, (*ParameterRef).clone)
		d.Bindings = ch.Bindings.clone()
		d.Tags = cloneSlice(ch.Tags, c, (*TagRef).clone)
		d.ExternalDocs = cloneSlice(ch.ExternalDocs, c, (*ExternalDocRef).clone)
	})
}

func (r *MessageRef) clone(c copies) *MessageRef {
	if r == nil {
		return nil
	}
	d := *r
	d.Value = r.Value.clone(c)
	return &d
}

func (m *Message) clone(c copies) *Message {
	return once(c, m, func(d *Message) {
		d.CorrelationId = m.CorrelationId.clone(c)
		d.Headers = m.Headers.clone(c)
		d.Payload = m.Payload.clone(c)
		d.Bindings = m.Bindings.clone(c)
		d.Traits = cloneSlice(m.Traits, c, (*MessageTraitRef).clone)
		// examples are literal values that are never patched
		d.Examples = slices.Clone(m.Examples)
		d.ExternalDocs = cloneSlice(m.ExternalDocs, c, (*ExternalDocRef).clone)
	})
}

func (r *MessageTraitRef) clone(c copies) *MessageTraitRef {
	if r == nil {
		return nil
	}
	d := *r
	d.Value = r.Value.clone(c)
	return &d
}

func (t *MessageTrait) clone(c copies) *MessageTrait {
	if t == nil {
		return nil
	}
	d := *t
	d.CorrelationId = t.CorrelationId.clone(c)
	d.Headers = t.Headers.clone(c)
	d.Bindings = t.Bindings.clone(c)
	d.Examples = cloneSlice(t.Examples, c, (*MessageExample).clone)
	d.ExternalDocs = cloneSlice(t.ExternalDocs, c, (*ExternalDocRef).clone)
	return &d
}

func (e *MessageExample) clone(_ copies) *MessageExample {
	if e == nil {
		return nil
	}
	d := *e
	d.Headers = maps.Clone(e.Headers)
	return &d
}

func (r *OperationRef) clone(c copies) *OperationRef {
	if r == nil {
		return nil
	}
	d := *r
	d.Value = r.Value.clone(c)
	return &d
}

func (o *Operation) clone(c copies) *Operation {
	if o == nil {
		return nil
	}
	d := *o
	d.Channel = o.Channel.clone(c)
	d.Bindings = o.Bindings.clone()
	d.Traits = cloneSlice(o.Traits, c, (*OperationTraitRef).clone)
	d.Messages = cloneSlice(o.Messages, c, (*MessageRef).clone)
	d.ExternalDocs = cloneSlice(o.ExternalDocs, c, (*ExternalDocRef).clone)
	return &d
}

func (r *OperationTraitRef) clone(c copies) *OperationTraitRef {
	if r == nil {
		return nil
	}
	d := *r
	d.Value = r.Value.clone(c)
	return &d
}

func (t *OperationTrait) clone(c copies) *OperationTrait {
	if t == nil {
		return nil
	}
	d := *t
	d.Channel = t.Channel.clone(c)
	d.Bindings = t.Bindings.clone()
	d.ExternalDocs = cloneSlice(t.ExternalDocs, c, (*ExternalDocRef).clone)
	return &d
}

func (r *ParameterRef) clone(_ copies) *ParameterRef {
	if r == nil {
		return nil
	}
	d := *r
	d.Value = r.Value.clone()
	return &d
}

func (p *Parameter) clone() *Parameter {
	if p == nil {
		return nil
	}
	d := *p
	d.Enum = slices.Clone(p.Enum)
	d.Examples = slices.Clone(p.Examples)
	return &d
}

func (r *CorrelationIdRef) clone(_ copies) *CorrelationIdRef {
	if r == nil {
		return nil
	}
	d := *r
	d.Value = clonePointer(r.Value)
	return &d
}

func (r *TagRef) clone(c copies) *TagRef {
	if r == nil {
		return nil
	}
	d := *r
	d.Value = r.Value.clone(c)
	return &d
}

func (t *Tag) clone(c copies) *Tag {
	if t == nil {
		return nil
	}
	d := *t
	d.ExternalDocs = cloneSlice(t.ExternalDocs, c, (*ExternalDocRef).clone)
	return &d
}

func (r *ExternalDocRef) clone(_ copies) *ExternalDocRef {
	if r == nil {
		return nil
	}
	d := *r
	d.Value = clonePointer(r.Value)
	return &d
}

func cloneExternalDocRefs(s []ExternalDocRef, c copies) []ExternalDocRef {
	return cloneSlice(s, c, func(r ExternalDocRef, c copies) ExternalDocRef { return *r.clone(c) })
}

func (b *ServerBindings) clone() *ServerBindings {
	if b == nil {
		return nil
	}
	d := *b
	d.Kafka.configs = maps.Clone(b.Kafka.configs)
	return &d
}

func (b *ChannelBindings) clone() *ChannelBindings {
	if b == nil {
		return nil
	}
	d := *b
	d.Kafka.configs = maps.Clone(b.Kafka.configs)
	d.Mqtt.configs = maps.Clone(b.Mqtt.configs)
	d.Websocket.Query = b.Websocket.Query.Clone()
	d.Websocket.Headers = b.Websocket.Headers.Clone()
	return &d
}

func (b *OperationBindings) clone() *OperationBindings {
	if b == nil {
		return nil
	}
	d := *b
	d.Kafka.GroupId = b.Kafka.GroupId.Clone()
	d.Kafka.ClientId = b.Kafka.ClientId.Clone()
	d.Mqtt.configs = maps.Clone(b.Mqtt.configs)
	d.Websocket.configs = maps.Clone(b.Websocket.configs)
	return &d
}

func (b *MessageBinding) clone(c copies) *MessageBinding {
	if b == nil {
		return nil
	}
	d := *b
	d.Kafka.Key = b.Kafka.Key.clone(c)
	d.Mqtt.configs = maps.Clone(b.Mqtt.configs)
	d.Websocket.configs = maps.Clone(b.Websocket.configs)
	return &d
}

func (r *SchemaRef) clone(c copies) *SchemaRef {
	if r == nil {
		return nil
	}
	d := *r
	d.Value = cloneSchema(r.Value, c)
	return &d
}

func cloneSchema(s Schema, c copies) Schema {
	switch v := s.(type) {
	case *jsonSchema.Schema:
		return v.Clone()
	case *openapi.Schema:
		return v.Clone()
	case *AvroRef:
		// Avro schemas are not patched, so the copy shares the parsed Avro schema
		return clonePointer(v)
	case *MultiSchemaFormat:
		return v.clone(c)
	case *SchemaRef:
		return v.clone(c)
	default:
		return s
	}
}

func (m *MultiSchemaFormat) clone(c copies) *MultiSchemaFormat {
	if m == nil {
		return nil
	}
	d := *m
	d.Schema = m.Schema.clone(c)
	return &d
}

func (c *Components) clone(cp copies) *Components {
	if c == nil {
		return nil
	}
	return &Components{
		Servers:         cloneMap(c.Servers, cp, (*ServerRef).clone),
		Tags:            cloneMap(c.Tags, cp, (*TagRef).clone),
		Channels:        cloneMap(c.Channels, cp, (*ChannelRef).clone),
		Schemas:         cloneMap(c.Schemas, cp, (*SchemaRef).clone),
		Messages:        cloneMap(c.Messages, cp, (*MessageRef).clone),
		Operations:      cloneMap(c.Operations, cp, (*OperationRef).clone),
		Parameters:      cloneMap(c.Parameters, cp, (*ParameterRef).clone),
		CorrelationIds:  cloneMap(c.CorrelationIds, cp, (*CorrelationIdRef).clone),
		ExternalDocs:    cloneMap(c.ExternalDocs, cp, (*ExternalDocRef).clone),
		OperationTraits: cloneMap(c.OperationTraits, cp, (*OperationTraitRef).clone),
		MessageTraits:   cloneMap(c.MessageTraits, cp, (*MessageTraitRef).clone),
		ServerVariables: cloneMap(c.ServerVariables, cp, (*ServerVariableRef).clone),
	}
}

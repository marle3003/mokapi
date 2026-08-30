package asyncapi3

import (
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/sortedmap"
	"net/url"
	"path"
	"strings"
)

type Export struct {
}

func (e *Export) ToJSON(c *Config) ([]byte, error) {
	r := e.build(c)
	return json.Marshal(r)
}

func (e *Export) build(c *Config) *Config {
	r := &Config{
		Version:            c.Version,
		Id:                 c.Id,
		Info:               c.Info,
		DefaultContentType: c.DefaultContentType,
	}

	e.buildServers(c, r)
	e.buildChannels(c, r)
	e.buildOperations(c, r)

	return r
}

func (e *Export) buildServers(origin, result *Config) {
	if origin.Servers == nil || origin.Servers.Len() == 0 {
		return
	}

	result.Servers = &sortedmap.LinkedHashMap[string, *ServerRef]{}
	for it := origin.Servers.Iter(); it.Next(); {
		s := it.Value()
		if s.Value == nil {
			continue
		}
		result.Servers.Set(it.Key(), e.buildServer(s, result))
	}
}

func (e *Export) buildChannels(origin, result *Config) {
	if origin.Channels == nil || len(origin.Channels) == 0 {
		return
	}

	result.Channels = map[string]*ChannelRef{}
	for k, ch := range origin.Channels {
		if ch.Value == nil {
			continue
		}
		result.Channels[k] = e.buildChannel(ch, result)
	}
}

func (e *Export) buildOperations(origin, result *Config) {
	if origin.Operations == nil || len(origin.Operations) == 0 {
		return
	}

	result.Operations = map[string]*OperationRef{}
	for k, op := range origin.Operations {
		if op.Value == nil {
			continue
		}
		result.Operations[k] = e.buildOperation(op, result)
	}
}

func (e *Export) buildOperation(origin *OperationRef, cfg *Config) *OperationRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *OperationRef
	var value Operation
	if origin.HasRef() {
		name := refName(origin.Ref)
		e.ensureComponents(cfg)
		if cfg.Components.Operations == nil {
			cfg.Components.Operations = map[string]*OperationRef{}
		}

		ref := fmt.Sprintf("#/components/operations/%s", name)
		result = &OperationRef{
			Reference: dynamic.Reference[*OperationRef]{Ref: ref},
		}

		if _, ok := cfg.Components.Operations[name]; !ok {
			value = *origin.Value
			cfg.Components.Operations[name] = &OperationRef{Value: &value}
		} else {
			return result
		}
	} else {
		value = *origin.Value
		result = &OperationRef{Value: &value}
	}

	ch := e.buildChannel(origin.Value.Channel, cfg)
	if ch != nil {
		name := e.getChannel(ch, cfg)
		if name != "" {
			value.Channel = &ChannelRef{Reference: dynamic.Reference[*ChannelRef]{
				Ref: fmt.Sprintf("#/channels/%s", name),
			}}
		} else {
			value.Channel = ch
		}

	}

	value.Traits = nil
	for _, trait := range origin.Value.Traits {
		if trait.Value == nil {
			continue
		}
		value.Traits = append(value.Traits, e.buildOperationTrait(trait, cfg))
	}

	value.Messages = nil
	for _, msg := range origin.Value.Messages {
		if msg.Value == nil {
			continue
		}
		value.Messages = append(value.Messages, e.buildMessage(msg, cfg))
	}

	value.ExternalDocs = nil
	for _, doc := range origin.Value.ExternalDocs {
		if doc.Value == nil {
			continue
		}
		value.ExternalDocs = append(value.ExternalDocs, e.buildExternalDoc(doc, cfg))
	}

	return result
}

func (e *Export) buildServer(origin *ServerRef, cfg *Config) *ServerRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *ServerRef
	var value Server
	if origin.HasRef() {
		name := refName(origin.Ref)
		e.ensureComponents(cfg)
		if cfg.Components.Servers == nil {
			cfg.Components.Servers = map[string]*ServerRef{}
		}

		ref := fmt.Sprintf("#/components/servers/%s", name)
		result = &ServerRef{
			Reference: dynamic.Reference[*ServerRef]{Ref: ref},
		}

		if _, ok := cfg.Components.Servers[name]; !ok {
			value = *origin.Value
			cfg.Components.Servers[name] = &ServerRef{Value: &value}
		} else {
			return result
		}
	} else {
		value = *origin.Value
		result = &ServerRef{Value: &value}
	}

	value.Variables = nil
	for k, v := range origin.Value.Variables {
		if v.Value == nil {
			continue
		}
		value.Variables[k] = e.buildServerVariable(v, cfg)
	}

	value.Tags = nil
	for _, v := range origin.Value.Tags {
		if v.Value == nil {
			continue
		}
		value.Tags = append(value.Tags, e.buildTag(v, cfg))
	}

	return result
}

func (e *Export) buildChannel(origin *ChannelRef, cfg *Config) *ChannelRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *ChannelRef
	var value Channel
	if origin.HasRef() {
		name := refName(origin.Ref)
		e.ensureComponents(cfg)
		if cfg.Components.Channels == nil {
			cfg.Components.Channels = map[string]*ChannelRef{}
		}

		ref := fmt.Sprintf("#/components/channels/%s", name)
		result = &ChannelRef{
			Reference: dynamic.Reference[*ChannelRef]{Ref: ref},
		}

		if _, ok := cfg.Components.Channels[name]; !ok {
			value = *origin.Value
			cfg.Components.Channels[name] = &ChannelRef{Value: &value}
		} else {
			return result
		}
	} else {
		value = *origin.Value
		result = &ChannelRef{Value: &value}
	}

	value.Servers = nil
	for _, server := range origin.Value.Servers {
		if server.Value == nil {
			continue
		}
		value.Servers = append(value.Servers, &ServerRef{Reference: dynamic.Reference[*ServerRef]{Ref: server.Ref}})
	}
	value.Messages = map[string]*MessageRef{}
	for key, message := range origin.Value.Messages {
		if message.Value == nil {
			continue
		}
		value.Messages[key] = e.buildMessage(message, cfg)
	}

	value.Parameters = map[string]*ParameterRef{}
	for key, param := range origin.Value.Parameters {
		if param.Value == nil {
			continue
		}
		value.Parameters[key] = e.buildParameter(param, cfg)
	}

	value.Tags = nil
	for _, v := range origin.Value.Tags {
		if v.Value == nil {
			continue
		}
		value.Tags = append(value.Tags, e.buildTag(v, cfg))
	}

	value.ExternalDocs = nil
	for _, v := range origin.Value.ExternalDocs {
		if v.Value == nil {
			continue
		}
		value.ExternalDocs = append(value.ExternalDocs, e.buildExternalDoc(v, cfg))
	}

	return result
}

func (e *Export) buildMessage(origin *MessageRef, cfg *Config) *MessageRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *MessageRef
	var value Message
	if origin.HasRef() {
		name := refName(origin.Ref)
		e.ensureComponents(cfg)
		if cfg.Components.Messages == nil {
			cfg.Components.Messages = map[string]*MessageRef{}
		}

		ref := fmt.Sprintf("#/components/messages/%s", name)
		result = &MessageRef{
			Reference: dynamic.Reference[*MessageRef]{Ref: ref},
		}

		if _, ok := cfg.Components.Messages[name]; !ok {
			value = *origin.Value
			cfg.Components.Messages[name] = &MessageRef{Value: &value}
		} else {
			return result
		}
	} else {
		value = *origin.Value
		result = &MessageRef{Value: &value}
	}

	value.CorrelationId = e.buildCorrelationId(origin.Value.CorrelationId, cfg)

	value.Traits = nil
	for _, trait := range origin.Value.Traits {
		if trait.Value == nil {
			continue
		}
		value.Traits = append(value.Traits, e.buildMessageTrait(trait, cfg))
	}

	value.ExternalDocs = nil
	for _, ed := range origin.Value.ExternalDocs {
		if ed.Value == nil {
			continue
		}
		value.ExternalDocs = append(value.ExternalDocs, e.buildExternalDoc(ed, cfg))
	}

	if origin.Value.Payload != nil {
		result.Value.Payload = e.buildSchema(origin.Value.Payload, cfg)
	}

	return result
}

func (e *Export) buildCorrelationId(origin *CorrelationIdRef, cfg *Config) *CorrelationIdRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	if origin.HasRef() {
		name := refName(origin.Ref)
		e.ensureComponents(cfg)
		if cfg.Components.CorrelationIds == nil {
			cfg.Components.CorrelationIds = map[string]*CorrelationIdRef{}
		}
		if _, ok := cfg.Components.CorrelationIds[name]; !ok {
			cfg.Components.CorrelationIds[name] = &CorrelationIdRef{Value: origin.Value}
		}
		ref := fmt.Sprintf("#/components/correlationIds/%s", name)
		return &CorrelationIdRef{
			Reference: dynamic.Reference[*CorrelationIdRef]{Ref: ref},
		}
	}
	return origin
}

func (e *Export) buildMessageTrait(origin *MessageTraitRef, cfg *Config) *MessageTraitRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *MessageTraitRef
	var value MessageTrait
	if origin.HasRef() {
		name := refName(origin.Ref)
		e.ensureComponents(cfg)
		if cfg.Components.MessageTraits == nil {
			cfg.Components.MessageTraits = map[string]*MessageTraitRef{}
		}

		ref := fmt.Sprintf("#/components/messageTraits/%s", name)
		result = &MessageTraitRef{
			Reference: dynamic.Reference[*MessageTraitRef]{Ref: ref},
		}

		if _, ok := cfg.Components.Messages[name]; !ok {
			value = *origin.Value
			cfg.Components.MessageTraits[name] = &MessageTraitRef{Value: &value}
		} else {
			return result
		}
	} else {
		value = *origin.Value
		result = &MessageTraitRef{Value: &value}
	}

	value.CorrelationId = e.buildCorrelationId(origin.Value.CorrelationId, cfg)

	value.ExternalDocs = nil
	for _, ed := range value.ExternalDocs {
		if ed.Value == nil {
			continue
		}
		value.ExternalDocs = append(value.ExternalDocs, e.buildExternalDoc(ed, cfg))
	}

	return result
}

func (e *Export) buildOperationTrait(origin *OperationTraitRef, cfg *Config) *OperationTraitRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *OperationTraitRef
	var value OperationTrait
	if origin.HasRef() {
		name := refName(origin.Ref)
		e.ensureComponents(cfg)
		if cfg.Components.OperationTraits == nil {
			cfg.Components.OperationTraits = map[string]*OperationTraitRef{}
		}

		ref := fmt.Sprintf("#/components/operationTraits/%s", name)
		result = &OperationTraitRef{
			Reference: dynamic.Reference[*OperationTraitRef]{Ref: ref},
		}

		if _, ok := cfg.Components.OperationTraits[name]; !ok {
			value = *origin.Value
			cfg.Components.OperationTraits[name] = &OperationTraitRef{Value: &value}
		} else {
			return result
		}
	} else {
		value = *origin.Value
		result = &OperationTraitRef{Value: &value}
	}

	value.Channel = e.buildChannel(origin.Value.Channel, cfg)

	value.ExternalDocs = nil
	for _, doc := range origin.Value.ExternalDocs {
		if doc.Value == nil {
			continue
		}
		value.ExternalDocs = append(value.ExternalDocs, e.buildExternalDoc(doc, cfg))
	}
	return result
}

func (e *Export) buildExternalDoc(origin *ExternalDocRef, cfg *Config) *ExternalDocRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	if origin.HasRef() {
		name := refName(origin.Ref)
		e.ensureComponents(cfg)
		if cfg.Components.ExternalDocs == nil {
			cfg.Components.ExternalDocs = map[string]*ExternalDocRef{}
		}
		if _, ok := cfg.Components.ExternalDocs[name]; !ok {
			cfg.Components.ExternalDocs[name] = &ExternalDocRef{Value: origin.Value}
		}
		ref := fmt.Sprintf("#/components/externalDocs/%s", name)
		return &ExternalDocRef{
			Reference: dynamic.Reference[*ExternalDocRef]{Ref: ref},
		}
	}
	return origin
}

func (e *Export) buildSchema(origin *SchemaRef, cfg *Config) *SchemaRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	if origin.HasRef() {
		name := refName(origin.Ref)
		e.ensureComponents(cfg)
		if cfg.Components.Schemas == nil {
			cfg.Components.Schemas = map[string]*SchemaRef{}
		}
		if _, ok := cfg.Components.Schemas[name]; !ok {
			cfg.Components.Schemas[name] = &SchemaRef{Value: origin.Value}
		}
		ref := fmt.Sprintf("#/components/schemas/%s", name)
		return &SchemaRef{
			Reference: dynamic.Reference[*SchemaRef]{Ref: ref},
		}
	}
	return origin
}

func (e *Export) buildParameter(origin *ParameterRef, cfg *Config) *ParameterRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	if origin.HasRef() {
		name := refName(origin.Ref)
		e.ensureComponents(cfg)
		if cfg.Components.Parameters == nil {
			cfg.Components.Parameters = map[string]*ParameterRef{}
		}
		if _, ok := cfg.Components.Parameters[name]; !ok {
			cfg.Components.Parameters[name] = &ParameterRef{Value: origin.Value}
		}
		ref := fmt.Sprintf("#/components/parameters/%s", name)
		return &ParameterRef{
			Reference: dynamic.Reference[*ParameterRef]{Ref: ref},
		}
	}
	return origin
}

func (e *Export) buildServerVariable(origin *ServerVariableRef, cfg *Config) *ServerVariableRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	if origin.HasRef() {
		name := refName(origin.Ref)
		e.ensureComponents(cfg)
		if cfg.Components.ServerVariables == nil {
			cfg.Components.ServerVariables = map[string]*ServerVariableRef{}
		}
		if _, ok := cfg.Components.ServerVariables[name]; !ok {
			cfg.Components.ServerVariables[name] = &ServerVariableRef{Value: origin.Value}
		}
		ref := fmt.Sprintf("#/components/serverVariables/%s", name)
		return &ServerVariableRef{
			Reference: dynamic.Reference[*ServerVariableRef]{Ref: ref},
		}
	}
	return origin
}

func (e *Export) buildTag(origin *TagRef, cfg *Config) *TagRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	if origin.HasRef() {
		name := refName(origin.Ref)
		e.ensureComponents(cfg)
		if cfg.Components.Tags == nil {
			cfg.Components.Tags = map[string]*TagRef{}
		}
		if _, ok := cfg.Components.Tags[name]; !ok {
			cfg.Components.Tags[name] = &TagRef{Value: origin.Value}
		}
		ref := fmt.Sprintf("#/components/tags/%s", name)
		return &TagRef{
			Reference: dynamic.Reference[*TagRef]{Ref: ref},
		}
	}
	return origin
}

func (e *Export) ensureComponents(cfg *Config) {
	if cfg.Components == nil {
		cfg.Components = &Components{}
	}
}

func (e *Export) ensureComponentChannels(cfg *Config) {
	e.ensureComponents(cfg)
	if cfg.Components.Channels == nil {
		cfg.Components.Channels = map[string]*ChannelRef{}
	}
}

func (e *Export) getChannel(ch *ChannelRef, cfg *Config) string {
	for name, item := range cfg.Channels {
		if item.Ref == ch.Ref {
			return name
		}
	}
	return ""
}

func refName(ref string) string {
	if strings.HasPrefix(ref, "#/") {
		parts := strings.Split(ref, "/")
		return parts[len(parts)-1]
	}

	u, err := url.Parse(ref)
	if err != nil {
		return sanitize(ref)
	}

	file := strings.TrimSuffix(path.Base(u.Path), path.Ext(u.Path))
	if u.Path == "" || u.Path == "/" {
		file = ""
	}

	if u.Host != "" {
		if file != "" {
			file = u.Host + "_" + file
		} else {
			file = u.Host
		}
	}

	var pointer string
	if u.Fragment != "" {
		segs := strings.Split(strings.Trim(u.Fragment, "/"), "/")
		last := segs[len(segs)-1]
		last = strings.ReplaceAll(last, "~1", "/")
		last = strings.ReplaceAll(last, "~0", "~")
		pointer = last
	}

	switch {
	case file != "" && pointer != "":
		return sanitize(file + "_" + pointer)
	case pointer != "":
		return sanitize(pointer)
	default:
		return sanitize(file)
	}
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	name := b.String()
	if name == "" {
		return "_"
	}
	if name[0] >= '0' && name[0] <= '9' {
		name = "_" + name // avoid a leading digit
	}
	return name
}

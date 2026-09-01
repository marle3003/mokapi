package asyncapi3

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/sortedmap"
)

type export struct {
	result *Config
}

func (c *Config) Export() *Config {
	e := &export{
		result: &Config{
			Version:            c.Version,
			Id:                 c.Id,
			Info:               c.Info,
			DefaultContentType: c.DefaultContentType,
		},
	}

	e.buildServers(c)
	e.buildChannels(c)
	e.buildOperations(c)

	return e.result
}

func (e *export) buildServers(origin *Config) {
	if origin.Servers == nil || origin.Servers.Len() == 0 {
		return
	}

	e.result.Servers = &sortedmap.LinkedHashMap[string, *ServerRef]{}
	for it := origin.Servers.Iter(); it.Next(); {
		s := it.Value()
		if s.Value == nil {
			continue
		}
		e.result.Servers.Set(it.Key(), e.buildServer(s))
	}
}

func (e *export) buildChannels(origin *Config) {
	if origin.Channels == nil || len(origin.Channels) == 0 {
		return
	}

	e.result.Channels = map[string]*ChannelRef{}
	for k, ch := range origin.Channels {
		if ch.Value == nil {
			continue
		}
		e.result.Channels[k] = e.buildChannel(ch)
	}
}

func (e *export) buildOperations(origin *Config) {
	if origin.Operations == nil || len(origin.Operations) == 0 {
		return
	}

	e.result.Operations = map[string]*OperationRef{}
	for k, op := range origin.Operations {
		if op.Value == nil {
			continue
		}
		e.result.Operations[k] = e.buildOperation(op)
	}
}

func (e *export) buildOperation(origin *OperationRef) *OperationRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *OperationRef
	var value Operation
	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		e.ensureComponents()
		if e.result.Components.Operations == nil {
			e.result.Components.Operations = map[string]*OperationRef{}
		}

		ref := fmt.Sprintf("#/components/operations/%s", name)
		result = &OperationRef{
			Reference: dynamic.Reference[*OperationRef]{Ref: ref},
		}

		if _, ok := e.result.Components.Operations[name]; !ok {
			value = *origin.Value
			e.result.Components.Operations[name] = &OperationRef{Value: &value}
		} else {
			return result
		}
	} else {
		value = *origin.Value
		result = &OperationRef{Value: &value}
	}

	ch := e.buildChannel(origin.Value.Channel)
	if ch != nil {
		name := e.getChannel(ch)
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
		value.Traits = append(value.Traits, e.buildOperationTrait(trait))
	}

	value.Messages = nil
	for _, msg := range origin.Value.Messages {
		if msg.Value == nil {
			continue
		}
		value.Messages = append(value.Messages, e.buildMessage(msg))
	}

	value.ExternalDocs = nil
	for _, doc := range origin.Value.ExternalDocs {
		if doc.Value == nil {
			continue
		}
		value.ExternalDocs = append(value.ExternalDocs, e.buildExternalDoc(doc))
	}

	return result
}

func (e *export) buildServer(origin *ServerRef) *ServerRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *ServerRef
	var value Server
	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		e.ensureComponents()
		if e.result.Components.Servers == nil {
			e.result.Components.Servers = map[string]*ServerRef{}
		}

		ref := fmt.Sprintf("#/components/servers/%s", name)
		result = &ServerRef{
			Reference: dynamic.Reference[*ServerRef]{Ref: ref},
		}

		if _, ok := e.result.Components.Servers[name]; !ok {
			value = *origin.Value
			e.result.Components.Servers[name] = &ServerRef{Value: &value}
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
		value.Variables[k] = e.buildServerVariable(v)
	}

	value.Tags = nil
	for _, v := range origin.Value.Tags {
		if v.Value == nil {
			continue
		}
		value.Tags = append(value.Tags, e.buildTag(v))
	}

	return result
}

func (e *export) buildChannel(origin *ChannelRef) *ChannelRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *ChannelRef
	var value Channel
	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		e.ensureComponents()
		if e.result.Components.Channels == nil {
			e.result.Components.Channels = map[string]*ChannelRef{}
		}

		ref := fmt.Sprintf("#/components/channels/%s", name)
		result = &ChannelRef{
			Reference: dynamic.Reference[*ChannelRef]{Ref: ref},
		}

		if _, ok := e.result.Components.Channels[name]; !ok {
			value = *origin.Value
			e.result.Components.Channels[name] = &ChannelRef{Value: &value}
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
		value.Messages[key] = e.buildMessage(message)
	}

	value.Parameters = map[string]*ParameterRef{}
	for key, param := range origin.Value.Parameters {
		if param.Value == nil {
			continue
		}
		value.Parameters[key] = e.buildParameter(param)
	}

	value.Tags = nil
	for _, v := range origin.Value.Tags {
		if v.Value == nil {
			continue
		}
		value.Tags = append(value.Tags, e.buildTag(v))
	}

	value.ExternalDocs = nil
	for _, v := range origin.Value.ExternalDocs {
		if v.Value == nil {
			continue
		}
		value.ExternalDocs = append(value.ExternalDocs, e.buildExternalDoc(v))
	}

	return result
}

func (e *export) buildMessage(origin *MessageRef) *MessageRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *MessageRef
	var value Message
	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		e.ensureComponents()
		if e.result.Components.Messages == nil {
			e.result.Components.Messages = map[string]*MessageRef{}
		}

		ref := fmt.Sprintf("#/components/messages/%s", name)
		result = &MessageRef{
			Reference: dynamic.Reference[*MessageRef]{Ref: ref},
		}

		if _, ok := e.result.Components.Messages[name]; !ok {
			value = *origin.Value
			e.result.Components.Messages[name] = &MessageRef{Value: &value}
		} else {
			return result
		}
	} else {
		value = *origin.Value
		result = &MessageRef{Value: &value}
	}

	value.CorrelationId = e.buildCorrelationId(origin.Value.CorrelationId)

	value.Traits = nil
	for _, trait := range origin.Value.Traits {
		if trait.Value == nil {
			continue
		}
		value.Traits = append(value.Traits, e.buildMessageTrait(trait))
	}

	value.ExternalDocs = nil
	for _, ed := range origin.Value.ExternalDocs {
		if ed.Value == nil {
			continue
		}
		value.ExternalDocs = append(value.ExternalDocs, e.buildExternalDoc(ed))
	}

	if origin.Value.Payload != nil {
		result.Value.Payload = e.buildSchema(origin.Value.Payload)
	}

	return result
}

func (e *export) buildCorrelationId(origin *CorrelationIdRef) *CorrelationIdRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		e.ensureComponents()
		if e.result.Components.CorrelationIds == nil {
			e.result.Components.CorrelationIds = map[string]*CorrelationIdRef{}
		}
		if _, ok := e.result.Components.CorrelationIds[name]; !ok {
			e.result.Components.CorrelationIds[name] = &CorrelationIdRef{Value: origin.Value}
		}
		ref := fmt.Sprintf("#/components/correlationIds/%s", name)
		return &CorrelationIdRef{
			Reference: dynamic.Reference[*CorrelationIdRef]{Ref: ref},
		}
	}
	return origin
}

func (e *export) buildMessageTrait(origin *MessageTraitRef) *MessageTraitRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *MessageTraitRef
	var value MessageTrait
	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		e.ensureComponents()
		if e.result.Components.MessageTraits == nil {
			e.result.Components.MessageTraits = map[string]*MessageTraitRef{}
		}

		ref := fmt.Sprintf("#/components/messageTraits/%s", name)
		result = &MessageTraitRef{
			Reference: dynamic.Reference[*MessageTraitRef]{Ref: ref},
		}

		if _, ok := e.result.Components.Messages[name]; !ok {
			value = *origin.Value
			e.result.Components.MessageTraits[name] = &MessageTraitRef{Value: &value}
		} else {
			return result
		}
	} else {
		value = *origin.Value
		result = &MessageTraitRef{Value: &value}
	}

	value.CorrelationId = e.buildCorrelationId(origin.Value.CorrelationId)

	value.ExternalDocs = nil
	for _, ed := range value.ExternalDocs {
		if ed.Value == nil {
			continue
		}
		value.ExternalDocs = append(value.ExternalDocs, e.buildExternalDoc(ed))
	}

	return result
}

func (e *export) buildOperationTrait(origin *OperationTraitRef) *OperationTraitRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	var result *OperationTraitRef
	var value OperationTrait
	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		e.ensureComponents()
		if e.result.Components.OperationTraits == nil {
			e.result.Components.OperationTraits = map[string]*OperationTraitRef{}
		}

		ref := fmt.Sprintf("#/components/operationTraits/%s", name)
		result = &OperationTraitRef{
			Reference: dynamic.Reference[*OperationTraitRef]{Ref: ref},
		}

		if _, ok := e.result.Components.OperationTraits[name]; !ok {
			value = *origin.Value
			e.result.Components.OperationTraits[name] = &OperationTraitRef{Value: &value}
		} else {
			return result
		}
	} else {
		value = *origin.Value
		result = &OperationTraitRef{Value: &value}
	}

	value.Channel = e.buildChannel(origin.Value.Channel)

	value.ExternalDocs = nil
	for _, doc := range origin.Value.ExternalDocs {
		if doc.Value == nil {
			continue
		}
		value.ExternalDocs = append(value.ExternalDocs, e.buildExternalDoc(doc))
	}
	return result
}

func (e *export) buildExternalDoc(origin *ExternalDocRef) *ExternalDocRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		e.ensureComponents()
		if e.result.Components.ExternalDocs == nil {
			e.result.Components.ExternalDocs = map[string]*ExternalDocRef{}
		}
		if _, ok := e.result.Components.ExternalDocs[name]; !ok {
			e.result.Components.ExternalDocs[name] = &ExternalDocRef{Value: origin.Value}
		}
		ref := fmt.Sprintf("#/components/externalDocs/%s", name)
		return &ExternalDocRef{
			Reference: dynamic.Reference[*ExternalDocRef]{Ref: ref},
		}
	}
	return origin
}

func (e *export) buildSchema(origin *SchemaRef) *SchemaRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		e.ensureComponents()
		if e.result.Components.Schemas == nil {
			e.result.Components.Schemas = map[string]*SchemaRef{}
		}
		if _, ok := e.result.Components.Schemas[name]; !ok {
			e.result.Components.Schemas[name] = &SchemaRef{Value: origin.Value}
		}
		ref := fmt.Sprintf("#/components/schemas/%s", name)
		return &SchemaRef{
			Reference: dynamic.Reference[*SchemaRef]{Ref: ref},
		}
	}
	return origin
}

func (e *export) buildParameter(origin *ParameterRef) *ParameterRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		e.ensureComponents()
		if e.result.Components.Parameters == nil {
			e.result.Components.Parameters = map[string]*ParameterRef{}
		}
		if _, ok := e.result.Components.Parameters[name]; !ok {
			e.result.Components.Parameters[name] = &ParameterRef{Value: origin.Value}
		}
		ref := fmt.Sprintf("#/components/parameters/%s", name)
		return &ParameterRef{
			Reference: dynamic.Reference[*ParameterRef]{Ref: ref},
		}
	}
	return origin
}

func (e *export) buildServerVariable(origin *ServerVariableRef) *ServerVariableRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		e.ensureComponents()
		if e.result.Components.ServerVariables == nil {
			e.result.Components.ServerVariables = map[string]*ServerVariableRef{}
		}
		if _, ok := e.result.Components.ServerVariables[name]; !ok {
			e.result.Components.ServerVariables[name] = &ServerVariableRef{Value: origin.Value}
		}
		ref := fmt.Sprintf("#/components/serverVariables/%s", name)
		return &ServerVariableRef{
			Reference: dynamic.Reference[*ServerVariableRef]{Ref: ref},
		}
	}
	return origin
}

func (e *export) buildTag(origin *TagRef) *TagRef {
	if origin == nil || origin.Value == nil {
		return nil
	}

	if origin.HasRef() {
		name := dynamic.RefName(origin.Ref)
		e.ensureComponents()
		if e.result.Components.Tags == nil {
			e.result.Components.Tags = map[string]*TagRef{}
		}
		if _, ok := e.result.Components.Tags[name]; !ok {
			e.result.Components.Tags[name] = &TagRef{Value: origin.Value}
		}
		ref := fmt.Sprintf("#/components/tags/%s", name)
		return &TagRef{
			Reference: dynamic.Reference[*TagRef]{Ref: ref},
		}
	}
	return origin
}

func (e *export) ensureComponents() {
	if e.result.Components == nil {
		e.result.Components = &Components{}
	}
}

func (e *export) ensureComponentChannels() {
	e.ensureComponents()
	if e.result.Components.Channels == nil {
		e.result.Components.Channels = map[string]*ChannelRef{}
	}
}

func (e *export) getChannel(ch *ChannelRef) string {
	for name, item := range e.result.Channels {
		if item.Ref == ch.Ref {
			return name
		}
	}
	return ""
}

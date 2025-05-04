package asyncapi3test

import (
	"mokapi/providers/asyncapi3"
)

type OperationOptions func(o *asyncapi3.Operation)

func NewOperation(opts ...OperationOptions) *asyncapi3.Operation {
	op := &asyncapi3.Operation{}
	for _, opt := range opts {
		opt(op)
	}
	return op
}

func WithOperationAction(action string) OperationOptions {
	return func(op *asyncapi3.Operation) {
		op.Action = action
	}
}

func UseOperationMessage(msg *asyncapi3.Message) OperationOptions {
	return func(o *asyncapi3.Operation) {
		o.Messages = append(o.Messages, &asyncapi3.MessageRef{Value: msg})
	}
}

func WithOperationChannel(ch *asyncapi3.Channel) OperationOptions {
	return func(o *asyncapi3.Operation) {
		o.Channel = asyncapi3.ChannelRef{Value: ch}
	}
}

func WithOperationBinding(b asyncapi3.KafkaOperationBinding) OperationOptions {
	return func(o *asyncapi3.Operation) {
		o.Bindings.Kafka = b
	}
}

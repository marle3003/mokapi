package event

import (
	"fmt"
	"mokapi/config/dynamic/asyncApi"
	"strings"
)

func buildMessageFromComponents(name string, config *asyncApi.Message, ctx *context) (msg *Message) {
	if m, exists := ctx.service.Messages[name]; exists && m.isResolved {
		return m
	} else if exists {
		msg = m
		m = newMessage(config)
		*msg = *m
		msg.Name = name
	} else {
		msg = newMessage(config)
		msg.Name = name
		ctx.service.Messages[name] = msg
	}

	msg.Reference = "#/components/messages/" + name
	msg.Payload = createSchema(config.Payload, ctx)

	return
}

func createMessage(config *asyncApi.Message, ctx *context) *Message {
	if config == nil {
		return nil
	}

	if len(config.Reference) > 0 {
		switch r := config.Reference; {
		case strings.HasPrefix(r, "#/components/messages"):
			seg := strings.Split(r, "/")
			name := seg[len(seg)-1]
			if m, exists := ctx.service.Messages[name]; exists {
				return m
			} else {
				m := &Message{Reference: r, isResolved: false}
				ctx.service.Messages[name] = m
				return m
			}
		default:
			ctx.error(fmt.Sprintf("$ref '%v' is not supported", r))
			return nil
		}
	}

	msg := newMessage(config)
	msg.Payload = createSchema(config.Payload, ctx)

	return msg
}

func newMessage(config *asyncApi.Message) *Message {
	msg := &Message{
		Description: config.Description,
		Reference:   config.Reference,
	}

	return msg
}

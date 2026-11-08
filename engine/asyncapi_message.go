package engine

import (
	"encoding/json"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
)

// selectMessage returns the message the script selected by messageId. Without a messageId it
// returns the choice of the protocol (pick). An unknown messageId is an error, never a fallback.
func selectMessage(ch *asyncapi3.Channel, messageId string, pick func() (*asyncapi3.Message, error)) (*asyncapi3.Message, error) {
	if messageId != "" {
		return ch.GetMessage(messageId)
	}
	return pick()
}

func payloadSchema(msg *asyncapi3.Message) *asyncapi3.SchemaRef {
	if msg == nil {
		return nil
	}
	return msg.Payload
}

// generatePayload creates a payload from the payload schema of a selected message.
// A message without payload schema has no payload (nil).
func generatePayload(msg *asyncapi3.Message) (any, error) {
	schema := payloadSchema(msg)
	if schema == nil {
		return nil, nil
	}
	return createValue(schema)
}

// generateAnyPayload creates a payload for a message that was not selected by messageId.
// Unlike generatePayload, a message without payload schema generates any value.
func generateAnyPayload(msg *asyncapi3.Message) (any, error) {
	return createValue(payloadSchema(msg))
}

// encodePayload validates data against the payload schema of the message and encodes it
// with the content type of the message. Without a message or payload schema, data is
// encoded as JSON.
func encodePayload(msg *asyncapi3.Message, data any) ([]byte, error) {
	schema := payloadSchema(msg)
	if schema == nil {
		return json.Marshal(data)
	}
	return schema.Marshal(data, media.ParseContentType(msg.ContentType))
}

// anyMessage returns a message of the channel in map order (not uniformly random), or nil if it has none.
func anyMessage(ch *asyncapi3.Channel) *asyncapi3.Message {
	for _, ref := range ch.Messages {
		if ref != nil && ref.Value != nil {
			return ref.Value
		}
	}
	return nil
}

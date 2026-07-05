package monitor

import (
	"context"
	"mokapi/runtime/metrics"
)

var websocket = contextKey("websocket")

type Websocket struct {
	Messages    *metrics.CounterMap
	LastMessage *metrics.GaugeMap
}

func NewWebsocket() *Websocket {
	messages := metrics.NewCounterMap(
		metrics.WithFQName("websocket", "messages_total"),
		metrics.WithLabelNames("service", "topic"))
	lastMessage := metrics.NewGaugeMap(
		metrics.WithFQName("websocket", "message_timestamp"),
		metrics.WithLabelNames("service", "topic"))

	return &Websocket{
		Messages:    messages,
		LastMessage: lastMessage,
	}
}

func (k *Websocket) Metrics() []metrics.Metric {
	return []metrics.Metric{k.Messages, k.LastMessage}
}

func (k *Websocket) Reset() {
	k.Messages.Reset()
	k.LastMessage.Reset()
}

func NewWebsocketContext(ctx context.Context, ws *Websocket) context.Context {
	return context.WithValue(ctx, websocket, ws)
}

func WebsocketFromContext(ctx context.Context) (*Websocket, bool) {
	m, ok := ctx.Value(websocket).(*Websocket)
	return m, ok
}

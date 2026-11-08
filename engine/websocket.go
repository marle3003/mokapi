package engine

import (
	"errors"
	"fmt"
	"maps"
	"mokapi/engine/common"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime"
	"slices"
	"time"

	log "github.com/sirupsen/logrus"
)

type WebsocketEventHandler = eventHandler[common.WebsocketFilter]
type WebsocketEventDispatcher struct {
	dispatcher[common.WebsocketFilter]
}

func (e *WebsocketEventDispatcher) EmitWebsocketConnect(event *common.WebsocketConnectEvent) []*common.Action {
	return e.emit(emitOptions[common.WebsocketFilter]{
		shouldEmit: func(f common.WebsocketFilter) bool {
			return shouldWebsocketEmit(f, event.WebsocketEvent)
		},
	}, event)
}

func (e *WebsocketEventDispatcher) EmitWebsocketClose(event *common.WebsocketCloseEvent) []*common.Action {
	return e.emit(emitOptions[common.WebsocketFilter]{
		shouldEmit: func(f common.WebsocketFilter) bool {
			return shouldWebsocketEmit(f, event.WebsocketEvent)
		},
	}, event)
}

func (e *WebsocketEventDispatcher) EmitWebsocketMessage(event *common.WebsocketMessageEvent) []*common.Action {
	return e.emit(emitOptions[common.WebsocketFilter]{
		shouldEmit: func(f common.WebsocketFilter) bool {
			return shouldWebsocketEmit(f, event.WebsocketEvent)
		},
	}, event)
}

func (sh *scriptHost) OnWebsocket(filter common.WebsocketFilter, do common.EventHandlerFunc, args common.EventArgs) {
	sh.engine.WebsocketEventDispatcher.Register(sh.name, newEventHandler(sh, "websocket", filter, do, args))
}

func shouldWebsocketEmit(filter common.WebsocketFilter, event common.WebsocketEvent) bool {
	if !matchesIfPresent(filter.Type, event.Type) {
		return false
	}
	if !matchesIfPresent(filter.Api, event.Api) {
		return false
	}
	key := ""
	if event.Channel != nil {
		key = event.Channel.Key
	}
	if !matchesIfPresent(filter.Channel, key) {
		return false
	}
	return true
}

type WebsocketClient struct {
	app *runtime.App
}

func NewWebsocketClient(app *runtime.App) *WebsocketClient {
	return &WebsocketClient{
		app: app,
	}
}

func (c *WebsocketClient) Send(args common.WebsocketSendArgs) (*common.WebsocketSendResult, error) {
	w, cfg, err := c.tryGet(args.Api, args.ChannelName, args.Retry)
	if err != nil {
		return nil, err
	}
	if args.ChannelName == "" {
		args.ChannelName = cfg.ResolveAddress()
	}
	ch, ok := w.Store.Channel(args.ChannelName)
	if !ok {
		return nil, fmt.Errorf("websocket channel '%v' not found", args.ChannelName)
	}

	selected, err := selectMessage(cfg, args.MessageId, func() (*asyncapi3.Message, error) {
		return anyMessage(cfg), nil
	})
	if err != nil {
		return nil, err
	}
	msg, err := websocketMessage(selected, args)
	if err != nil {
		return nil, err
	}

	err = ch.Broadcast(msg)
	if err != nil {
		return nil, err
	}

	return &common.WebsocketSendResult{
		Api:     w.Info.Name,
		Channel: ch.Name,
		Message: msg,
	}, nil
}

// websocketMessage returns the message to send. A given message is validated against the message
// selected by messageId, otherwise it is sent as is. Without a message, a payload is generated.
func websocketMessage(selected *asyncapi3.Message, args common.WebsocketSendArgs) (any, error) {
	if args.Message != nil {
		if args.MessageId != "" {
			if _, err := encodePayload(selected, args.Message); err != nil {
				return nil, err
			}
		}
		return args.Message, nil
	}
	if selected == nil {
		return nil, nil
	}

	generate := generateAnyPayload
	if args.MessageId != "" {
		generate = generatePayload
	}
	msg, err := generate(selected)
	if err != nil {
		return nil, fmt.Errorf("unable to generate websocket message: %v", err)
	}
	return msg, nil
}

func (c *WebsocketClient) tryGet(api string, channelName string, retry common.RetryArgs) (w *runtime.WebsocketInfo, cfg *asyncapi3.Channel, err error) {
	count := 0
	backoff := retry.InitialRetryTime
	for {
		w, cfg, err = c.get(api, channelName)
		ambiguous := &ambiguousError{}
		if err == nil || errors.As(err, &ambiguous) {
			return
		}
		count++
		if count >= retry.Retries || backoff > retry.MaxRetryTime {
			return
		}
		log.Debugf("websocket channel '%v' not found. Retry in %v", channelName, backoff)
		time.Sleep(backoff)
		backoff *= time.Duration(retry.Factor)
	}
}

func (c *WebsocketClient) get(api string, channelName string) (k *runtime.WebsocketInfo, cfg *asyncapi3.Channel, err error) {
	if len(api) == 0 {
		if len(channelName) == 0 {
			apis := c.app.Websocket.List()
			if len(apis) > 1 {
				err = newAmbiguousError("ambiguous websocket API: specify the cluster")
				return
			}
			channels := slices.Collect(maps.Values(apis[0].Store.Channels))
			if len(channels) > 1 {
				err = newAmbiguousError("ambiguous channel %v. Specify the API", channelName)
				return
			}
			return apis[0], channels[0].Config, nil
		}

		var topics []*asyncapi3.Channel
		for _, v := range c.app.Websocket.List() {
			if t, ok := v.Channel(channelName); ok {
				k = v
				if len(api) == 0 {
					api = v.Info.Name
				}
				topics = append(topics, t.Config)
			} else if channelName != "" {
				for _, ref := range v.Config.Channels {
					if ref.Value == nil {
						continue
					}
					if err = ref.Value.IsNameValid(channelName); err == nil {
						k = v
						topics = append(topics, ref.Value)
					}
				}
			}
		}
		if len(topics) > 1 {
			err = newAmbiguousError("ambiguous channel %v. Specify the cluster", channelName)
			return
		} else if len(topics) == 1 {
			cfg = topics[0]
		}
	} else {
		if k = c.app.Websocket.Get(api); k != nil {
			ch, ok := k.Store.Channel(channelName)
			if ok {
				cfg = ch.Config
			}
		} else {
			return nil, nil, fmt.Errorf("websocket API '%v' not found", api)
		}
	}

	if cfg == nil {
		err = fmt.Errorf("websocket channel '%v' not found", channelName)
		return
	}

	return
}

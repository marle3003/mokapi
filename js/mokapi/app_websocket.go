package mokapi

import (
	"fmt"
	"mokapi/engine/common"
	"mokapi/js/util"

	"github.com/dop251/goja"
)

type Websocket struct {
	filter common.WebsocketFilter
	m      *Module
}

type WebsocketChannel struct {
	filter common.WebsocketFilter
	m      *Module
}

func (w *Websocket) Api(name string) *Websocket {
	f := w.filter
	f.Api = name
	return &Websocket{filter: f, m: w.m}
}

func (w *Websocket) Channel(channelName string) *WebsocketChannel {
	f := w.filter
	f.Channel = channelName
	return &WebsocketChannel{filter: f, m: w.m}
}

func (w *Websocket) Connect(do goja.Value, vArgs goja.Value) *Websocket {
	args, err := getOnArgs(w.m.vm, vArgs)
	if err != nil {
		panic(err)
	}
	h := getHandler(do, args, w.m.vm, w.m.loop)
	f := w.filter
	f.Type = common.WebsocketConnectEventType

	w.m.host.OnWebsocket(f, h, common.EventArgs{Tags: args.tags, Priority: args.priority})
	return w
}

func (w *Websocket) Message(do goja.Value, vArgs goja.Value) *Websocket {
	args, err := getOnArgs(w.m.vm, vArgs)
	if err != nil {
		panic(err)
	}
	h := getHandler(do, args, w.m.vm, w.m.loop)
	f := w.filter
	f.Type = common.WebsocketMessageEventType

	w.m.host.OnWebsocket(f, h, common.EventArgs{Tags: args.tags, Priority: args.priority})
	return w
}

func (w *Websocket) Close(do goja.Value, vArgs goja.Value) *Websocket {
	args, err := getOnArgs(w.m.vm, vArgs)
	if err != nil {
		panic(err)
	}
	h := getHandler(do, args, w.m.vm, w.m.loop)
	f := w.filter
	f.Type = common.WebsocketCloseEventType

	w.m.host.OnWebsocket(f, h, common.EventArgs{Tags: args.tags, Priority: args.priority})
	return w
}

func (w *WebsocketChannel) Connect(do goja.Value, vArgs goja.Value) *WebsocketChannel {
	args, err := getOnArgs(w.m.vm, vArgs)
	if err != nil {
		panic(err)
	}
	h := getHandler(do, args, w.m.vm, w.m.loop)
	f := w.filter
	f.Type = common.WebsocketConnectEventType

	w.m.host.OnWebsocket(f, h, common.EventArgs{Tags: args.tags, Priority: args.priority})
	return w
}

func (w *WebsocketChannel) Message(do goja.Value, vArgs goja.Value) *WebsocketChannel {
	args, err := getOnArgs(w.m.vm, vArgs)
	if err != nil {
		panic(err)
	}
	h := getHandler(do, args, w.m.vm, w.m.loop)
	f := w.filter
	f.Type = common.WebsocketMessageEventType

	w.m.host.OnWebsocket(f, h, common.EventArgs{Tags: args.tags, Priority: args.priority})
	return w
}

func (w *WebsocketChannel) Close(do goja.Value, vArgs goja.Value) *WebsocketChannel {
	args, err := getOnArgs(w.m.vm, vArgs)
	if err != nil {
		panic(err)
	}
	h := getHandler(do, args, w.m.vm, w.m.loop)
	f := w.filter
	f.Type = common.WebsocketCloseEventType

	w.m.host.OnWebsocket(f, h, common.EventArgs{Tags: args.tags, Priority: args.priority})
	return w
}

func (w *Websocket) Send(channelName string, vMsg goja.Value, vArgs goja.Value) *Websocket {
	f := w.filter
	f.Channel = channelName
	sendWebsocket(w.m, f, vMsg, vArgs)
	return w
}

func (w *Websocket) SendAsync(channelName string, vMsg goja.Value, vArgs goja.Value) *goja.Promise {
	f := w.filter
	f.Channel = channelName
	return sendWebsocketAsync(w.m, f, vMsg, vArgs, w)
}

func (w *WebsocketChannel) Send(vMsg goja.Value, vArgs goja.Value) *WebsocketChannel {
	sendWebsocket(w.m, w.filter, vMsg, vArgs)
	return w
}

func (w *WebsocketChannel) SendAsync(vMsg goja.Value, vArgs goja.Value) *goja.Promise {
	return sendWebsocketAsync(w.m, w.filter, vMsg, vArgs, w)
}

func sendWebsocket(m *Module, filter common.WebsocketFilter, vMsg goja.Value, vArgs goja.Value) {
	args, callback := prepareWebsocketSend(m.vm, filter, vMsg, vArgs)
	result, err := m.host.WebsocketClient().Send(args)
	if err != nil {
		panic(m.vm.ToValue(err.Error()))
	}
	if callback != nil {
		if err := callback(result); err != nil {
			panic(m.vm.ToValue(err.Error()))
		}
	}
}

// sendWebsocketAsync resolves the promise with self, like the synchronous variant returns it.
func sendWebsocketAsync[T any](m *Module, filter common.WebsocketFilter, vMsg goja.Value, vArgs goja.Value, self T) *goja.Promise {
	args, callback := prepareWebsocketSend(m.vm, filter, vMsg, vArgs)
	client := m.host.WebsocketClient()
	return util.RunAsync(m.vm, m.loop,
		func() (*common.WebsocketSendResult, error) { return client.Send(args) },
		func(result *common.WebsocketSendResult) (any, error) {
			if callback != nil {
				if err := callback(result); err != nil {
					return nil, err
				}
			}
			return self, nil
		})
}

// prepareWebsocketSend converts message and arguments to plain Go values. It has to run on
// the event loop because it reads from the goja runtime. Invalid input throws.
func prepareWebsocketSend(vm *goja.Runtime, filter common.WebsocketFilter, vMsg goja.Value, vArgs goja.Value) (common.WebsocketSendArgs, WebsocketSendResultCallback) {
	defer util.ThrowPanic(vm)

	args, err := mapSendArgs(vArgs, vm)
	if err != nil {
		panic(vm.ToValue(err.Error()))
	}

	return common.WebsocketSendArgs{
		Api:         filter.Api,
		ChannelName: filter.Channel,
		MessageId:   args.MessageId,
		Message:     vMsg.Export(),
		ScriptFile:  util.GetScriptFile(vm).Info.Key(),
		Retry:       args.Retry,
	}, args.result
}

type WebsocketSendArgs struct {
	MessageId string
	Retry     common.RetryArgs
	result    WebsocketSendResultCallback
}

type WebsocketSendResultCallback func(result *common.WebsocketSendResult) error

func mapSendArgs(v goja.Value, vm *goja.Runtime) (*WebsocketSendArgs, error) {
	args := &WebsocketSendArgs{
		Retry: util.DefaultRetryArgs(),
	}

	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return args, nil
	}

	jsObj := v.ToObject(vm)
	for _, propName := range jsObj.Keys() {
		switch propName {
		case "messageId":
			var err error
			args.MessageId, err = util.MessageId(jsObj.Get(propName).Export())
			if err != nil {
				return nil, err
			}
		case "retry":
			retry, ok := jsObj.Get(propName).Export().(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid property type '%s'", propName)
			}
			var err error
			args.Retry, err = util.ConvertToRetryArgs(retry)
			if err != nil {
				return nil, err
			}
		case "result":
			call, ok := goja.AssertFunction(jsObj.Get(propName))
			if !ok {
				return nil, fmt.Errorf("unexpected type for 'send': expected function, got %s", util.JsType(jsObj.Get(propName)))
			}
			args.result = func(result *common.WebsocketSendResult) error {
				_, err := call(goja.Undefined(), vm.ToValue(result))
				return err
			}
		}
	}

	return args, nil
}

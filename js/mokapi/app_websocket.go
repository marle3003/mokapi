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
	callback, result := sendWebsocket(f, vMsg, vArgs, w.m.host.WebsocketClient(), w.m.vm)
	if callback != nil {
		if err := callback(result); err != nil {
			panic(w.m.vm.ToValue(err.Error()))
		}
	}
	return w
}

func (w *Websocket) SendAsync(channelName string, vMsg goja.Value, vArgs goja.Value) *goja.Promise {
	p, resolve, reject := w.m.vm.NewPromise()
	f := w.filter
	f.Channel = channelName
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				w.m.loop.Run(func(vm *goja.Runtime) {
					_ = reject(r)
				})
			}
		}()

		sendWebsocket(w.filter, vMsg, vArgs, w.m.host.WebsocketClient(), w.m.vm)
		callback, result := sendWebsocket(f, vMsg, vArgs, w.m.host.WebsocketClient(), w.m.vm)
		w.m.loop.Run(func(vm *goja.Runtime) {
			if callback != nil {
				if err := callback(result); err != nil {
					panic(vm.ToValue(err.Error()))
				}
			}
			_ = resolve(w)
		})
	}()
	return p
}

func (w *WebsocketChannel) Send(vMsg goja.Value, vArgs goja.Value) *WebsocketChannel {
	callback, result := sendWebsocket(w.filter, vMsg, vArgs, w.m.host.WebsocketClient(), w.m.vm)
	if callback != nil {
		if err := callback(result); err != nil {
			panic(w.m.vm.ToValue(err.Error()))
		}
	}
	return w
}

func (w *WebsocketChannel) SendAsync(vMsg goja.Value, vArgs goja.Value) *goja.Promise {
	p, resolve, reject := w.m.vm.NewPromise()
	f := w.filter
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				w.m.loop.Run(func(vm *goja.Runtime) {
					_ = reject(r)
				})
			}
		}()

		sendWebsocket(w.filter, vMsg, vArgs, w.m.host.WebsocketClient(), w.m.vm)
		callback, result := sendWebsocket(f, vMsg, vArgs, w.m.host.WebsocketClient(), w.m.vm)
		w.m.loop.Run(func(vm *goja.Runtime) {
			if callback != nil {
				if err := callback(result); err != nil {
					panic(vm.ToValue(err.Error()))
				}
			}
			_ = resolve(w)
		})
	}()
	return p
}

func sendWebsocket(filter common.WebsocketFilter, vMsg goja.Value, vArgs goja.Value, client common.WebsocketClient, vm *goja.Runtime) (WebsocketSendResultCallback, *common.WebsocketSendResult) {
	defer func() {
		r := recover()
		if r != nil {
			panic(vm.ToValue(fmt.Sprintf("%v", r)))
		}
	}()

	msg := vMsg.Export()
	args, err := mapSendArgs(vArgs, vm)
	if err != nil {
		panic(vm.ToValue(err.Error()))
	}

	file := util.GetScriptFile(vm)
	result, err := client.Send(common.WebsocketSendArgs{
		Api:         filter.Api,
		ChannelName: filter.Channel,
		Message:     msg,
		ScriptFile:  file.Info.Key(),
		Retry:       args.Retry,
	})
	if err != nil {
		panic(vm.ToValue(err.Error()))
	}
	return args.result, result
}

type WebsocketSendArgs struct {
	Retry  common.RetryArgs
	result WebsocketSendResultCallback
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
				return nil, fmt.Errorf("unexpected type for 'fake': expected function, got %s", util.JsType(jsObj.Get(propName)))
			}
			args.result = func(result *common.WebsocketSendResult) error {
				_, err := call(goja.Undefined(), vm.ToValue(result))
				return err
			}
		}
	}

	return args, nil
}

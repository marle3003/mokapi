package mokapi_test

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/mokapi"
	"mokapi/js/require"
	"sync"
	"testing"
	"time"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

func TestModule_AppWebsocket_SendAsync(t *testing.T) {
	testcases := []struct {
		name            string
		script          string
		expectedChannel string
	}{
		{
			name: "sendAsync on channel sends the message once",
			script: `
const m = require('mokapi')
const p = m.app.websocket().channel('/foo').sendAsync({ text: 'hello' })
`,
			expectedChannel: "/foo",
		},
		{
			name: "sendAsync with channel name sends the message once to that channel",
			script: `
const m = require('mokapi')
const p = m.app.websocket().sendAsync('/foo', { text: 'hello' })
`,
			expectedChannel: "/foo",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			reg, err := require.NewRegistry(nil)
			reg.RegisterNativeModule("mokapi", mokapi.Require)
			r.NoError(t, err)

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			loop := eventloop.New(vm, host)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			var mu sync.Mutex
			var sends []common.WebsocketSendArgs
			host.WebsocketClientTest = &enginetest.WebsocketClient{
				SendFunc: func(args common.WebsocketSendArgs) (*common.WebsocketSendResult, error) {
					mu.Lock()
					defer mu.Unlock()
					sends = append(sends, args)
					return nil, nil
				},
			}

			_, err = vm.RunString(tc.script)
			r.NoError(t, err)

			p := vm.Get("p").Export().(*goja.Promise)
			counter := 0
			for p.State() == goja.PromiseStatePending && counter < 10 {
				time.Sleep(100 * time.Millisecond)
				counter++
			}

			mu.Lock()
			defer mu.Unlock()
			r.Len(t, sends, 1, "one sendAsync call must send exactly one message")
			r.Equal(t, tc.expectedChannel, sends[0].ChannelName)
		})
	}
}

func TestModule_AppWebsocket_SendAsync_SeveralChannels(t *testing.T) {
	reg, err := require.NewRegistry(nil)
	reg.RegisterNativeModule("mokapi", mokapi.Require)
	r.NoError(t, err)

	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	host := &enginetest.Host{}
	loop := eventloop.New(vm, host)
	defer loop.Stop()
	loop.StartLoop()
	js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
	reg.Enable(vm)

	var mu sync.Mutex
	var sends []common.WebsocketSendArgs
	host.WebsocketClientTest = &enginetest.WebsocketClient{
		SendFunc: func(args common.WebsocketSendArgs) (*common.WebsocketSendResult, error) {
			// a send without channel name is ambiguous when the API has several channels
			if args.ChannelName == "" {
				return nil, fmt.Errorf("ambiguous channel")
			}
			mu.Lock()
			defer mu.Unlock()
			sends = append(sends, args)
			return nil, nil
		},
	}

	_, err = vm.RunString(`
const m = require('mokapi')
const p = m.app.websocket().sendAsync('/foo', { text: 'hello' })
`)
	r.NoError(t, err)

	p := vm.Get("p").Export().(*goja.Promise)
	counter := 0
	for p.State() == goja.PromiseStatePending && counter < 10 {
		time.Sleep(100 * time.Millisecond)
		counter++
	}

	mu.Lock()
	defer mu.Unlock()
	r.Equal(t, goja.PromiseStateFulfilled, p.State())
	r.Len(t, sends, 1)
	r.Equal(t, "/foo", sends[0].ChannelName)
}

func TestModule_AppWebsocket_Send(t *testing.T) {
	testcases := []struct {
		name   string
		script string
	}{
		{
			name: "send on router with channel name",
			script: `
const m = require('mokapi')
m.app.websocket().send('/foo', { text: 'hello' })
`,
		},
		{
			name: "send on channel",
			script: `
const m = require('mokapi')
m.app.websocket().channel('/foo').send({ text: 'hello' })
`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			reg, err := require.NewRegistry(nil)
			reg.RegisterNativeModule("mokapi", mokapi.Require)
			r.NoError(t, err)

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			loop := eventloop.New(vm, host)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			var sends []common.WebsocketSendArgs
			host.WebsocketClientTest = &enginetest.WebsocketClient{
				SendFunc: func(args common.WebsocketSendArgs) (*common.WebsocketSendResult, error) {
					sends = append(sends, args)
					return nil, nil
				},
			}

			_, err = vm.RunString(tc.script)
			r.NoError(t, err)

			r.Len(t, sends, 1)
			r.Equal(t, "/foo", sends[0].ChannelName)
		})
	}
}

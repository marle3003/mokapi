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

			_, err = evalOnLoop(loop, tc.script)
			r.NoError(t, err)

			awaitPromise(t, loop, "p")

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

	_, err = evalOnLoop(loop, `
const m = require('mokapi')
const p = m.app.websocket().sendAsync('/foo', { text: 'hello' })
`)
	r.NoError(t, err)

	state := awaitPromise(t, loop, "p")

	mu.Lock()
	defer mu.Unlock()
	r.Equal(t, goja.PromiseStateFulfilled, state)
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

			_, err = evalOnLoop(loop, tc.script)
			r.NoError(t, err)

			r.Len(t, sends, 1)
			r.Equal(t, "/foo", sends[0].ChannelName)
		})
	}
}

func TestModule_AppWebsocket_SendAsync_Errors(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		client func(args common.WebsocketSendArgs) (*common.WebsocketSendResult, error)
		reason string // part of the rejection reason of the promise
		thrown string // part of the exception when the call itself must throw
	}{
		{
			name:   "invalid retry throws synchronously",
			script: `m.app.websocket().sendAsync('/foo', { text: 'hello' }, { retry: 'foo' })`,
			thrown: "invalid property type 'retry'",
		},
		{
			name:   "invalid result callback throws synchronously",
			script: `m.app.websocket().channel('/foo').sendAsync({ text: 'hello' }, { result: 'foo' })`,
			thrown: "unexpected type for 'send': expected function, got String",
		},
		{
			name:   "client error rejects the promise",
			script: `m.app.websocket().sendAsync('/foo', { text: 'hello' })`,
			client: func(common.WebsocketSendArgs) (*common.WebsocketSendResult, error) {
				return nil, fmt.Errorf("TEST ERROR")
			},
			reason: "TEST ERROR",
		},
		{
			name:   "result callback error rejects the promise",
			script: `m.app.websocket().sendAsync('/foo', { text: 'hello' }, { result: () => { throw new Error('CALLBACK ERROR') } })`,
			reason: "CALLBACK ERROR",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			loop, host := newModuleLoop(t)
			host.WebsocketClientTest = &enginetest.WebsocketClient{
				SendFunc: func(args common.WebsocketSendArgs) (*common.WebsocketSendResult, error) {
					if tc.client != nil {
						return tc.client(args)
					}
					return &common.WebsocketSendResult{}, nil
				},
			}

			_, err := evalOnLoop(loop, `const m = require('mokapi')`)
			r.NoError(t, err)
			_, err = evalOnLoop(loop, `const p = `+tc.script)
			if tc.thrown != "" {
				r.ErrorContains(t, err, tc.thrown)
				return
			}
			r.NoError(t, err)

			state, reason := awaitPromiseResult(t, loop, "p")
			r.Equal(t, goja.PromiseStateRejected, state)
			r.Contains(t, reason, tc.reason)
		})
	}
}

func TestModule_AppWebsocket_SendAsync_SendsMessageAsItWasAtCallTime(t *testing.T) {
	loop, host := newModuleLoop(t)
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

	_, err := evalOnLoop(loop, `
const m = require('mokapi')
const msg = { text: 'hello' }
const p = m.app.websocket().sendAsync('/foo', msg)
msg.text = 'changed'
`)
	r.NoError(t, err)
	r.Equal(t, goja.PromiseStateFulfilled, awaitPromise(t, loop, "p"))

	mu.Lock()
	defer mu.Unlock()
	r.Len(t, sends, 1)
	r.Equal(t, map[string]any{"text": "hello"}, sends[0].Message)
}

func TestModule_AppWebsocket_Send_Errors(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		client func(args common.WebsocketSendArgs) (*common.WebsocketSendResult, error)
		thrown string
	}{
		{
			name:   "client error throws",
			script: `m.app.websocket().send('/foo', { text: 'hello' })`,
			client: func(common.WebsocketSendArgs) (*common.WebsocketSendResult, error) {
				return nil, fmt.Errorf("TEST ERROR")
			},
			thrown: "TEST ERROR",
		},
		{
			name:   "channel: client error throws",
			script: `m.app.websocket().channel('/foo').send({ text: 'hello' })`,
			client: func(common.WebsocketSendArgs) (*common.WebsocketSendResult, error) {
				return nil, fmt.Errorf("TEST ERROR")
			},
			thrown: "TEST ERROR",
		},
		{
			name:   "result callback error throws",
			script: `m.app.websocket().send('/foo', { text: 'hello' }, { result: () => { throw new Error('CALLBACK ERROR') } })`,
			client: func(common.WebsocketSendArgs) (*common.WebsocketSendResult, error) {
				return &common.WebsocketSendResult{}, nil
			},
			thrown: "CALLBACK ERROR",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			loop, host := newModuleLoop(t)
			host.WebsocketClientTest = &enginetest.WebsocketClient{SendFunc: tc.client}

			_, err := evalOnLoop(loop, `const m = require('mokapi')`)
			r.NoError(t, err)
			_, err = evalOnLoop(loop, tc.script)
			r.ErrorContains(t, err, tc.thrown)
		})
	}
}

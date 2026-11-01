package engine_test

import (
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js/mokapi"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvent_Http(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		run    func(evt common.EventEmitter) []*common.Action
		test   func(t *testing.T, actions []*common.Action, err error)
	}{
		{
			name: "method does not match",
			script: `import { app } from 'mokapi'
export default () => {
	app.http().get('/foo', (req, res) => {
		res.data = 'bar'
	})
}`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{}
				actions := evt.EmitHttp(&common.HttpEventRequest{
					Method: "POST",
				}, res)
				require.Nil(t, res.Data)
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				require.NoError(t, err)
				require.Len(t, actions, 0)
			},
		},
		{
			name: "route does not match",
			script: `import { app } from 'mokapi'
export default () => {
	app.http().get('/foo', (req, res) => {
		res.data = 'bar'
	})
}`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{}
				actions := evt.EmitHttp(&common.HttpEventRequest{
					Method: "GET",
					Key:    "/bar",
				}, res)
				require.Nil(t, res.Data)
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				require.NoError(t, err)
				require.Len(t, actions, 0)
			},
		},
		{
			name: "API does not match",
			script: `import { app } from 'mokapi'
export default () => {
	app.api('api1').http().get('/foo', (req, res) => {
		res.data = 'bar'
	})
}`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{}
				actions := evt.EmitHttp(&common.HttpEventRequest{
					Method: "GET",
					Key:    "/foo",
					Api:    "api2",
				}, res)
				require.Nil(t, res.Data)
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				require.NoError(t, err)
				require.Len(t, actions, 0)
			},
		},
		{
			name: "all match",
			script: `import { app } from 'mokapi'
export default () => {
	app.api('api1').http().get('/foo', (req, res) => {
		res.data = 'bar'
	})
}`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{}
				actions := evt.EmitHttp(&common.HttpEventRequest{
					Method: "GET",
					Key:    "/foo",
					Api:    "api1",
				}, res)
				require.Equal(t, "bar", mokapi.Export(res.Data))
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				require.NoError(t, err)
				require.Len(t, actions, 1)
			},
		},
		{
			name: "operationId does not match",
			script: `import { app } from 'mokapi'
export default () => {
	app.http().operation('foo', (req, res) => {
		res.data = 'bar'
	})
}`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{}
				actions := evt.EmitHttp(&common.HttpEventRequest{
					OperationId: "bar",
				}, res)
				require.Nil(t, mokapi.Export(res.Data))
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				require.NoError(t, err)
				require.Len(t, actions, 0)
			},
		},
		{
			name: "operationId match",
			script: `import { app } from 'mokapi'
export default () => {
	app.http().operation('foo', (req, res) => {
		res.data = 'bar'
	})
}`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{}
				actions := evt.EmitHttp(&common.HttpEventRequest{
					OperationId: "foo",
				}, res)
				require.Equal(t, "bar", mokapi.Export(res.Data))
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				require.NoError(t, err)
				require.Len(t, actions, 1)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			e := enginetest.NewEngine()
			err := e.AddScript(newScript("test.js", tc.script))

			var actions []*common.Action
			if err == nil {
				actions = tc.run(e)
			}
			tc.test(t, actions, err)
		})
	}
}

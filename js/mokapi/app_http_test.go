package mokapi_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/mokapi"
	"mokapi/js/require"
	"net/http"
	"testing"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

func TestModule_AppHttp(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		logger *enginetest.Logger
		test   func(t *testing.T, handlers []common.HTTPHandler, err error)
	}{
		{
			name: "GET handler",
			script: `
const m = require('mokapi')
m.app.http().get('/pets', (req, res) => {})
`,
			test: func(t *testing.T, handlers []common.HTTPHandler, err error) {
				r.Len(t, handlers, 1)
				r.Equal(t, http.MethodGet, handlers[0].Filter.Method)
				r.Equal(t, "/pets", handlers[0].Filter.Path)
			},
		},
		{
			name: "custom method handler",
			script: `
const m = require('mokapi')
m.app.http().foo('/pets', (req, res) => {})
`,
			test: func(t *testing.T, handlers []common.HTTPHandler, err error) {
				r.Len(t, handlers, 1)
				r.Equal(t, "FOO", handlers[0].Filter.Method)
				r.Equal(t, "/pets", handlers[0].Filter.Path)
			},
		},
		{
			name: "api and post handler",
			script: `
const m = require('mokapi')
m.app.api('foo').http().post('/pets', (req, res) => {})
`,
			test: func(t *testing.T, handlers []common.HTTPHandler, err error) {
				r.Len(t, handlers, 1)
				r.Equal(t, "foo", handlers[0].Filter.Api)
				r.Equal(t, http.MethodPost, handlers[0].Filter.Method)
				r.Equal(t, "/pets", handlers[0].Filter.Path)
			},
		},
		{
			name: "api and delete handler",
			script: `
const m = require('mokapi')
m.app.http().api('foo').delete('/pets', (req, res) => {})
`,
			test: func(t *testing.T, handlers []common.HTTPHandler, err error) {
				r.Len(t, handlers, 1)
				r.Equal(t, "foo", handlers[0].Filter.Api)
				r.Equal(t, http.MethodDelete, handlers[0].Filter.Method)
				r.Equal(t, "/pets", handlers[0].Filter.Path)
			},
		},
		{
			name: "use handler",
			script: `
const m = require('mokapi')
m.app.http().use((req, res) => {})
`,
			test: func(t *testing.T, handlers []common.HTTPHandler, err error) {
				r.Len(t, handlers, 1)
				r.Equal(t, "", handlers[0].Filter.Method)
				r.Equal(t, "", handlers[0].Filter.Path)
			},
		},
		{
			name: "route handler",
			script: `
const m = require('mokapi')
m.app.http().route('/pets').get((req, res) => {})
`,
			test: func(t *testing.T, handlers []common.HTTPHandler, err error) {
				r.Len(t, handlers, 1)
				r.Equal(t, http.MethodGet, handlers[0].Filter.Method)
				r.Equal(t, "/pets", handlers[0].Filter.Path)
			},
		},
		{
			name: "route handler custom method",
			script: `
const m = require('mokapi')
m.app.http().route('/pets').foo((req, res) => {})
`,
			test: func(t *testing.T, handlers []common.HTTPHandler, err error) {
				r.Len(t, handlers, 1)
				r.Equal(t, "FOO", handlers[0].Filter.Method)
				r.Equal(t, "/pets", handlers[0].Filter.Path)
			},
		},
		{
			name: "route handler use",
			script: `
const m = require('mokapi')
m.app.http().route('/pets').use((req, res) => {})
`,
			test: func(t *testing.T, handlers []common.HTTPHandler, err error) {
				r.Len(t, handlers, 1)
				r.Equal(t, "", handlers[0].Filter.Method)
				r.Equal(t, "/pets", handlers[0].Filter.Path)
			},
		},
		{
			name: "route handler use and get",
			script: `
const m = require('mokapi')
m.app.http().route('/pets').use((req, res) => {}).get((req, res) => {})
`,
			test: func(t *testing.T, handlers []common.HTTPHandler, err error) {
				r.Len(t, handlers, 2)
				r.Equal(t, "", handlers[0].Filter.Method)
				r.Equal(t, "/pets", handlers[0].Filter.Path)
				r.Equal(t, http.MethodGet, handlers[1].Filter.Method)
				r.Equal(t, "/pets", handlers[1].Filter.Path)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			reg, err := require.NewRegistry()
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

			var handlers []common.HTTPHandler
			host.OnHttpFunc = func(filter common.HttpFilter, do common.EventHandler, args common.EventArgs) {
				handlers = append(handlers, common.HTTPHandler{
					Filter:  filter,
					Execute: do,
				})
			}
			_, err = vm.RunString(tc.script)
			r.NoError(t, err)

			tc.test(t, handlers, err)
		})
	}
}

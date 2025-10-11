package mustache_test

import (
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/mustache"
	"mokapi/js/require"
	"testing"
)

func TestMustache(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "render",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				}}

				v, err := vm.RunString(`
					const m = require("mokapi/mustache")
					m.render('{{ foo }}', {'foo': 'bar'})
				`)
				r.NoError(t, err)
				r.Equal(t, "bar", v.Export())
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			vm := goja.New()
			host := &enginetest.Host{}
			js.EnableInternal(vm, host, &eventloop.EventLoop{}, &dynamic.Config{})
			req, err := require.NewRegistry()
			r.NoError(t, err)
			req.Enable(vm)
			req.RegisterNativeModule("mokapi/mustache", mustache.Require)

			tc.test(t, vm, host)
		})
	}
}

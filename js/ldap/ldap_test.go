package ldap_test

import (
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/ldap"
	"mokapi/js/require"
	"testing"
)

func TestLdap(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "ResultCode",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				}}

				v, err := vm.RunString(`
					const ldap = require("mokapi/ldap")
					ldap.ResultCode.SizeLimitExceeded
				`)
				r.NoError(t, err)
				r.Equal(t, int64(4), v.Export())
			},
		},
		{
			name: "SearchScope",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.KafkaClientTest = &enginetest.KafkaClient{ProduceFunc: func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				}}

				v, err := vm.RunString(`
					const ldap = require("mokapi/ldap")
					ldap.SearchScope.WholeSubtree
				`)
				r.NoError(t, err)
				r.Equal(t, int64(3), v.Export())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			vm := goja.New()
			host := &enginetest.Host{}
			js.EnableInternal(vm, host, &eventloop.EventLoop{}, &dynamic.Config{})
			req, err := require.NewRegistry()
			r.NoError(t, err)
			req.Enable(vm)
			req.RegisterNativeModule("mokapi/ldap", ldap.Require)

			tc.test(t, vm, host)
		})
	}
}

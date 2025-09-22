package faker_test

import (
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/faker"
	"mokapi/js/require"
	"mokapi/schema/json/generator"
	"testing"
)

func TestOpenAPI(t *testing.T) {
	cleanup := func(host *enginetest.Host) {
		for index := len(host.CleanupFuncs) - 1; index >= 0; index-- {
			host.CleanupFuncs[index]()
		}
	}

	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
type: 'string' 
})
				`)
				r.NoError(t, err)
				r.Equal(t, "XidZuoWq ", v.Export())
			},
		},
		{
			name: "types",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
type: ['string', 'integer']
})
				`)
				r.NoError(t, err)
				r.Equal(t, 7.424164296119123e+18, v.Export())
			},
		},
		{
			name: "invalid type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
type: 123
})
				`)
				r.EqualError(t, err, "unexpected type for 'type': Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "one type is invalid",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
type: [123]
})
				`)
				r.EqualError(t, err, "unexpected type for 'type': Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "enum",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
enum: [123, 'foo']
})
				`)
				r.NoError(t, err)
				r.Equal(t, "foo", v.Export())
			},
		},
		{
			name: "invalid enum type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
enum: 123
})
				`)
				r.EqualError(t, err, "unexpected type for 'enum': got Integer, expected Array at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "const",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
const: 'foo'
})
				`)
				r.NoError(t, err)
				r.Equal(t, "foo", v.Export())
			},
		},
		{
			name: "default",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
default: 'foo'
})
				`)
				r.NoError(t, err)
				r.Equal(t, "", v.Export())
			},
		},
		{
			name: "example",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({ example: 123 })
				`)
				r.NoError(t, err)
				r.Equal(t, int64(123), v.Export())
			},
		},
		{
			name: "examples",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
examples: [123, 789]
})
				`)
				r.NoError(t, err)
				r.Equal(t, int64(789), v.Export())
			},
		},
		{
			name: "invalid example type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
examples: 123
})
				`)
				r.EqualError(t, err, "unexpected type for 'enum': got Integer, expected Array at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "multipleOf",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
multipleOf: 3
})
				`)
				r.NoError(t, err)
				r.Equal(t, int64(6648), v.Export())
			},
		},
		{
			name: "invalid multipleOf type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
multipleOf: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'multipleOf': got String, expected Number at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "maximum",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
maximum: 3
})
				`)
				r.NoError(t, err)
				r.Equal(t, -1.6332447240352712e+308, v.Export())
			},
		},
		{
			name: "invalid maximum type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
maximum: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'maximum': got String, expected Number at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "exclusiveMaximum",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
exclusiveMaximum: 3
})
				`)
				r.NoError(t, err)
				r.Equal(t, -1.6332447240352712e+308, v.Export())
			},
		},
		{
			name: "invalid exclusiveMaximum type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
exclusiveMaximum: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'exclusiveMaximum': got String, expected Number or Boolean at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "minimum",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
minimum: 3
})
				`)
				r.NoError(t, err)
				r.Equal(t, 1.644484108270445e+307, v.Export())
			},
		},
		{
			name: "invalid minimum type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
minimum: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'minimum': got String, expected Number at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "exclusiveMinimum",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
exclusiveMinimum: 3
})
				`)
				r.NoError(t, err)
				r.Equal(t, 1.644484108270445e+307, v.Export())
			},
		},
		{
			name: "invalid exclusiveMinimum type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
exclusiveMinimum: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'exclusiveMinimum': got String, expected Number or Boolean at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "maxLength",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
maxLength: 3
})
				`)
				r.NoError(t, err)
				r.Equal(t, "X", v.Export())
			},
		},
		{
			name: "invalid maxLength type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
maxLength: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'maxLength': got String, expected Number at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "minLength",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
minLength: 3
})
				`)
				r.NoError(t, err)
				r.Equal(t, "XidZuoWq vY5el", v.Export())
			},
		},
		{
			name: "invalid minLength type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
minLength: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'minLength': got String, expected Number at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "pattern",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
pattern: 'foo.{3}'
})
				`)
				r.NoError(t, err)
				r.Equal(t, "foo@|R", v.Export())
			},
		},
		{
			name: "invalid pattern type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
pattern: 123
})
				`)
				r.EqualError(t, err, "unexpected type for 'pattern': got Integer, expected String at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "format",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
format: 'date-time',
})
				`)
				r.NoError(t, err)
				r.Equal(t, "2035-01-24T13:00:35Z", v.Export())
			},
		},
		{
			name: "invalid format type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
format: 123
})
				`)
				r.EqualError(t, err, "unexpected type for 'format': got Integer, expected String at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "items",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
items: { type: 'string' },
})
				`)
				r.NoError(t, err)
				r.Equal(t, []any{"", "idZ", "wYx?vY5elXhlD", "VhgPevuwyrNrL", "lVeCZKW1JKqG"}, v.Export())
			},
		},
		{
			name: "invalid format type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
items: 123
})
				`)
				r.EqualError(t, err, "expect JSON schema but got: Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "maxItems",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
maxItems: 3,
})
				`)
				r.NoError(t, err)
				r.Equal(t, []any{true}, v.Export())
			},
		},
		{
			name: "invalid maxItems type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
maxItems: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'maxItems': got String, expected Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "minItems",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
minItems: 3,
})
				`)
				r.NoError(t, err)
				r.Equal(t, []any{true, int64(6224634831868504800), "ZuoWq vY5elXhlD", []interface{}{2.3090412168364615e+307, "lYehCIA", map[string]interface{}{"caravan": true, "hail": 2.536044080333601e+307, "mob": int64(-287411453310397474), "scale": true}, false}, false, []interface{}{true, int64(1769588974883905016), 7.53537259781811e+307}, false, "LWxKt"}, v.Export())
			},
		},
		{
			name: "invalid minItems type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
minItems: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'minItems': got String, expected Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "uniqueItems",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
uniqueItems: true,
})
				`)
				r.NoError(t, err)
				r.Equal(t, []any{true, int64(6224634831868504800), "ZuoWq vY5elXhlD", []any{2.3090412168364615e+307, "lYehCIA", map[string]any{"caravan": true, "hail": 2.536044080333601e+307, "mob": int64(-287411453310397474), "scale": true}, false}, false}, v.Export())
			},
		},
		{
			name: "invalid uniqueItems type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
uniqueItems: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'uniqueItems': got String, expected Boolean at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "contains",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
contains: { type: 'string' }
})
				`)
				r.NoError(t, err)
				r.Equal(t, []any{[]any{true, int64(1769588974883905016), 7.53537259781811e+307}, "wYx?vY5elXhlD", []any{2.3090412168364615e+307, "lYehCIA", map[string]any{"caravan": true, "hail": 2.536044080333601e+307, "mob": int64(-287411453310397474), "scale": true}, false}, false, "idZ"}, v.Export())
			},
		},
		{
			name: "invalid contains type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
contains: 'foo'
})
				`)
				r.EqualError(t, err, "expect JSON schema but got: String at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "minContains",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
contains: { type: 'string' },
minContains: 4
})
				`)
				r.NoError(t, err)
				r.Equal(t, []any{"LWxKt", "wYx?vY5elXhlD", "VhgPevuwyrNrL", 3.411011869388561e+307, 6.118679555323457e+307, "lVeCZKW1JKqG", true, []interface{}{}, "idZ"}, v.Export())
			},
		},
		{
			name: "invalid minContains type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
minContains: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'minContains': got String, expected Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "maxContains",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
contains: { type: 'string' },
maxContains: 2
})
				`)
				r.NoError(t, err)
				r.Equal(t, []any{" vY5elXhlD4ezl", 1.260262124648351e+307, 1.1925608819819514e+308, int64(3859340644268985534), "idZ"}, v.Export())
			},
		},
		{
			name: "invalid maxContains type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
maxContains: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'maxContains': got String, expected Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "properties",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
properties: { foo: { type: 'string' } }
},)
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]any{"foo": "XidZuoWq "}, v.Export())
			},
		},
		{
			name: "invalid properties type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
properties: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'properties': got String, expected Object at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "maxProperties",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
maxProperties: 3
},)
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]any{"woman": int64(6224634831868504800)}, v.Export())
			},
		},
		{
			name: "invalid maxProperties type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
maxProperties: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'maxProperties': got String, expected Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "minProperties",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
minProperties: 3
},)
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]any{"bunch": 1.1925608819819514e+308, "gang": int64(3859340644268985534), "growth": " vY5elXhlD4ezl", "woman": 1.260262124648351e+307}, v.Export())
			},
		},
		{
			name: "invalid minProperties type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
minProperties: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'minProperties': got String, expected Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "required",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
required: ['foo', 'bar', 'baz']
},)
				`)
				r.NoError(t, err)
				m := v.Export()
				r.Contains(t, m, "foo")
				r.Contains(t, m, "bar")
				r.Contains(t, m, "baz")
				r.Equal(t, map[string]any{"bar": int64(6224634831868504800), "baz": "ZuoWq vY5elXhlD", "foo": true}, v.Export())
			},
		},
		{
			name: "required but not defined in properties",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
properties: { foo: { type: 'string' } },
required: ['bar']
},)
				`)
				r.NoError(t, err)
				m := v.Export()
				r.Contains(t, m, "foo")
				r.Contains(t, m, "bar")
				r.Equal(t, map[string]any{"bar": 1.1291386311317026e+308, "foo": "XidZuoWq "}, v.Export())
			},
		},
		{
			name: "invalid required type",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('faker')
					m.fake({
'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base',
required: 'foo'
})
				`)
				r.EqualError(t, err, "unexpected type for 'required': got String, expected Array at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(11)

			reg, err := require.NewRegistry()
			reg.RegisterNativeModule("faker", faker.Require)
			r.NoError(t, err)

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			defer cleanup(host)
			loop := eventloop.New(vm)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			tc.test(t, vm, host)
		})
	}
}

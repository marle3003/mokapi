package faker_test

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/faker"
	"mokapi/js/require"
	"mokapi/schema/json/generator"
	"testing"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

func TestFaker_Schema(t *testing.T) {
	cleanup := func(host *enginetest.Host) {
		for index := len(host.CleanupFuncs) - 1; index >= 0; index-- {
			host.CleanupFuncs[index]()
		}
	}

	testcases := []struct {
		name               string
		schema             string
		optionalProperties string
		test               func(t *testing.T, v goja.Value, err error)
	}{
		{
			name:   "type",
			schema: "{ type: 'string' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, "XidZuoWq ", v.Export())
			},
		},
		{
			name:   "types",
			schema: "{ type: ['string', 'integer'] }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, 7.424164296119123e+18, v.Export())
			},
		},
		{
			name:   "invalid type",
			schema: "{ type: 123 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'type': Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "one type is invalid",
			schema: "{ type: [123] }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'type': Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "enum",
			schema: "{ enum: [123, 'foo'] }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, "foo", v.Export())
			},
		},
		{
			name:   "invalid enum type",
			schema: "{ enum: 123 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'enum': got Integer, expected Array at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "const",
			schema: "{ const: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, "foo", v.Export())
			},
		},
		{
			name:   "default",
			schema: "{ default: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, "", v.Export())
			},
		},
		{
			name:   "example",
			schema: "{ example: 123 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, int64(123), v.Export())
			},
		},
		{
			name:   "examples",
			schema: "{ examples: [123, 789] }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, int64(789), v.Export())
			},
		},
		{
			name:   "invalid example type",
			schema: "{ examples: 123 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'enum': got Integer, expected Array at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "multipleOf",
			schema: "{ multipleOf: 3 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, int64(6648), v.Export())
			},
		},
		{
			name:   "invalid multipleOf type",
			schema: "{ multipleOf: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'multipleOf': got String, expected Number at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "maximum",
			schema: "{ maximum: 3 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, -1.6332447240352712e+308, v.Export())
			},
		},
		{
			name:   "invalid maximum type",
			schema: "{ maximum: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'maximum': got String, expected Number at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "exclusiveMaximum",
			schema: "{ exclusiveMaximum: 3 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, -1.6332447240352712e+308, v.Export())
			},
		},
		{
			name:   "invalid exclusiveMaximum type",
			schema: "{ exclusiveMaximum: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'exclusiveMaximum': got String, expected Number or Boolean at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "minimum",
			schema: "{ minimum: 3 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, 1.644484108270445e+307, v.Export())
			},
		},
		{
			name:   "invalid minimum type",
			schema: "{ minimum: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'minimum': got String, expected Number at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "exclusiveMinimum",
			schema: "{ exclusiveMinimum: 3 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, 1.644484108270445e+307, v.Export())
			},
		},
		{
			name:   "invalid exclusiveMinimum type",
			schema: "{ exclusiveMinimum: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'exclusiveMinimum': got String, expected Number or Boolean at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "maxLength",
			schema: "{ maxLength: 3 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, "X", v.Export())
			},
		},
		{
			name:   "invalid maxLength type",
			schema: "{ maxLength: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'maxLength': got String, expected Number at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "minLength",
			schema: "{ minLength: 3 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, "XidZuoWq vY5el", v.Export())
			},
		},
		{
			name:   "invalid minLength type",
			schema: "{ minLength: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'minLength': got String, expected Number at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "pattern",
			schema: "{ pattern: 'foo.{3}' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, "foo@|R", v.Export())
			},
		},
		{
			name:   "invalid pattern type",
			schema: "{ pattern: 123 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'pattern': got Integer, expected String at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "format",
			schema: "{ format: 'date-time', }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, "2035-01-24T13:00:35Z", v.Export())
			},
		},
		{
			name:   "invalid format type",
			schema: "{ format: 123 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'format': got Integer, expected String at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "items",
			schema: "{ items: { type: 'string' }, }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, []any{"", "idZ", "wYx?vY5elXhlD", "VhgPevuwyrNrL", "lVeCZKW1JKqG"}, v.Export())
			},
		},
		{
			name:   "invalid format type",
			schema: "{ items: 123 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "expect JSON schema but got: Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "maxItems",
			schema: "{ maxItems: 3, }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, []any{true}, v.Export())
			},
		},
		{
			name:   "invalid maxItems type",
			schema: "{ maxItems: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'maxItems': got String, expected Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "minItems",
			schema: "{ minItems: 3, }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, []any{true, int64(6224634831868504800), "ZuoWq vY5elXhlD", []interface{}{2.3090412168364615e+307, "lYehCIA", map[string]interface{}{"caravan": true, "hail": 2.536044080333601e+307, "mob": int64(-287411453310397474), "scale": true}, false}, false, []interface{}{true, int64(1769588974883905016), 7.53537259781811e+307}, false, "LWxKt"}, v.Export())
			},
		},
		{
			name:   "invalid minItems type",
			schema: "{ minItems: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'minItems': got String, expected Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "uniqueItems",
			schema: "{ uniqueItems: true, }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, []any{true, int64(6224634831868504800), "ZuoWq vY5elXhlD", []any{2.3090412168364615e+307, "lYehCIA", map[string]any{"caravan": true, "hail": 2.536044080333601e+307, "mob": int64(-287411453310397474), "scale": true}, false}, false}, v.Export())
			},
		},
		{
			name:   "invalid uniqueItems type",
			schema: "{ uniqueItems: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'uniqueItems': got String, expected Boolean at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "prefixItems",
			schema: "{ prefixItems: [{ type: 'string' }, { type: 'boolean' }], items: false }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, []any{"", true}, v.Export())
			},
		},
		{
			name:   "invalid uniqueItems type",
			schema: "{ uniqueItems: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'uniqueItems': got String, expected Boolean at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "contains",
			schema: "{ contains: { type: 'string' } }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, []any{[]any{true, int64(1769588974883905016), 7.53537259781811e+307}, "wYx?vY5elXhlD", []any{2.3090412168364615e+307, "lYehCIA", map[string]any{"caravan": true, "hail": 2.536044080333601e+307, "mob": int64(-287411453310397474), "scale": true}, false}, false, "idZ"}, v.Export())
			},
		},
		{
			name:   "invalid contains type",
			schema: "{ contains: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "expect JSON schema but got: String at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "minContains",
			schema: "{ contains: { type: 'string' }, minContains: 4 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, []any{"LWxKt", "wYx?vY5elXhlD", "VhgPevuwyrNrL", 3.411011869388561e+307, 6.118679555323457e+307, "lVeCZKW1JKqG", true, []interface{}{}, "idZ"}, v.Export())
			},
		},
		{
			name:   "invalid minContains type",
			schema: "{ minContains: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'minContains': got String, expected Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "maxContains",
			schema: "{ contains: { type: 'string' }, maxContains: 2 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, []any{" vY5elXhlD4ezl", 1.260262124648351e+307, 1.1925608819819514e+308, int64(3859340644268985534), "idZ"}, v.Export())
			},
		},
		{
			name:   "invalid maxContains type",
			schema: "{ maxContains: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'maxContains': got String, expected Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "properties",
			schema: "{ properties: { foo: { type: 'string' } } }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, map[string]any{"foo": "XidZuoWq "}, v.Export())
			},
		},
		{
			name:   "invalid properties type",
			schema: "{ properties: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'properties': got String, expected Object at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "maxProperties",
			schema: "{ maxProperties: 3 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, map[string]any{"woman": int64(6224634831868504800)}, v.Export())
			},
		},
		{
			name:   "invalid maxProperties type",
			schema: "{ maxProperties: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'maxProperties': got String, expected Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "minProperties",
			schema: "{ minProperties: 3 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, map[string]any{"bunch": 1.1925608819819514e+308, "gang": int64(3859340644268985534), "growth": " vY5elXhlD4ezl", "woman": 1.260262124648351e+307}, v.Export())
			},
		},
		{
			name:   "invalid minProperties type",
			schema: "{ minProperties: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'minProperties': got String, expected Integer at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "patternProperties",
			schema: "{ patternProperties: { '^S_': { type: 'string' } }, type: 'object' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, map[string]any{"S_4VX": "Ozw"}, v.Export())
			},
		},
		{
			name:   "invalid patternProperties type",
			schema: "{ patternProperties: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'patternProperties': got String, expected Object at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "additionalProperties false",
			schema: "{ additionalProperties: false }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{}, v.Export())
			},
		},
		{
			name:   "additionalProperties with type string",
			schema: "{ additionalProperties: { type: 'string' }, maxProperties: 3 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, map[string]any{"bunch": "", "gang": "x?vY5elXhlD4ez", "woman": "zw"}, v.Export())
			},
		},
		{
			name:   "invalid additionalProperties type",
			schema: "{ additionalProperties: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "parse 'additionalProperties' failed: expect JSON schema but got: String at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "properties, patternProperties and additionalProperties",
			schema: "{ properties: { builtin: { type: 'number' } }, patternProperties: { '^S_': { type: 'string' }, '^I_': { type: 'integer' } }, additionalProperties: { type: 'string' } }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, map[string]any{
					"I_4VX":    int64(4336346668576341540),
					"S_kY9X3W": "m",
					"builtin":  5.544677937412537e+307,
					"group":    "CKu",
					"ocean":    "LJgmr9arWgSfi",
					"party":    "m",
					"sock":     "hlD4ezlYehCIA0O",
					"time":     "jLWxKtR4",
				}, v.Export())
			},
		},
		{
			name:   "propertyNames with only capital letter at the beginning",
			schema: "{ propertyNames: { pattern: '^[A-Z][A-Za-z0-9_]*$' }, minProperties: 4, maxProperties: 4 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, map[string]any{
					"Gt8":     "X",
					"Jd":      "D4ezlYehCIA0O",
					"MTgAEPI": int64(-3961972703762434141),
					"PgmZE3":  false,
				}, v.Export())
			},
		},
		{
			name:   "invalid propertyNames type",
			schema: "{ propertyNames: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "parse 'propertyNames' failed: expect JSON schema but got: String at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name: "dependentRequired",
			// to force taking into account dependentRequired,
			optionalProperties: "0",
			schema: `{ 
  properties: {
    name: { type: 'string' },
    credit_card: { type: 'string' },
    billing_address: { type: 'string' }
  },
  required: ['name','credit_card'],
  dependentRequired: {
    credit_card: ['billing_address']
  }
}`,
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, map[string]any{
					"name":            "TerraCove",
					"credit_card":     "3645994899536906",
					"billing_address": "hgPevuwyr",
				}, v.Export())
			},
		},
		{
			name:   "invalid dependentRequired type",
			schema: "{ dependentRequired: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'dependentRequired': got String, expected Object at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "invalid dependentRequired type not array",
			schema: "{ dependentRequired: { foo: 123 } }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'dependentRequired.foo': got Integer, expected Array at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "invalid dependentRequired type not string array",
			schema: "{ dependentRequired: { foo: [123] } }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'dependentRequired.foo[0]': got Integer, expected String at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},

		{
			name:   "required",
			schema: "{ required: ['foo', 'bar', 'baz'] }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				m := v.Export()
				r.Contains(t, m, "foo")
				r.Contains(t, m, "bar")
				r.Contains(t, m, "baz")
				r.Equal(t, map[string]any{"bar": int64(6224634831868504800), "baz": "ZuoWq vY5elXhlD", "foo": true}, m)
			},
		},
		{
			name: "dependentSchemas",
			// to force taking into account dependentRequired,
			optionalProperties: "0",
			schema: `{ 
  properties: {
    name: { type: 'string' },
    credit_card: { type: 'string' },
  },
  required: ['name','credit_card'],
  dependentSchemas: {
    credit_card: {
		properties: {
		  billing_address: { type: 'string' },
        },
        required: ['billing_address']
	}
  }
}`,
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, map[string]any{
					"name":            "TerraCove",
					"credit_card":     "3645994899536906",
					"billing_address": "hgPevuwyr",
				}, v.Export())
			},
		},
		{
			name:   "invalid dependentSchemas type",
			schema: "{ dependentSchemas: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'dependentSchemas': got String, expected Object at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
		{
			name:   "required but not defined in properties",
			schema: "{ properties: { foo: { type: 'string' } }, required: ['bar'] }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				m := v.Export()
				r.Contains(t, m, "bar")
				r.Equal(t, map[string]any{"bar": 1.1291386311317026e+308, "foo": "XidZuoWq "}, m)
			},
		},
		{
			name:   "invalid required type",
			schema: "{ required: 'foo' }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.EqualError(t, err, "unexpected type for 'required': got String, expected Array at mokapi/js/faker.(*Module).Fake-fm (native)")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.optionalProperties != "" {
				generator.SetConfig(static.DataGen{
					OptionalProperties: tc.optionalProperties,
				})
				defer generator.SetConfig(static.DataGen{})
			}

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

			_, err = vm.RunString("const m = require('faker')")
			r.NoError(t, err)

			t.Run("json", func(t *testing.T) {
				generator.Seed(11)

				v, err := vm.RunString(fmt.Sprintf(`
m.fake(%s)`, tc.schema))
				tc.test(t, v, err)
			})

			t.Run("openapi", func(t *testing.T) {
				generator.Seed(11)

				v, err := vm.RunString(fmt.Sprintf(`
m.fake(Object.assign(
	{'$schema': 'https://spec.openapis.org/oas/3.1/dialect/base'},
	%s
))`, tc.schema))
				tc.test(t, v, err)
			})
		})
	}
}

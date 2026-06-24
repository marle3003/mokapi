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
				r.Equal(t, "fnsy", v.Export())
			},
		},
		{
			name:   "types",
			schema: "{ type: ['string', 'integer'] }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, "fg", v.Export())
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
				r.Equal(t, int64(123), v.Export())
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
				r.InDelta(t, 971925.852188296, v.Export(), 0.000001)
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
				r.Equal(t, int64(123), v.Export())
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
				r.Equal(t, int64(270), v.Export())
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
				r.Equal(t, -229696.7569350967, v.Export())
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
				r.Equal(t, -229696.7569350967, v.Export())
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
				r.Equal(t, 770301.6212593103, v.Export())
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
				r.Equal(t, 770301.6212593103, v.Export())
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
				r.Equal(t, "", v.Export())
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
				r.Equal(t, "fnsyx7yIkhyaaK", v.Export())
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
				r.Equal(t, "2033-11-06T04:31:13Z", v.Export())
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
				r.Equal(t, []any{"fg", "yx7yIkhya", "y", "AQyBy", "LPbGpPaituc"}, v.Export())
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
				r.Equal(t, []any{}, v.Export())
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
			schema: "{ minItems: 3, items: { type: 'string' } }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, []interface{}{"fg", "yx7yIkhya", "y", "AQyBy", "LPbGpPaituc", "Lx47fDnQE", "gPAl89Xbz vlNV", "Zwkx5"}, v.Export())
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
			schema: "{ items: { type: 'integer' }, uniqueItems: true, }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, []any{int64(711901), int64(-600929), int64(-599435), int64(56944), int64(-537198)}, v.Export())
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
				r.Equal(t, []any{"fg", true}, v.Export())
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
			schema: "{ contains: { const: 'foo' }, items: { type: 'string' } }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, []any{"foo", "foo", "foo", "foo", "foo"}, v.Export())
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
				r.Equal(t, []any{"ftyw5heL", "", "IkhyaaKAQyByP", "ZpI", "kx5Kt6ckaieG", "PbG", "2mXR5eZVgPAl8", "blIzvlNVuk", "nsyx7"}, v.Export())
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
			schema: "{ contains: { type: 'string' }, maxContains: 2, items: { type: 'string' } }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, []any{"", "nsyx7"}, v.Export())
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
				r.Equal(t, map[string]any{}, v.Export())
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
				r.Equal(t, map[string]any{}, v.Export())
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
			schema: "{ minProperties: 3, additionalProperties: { type: 'string' } }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, map[string]any{"class": "zvlNVukPyqDClq", "fear": "lsTI5ydf yByPS<", "heart": "l89Xb", "pack": "GpPaitucts2mXR5", "problem": "Qzy", "trip": "QEmWm"}, v.Export())
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
				r.Equal(t, map[string]any{"S_4V": "syx"}, v.Export())
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
				r.Equal(t, map[string]any{"fear": "VbyxlsTI5ydf y", "pack": " yPS<", "trip": "GpPaitucts2mXR5"}, v.Export())
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
			schema: "{ properties: { builtin: { type: 'integer' } }, patternProperties: { '^S_': { type: 'string' }, '^I_': { type: 'integer' } }, additionalProperties: { type: 'string' } }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, map[string]any{
					"I_4V":      int64(-927324),
					"S_XkY9X3":  " yByPS<qbft",
					"ambulance": "nI fpidfoDeHcd",
					"boat":      "Clq kaieGDf",
					"body":      "VgPAl89Xbz v",
					"builtin":   int64(-697900),
					"cackle":    "IqiIQzyMKld",
					"grammar":   "qdeZwkx5",
					"harm":      "DnQ",
					"jealousy":  "",
					"problem":   "itucts2mX",
				}, v.Export())
			},
		},
		{
			name:   "propertyNames with only capital letter at the beginning",
			schema: "{ propertyNames: { pattern: '^[A-Z][A-Za-z0-9_]*$' }, minProperties: 4, maxProperties: 4 }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				r.Equal(t, map[string]any{
					"AgAE": int64(-927324), "Gt8P": "K", "KItde": int64(288920), "TmZE": "B",
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
					"billing_address": "PbG", "credit_card": "2252759890799934473", "name": "GroveGuard",
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
				m := v.Export().(map[string]any)
				r.Contains(t, m, "foo")
				r.Contains(t, m, "bar")
				r.Contains(t, m, "baz")
				r.InDelta(t, 824801.9947984695, m["bar"], 0.000001)
				r.Equal(t, int64(-342586), m["baz"])
				r.Equal(t, "nsyx7", m["foo"])
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
					"billing_address": "PbG", "credit_card": "2252759890799934473", "name": "GroveGuard",
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
			schema: "{ properties: { foo: { type: 'string' }}, additionalProperties: { type: 'string' }, required: ['bar'] }",
			test: func(t *testing.T, v goja.Value, err error) {
				r.NoError(t, err)
				m := v.Export()
				r.Contains(t, m, "bar")
				r.Equal(t, map[string]any{"bar": "f yByPS<qbft", "fear": "K", "foo": "gnVbyxlsTI5"}, m)
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
			loop := eventloop.New(vm, host)
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

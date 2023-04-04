package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/media"
	"mokapi/sortedmap"
	"testing"
)

func TestRef_Marshal(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Ref
		data   interface{}
		ct     media.ContentType
		exp    string
	}{
		{
			"no schema",
			&schema.Ref{},
			"foo",
			media.ParseContentType("application/json"),
			`"foo"`,
		},
		{
			"number",
			&schema.Ref{Value: schematest.New("number")},
			3.141,
			media.ParseContentType("application/json"),
			`3.141`,
		},
		{
			"string",
			&schema.Ref{Value: schematest.New("string")},
			"12",
			media.ParseContentType("application/json"),
			`"12"`,
		},
		{
			"integer as string",
			&schema.Ref{Value: schematest.New("integer")},
			"12",
			media.ParseContentType("application/json"),
			`12`,
		},
		{
			"array of integer",
			&schema.Ref{Value: schematest.New("array", schematest.WithItems(schematest.New("integer")))},
			[]interface{}{12, 13},
			media.ParseContentType("application/json"),
			`[12,13]`,
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := tc.schema.Marshal(tc.data, tc.ct)
			require.NoError(t, err)
			require.Equal(t, tc.exp, string(b))
		})
	}
}

func TestRef_Marshal_Object(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Ref
		data   interface{}
		ct     media.ContentType
		fn     func(t *testing.T, s string)
	}{
		{
			"struct",
			&schema.Ref{Value: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("value", schematest.New("integer")))},
			struct {
				Name  string
				Value int
			}{"foo", 12},
			media.ParseContentType("application/json"),
			func(t *testing.T, s string) {
				require.Equal(t, `{"name":"foo","value":12}`, s)
			},
		},
		{
			"map with key string",
			&schema.Ref{Value: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("value", schematest.New("integer")))},
			map[string]interface{}{"name": "foo", "value": 12},
			media.ParseContentType("application/json"),
			func(t *testing.T, s string) {
				require.Equal(t, `{"name":"foo","value":12}`, s)
			},
		},
		{
			"map with key interface{}",
			&schema.Ref{Value: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("value", schematest.New("integer")))},
			map[interface{}]interface{}{"name": "foo", "value": 12},
			media.ParseContentType("application/json"),
			func(t *testing.T, s string) {
				require.Equal(t, `{"name":"foo","value":12}`, s)
			},
		},
		{
			"map with key interface{} and empty properties",
			&schema.Ref{Value: schematest.New("object")},
			map[interface{}]interface{}{"name": "foo", "value": 12},
			media.ParseContentType("application/json"),
			func(t *testing.T, s string) {
				// order of properties is not guaranteed
				require.True(t, s == `{"name":"foo","value":12}` || s == `{"value":12,"name":"foo"}`, s)
			},
		},
		{
			"map as map",
			&schema.Ref{Value: schematest.New("object",
				schematest.WithAdditionalProperties(schematest.New("object", schematest.WithProperty("name", schematest.New("string")))),
			)},
			map[interface{}]interface{}{"x": map[string]string{"name": "x"}, "y": map[string]string{"name": "y"}},
			media.ParseContentType("application/json"),
			func(t *testing.T, s string) {
				require.True(t, s == `{"x":{"name":"x"},"y":{"name":"y"}}` || s == `{"y":{"name":"y"},"x":{"name":"x"}}`, s)
			},
		},
		{
			"map not free-form",
			&schema.Ref{Value: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithFreeForm(false),
			)},
			map[interface{}]interface{}{"name": "foo", "value": 12},
			media.ParseContentType("application/json"),
			func(t *testing.T, s string) {
				require.True(t, s == `{"name":"foo"}`, s)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := tc.schema.Marshal(tc.data, tc.ct)
			require.NoError(t, err)
			tc.fn(t, string(b))
		})
	}
}

func TestRef_Marshal_AnyOf(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"any",
			func(t *testing.T) {
				s := &schema.Ref{Value: schematest.New("",
					schematest.Any(
						schematest.New("object", schematest.WithProperty("foo", schematest.New("string")), schematest.WithFreeForm(false)),
						schematest.New("object", schematest.WithProperty("bar", schematest.New("string")), schematest.WithFreeForm(false)),
					))}
				data := map[string]interface{}{"foo": "foo", "value": 12}

				b, err := s.Marshal(data, media.ParseContentType("application/json"))
				require.NoError(t, err)
				require.Equal(t, `{"foo":"foo"}`, string(b))
			},
		},
		{
			"any mixed",
			func(t *testing.T) {
				s := &schema.Ref{Value: schematest.New("",
					schematest.Any(
						schematest.New("object", schematest.WithProperty("foo", schematest.New("string")), schematest.WithFreeForm(false)),
						schematest.New("object", schematest.WithProperty("bar", schematest.New("string")), schematest.WithFreeForm(false)),
					))}
				data := map[string]interface{}{"foo": "foo", "bar": "bar", "value": 12}

				b, err := s.Marshal(data, media.ParseContentType("application/json"))
				require.NoError(t, err)
				require.Equal(t, `{"foo":"foo","bar":"bar"}`, string(b))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.f(t)
		})
	}
}

func TestRef_Marshal_AllOf(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"all from map",
			func(t *testing.T) {
				s := &schema.Ref{Value: schematest.New("",
					schematest.AllOf(
						schematest.New("object", schematest.WithProperty("foo", schematest.New("number"))),
						schematest.New("object", schematest.WithProperty("bar", schematest.New("integer"))),
					))}
				data := map[string]interface{}{"foo": "3.141", "bar": "12"}

				b, err := s.Marshal(data, media.ParseContentType("application/json"))
				require.NoError(t, err)
				require.Equal(t, `{"foo":3.141,"bar":12}`, string(b))
			},
		},
		{
			"all from linked map",
			func(t *testing.T) {
				s := &schema.Ref{Value: schematest.New("",
					schematest.AllOf(
						schematest.New("object", schematest.WithProperty("foo", schematest.New("number"))),
						schematest.New("object", schematest.WithProperty("bar", schematest.New("integer"))),
					))}
				data := sortedmap.NewLinkedHashMap()
				data.Set("foo", "3.141")
				data.Set("bar", "12")

				b, err := s.Marshal(data, media.ParseContentType("application/json"))
				require.NoError(t, err)
				require.Equal(t, `{"foo":3.141,"bar":12}`, string(b))
			},
		},
		{
			"all with additional field",
			func(t *testing.T) {
				s := &schema.Ref{Value: schematest.New("",
					schematest.AllOf(
						schematest.New("object", schematest.WithProperty("foo", schematest.New("number"))),
						schematest.New("object", schematest.WithProperty("bar", schematest.New("integer"))),
					))}
				data := sortedmap.NewLinkedHashMap()
				data.Set("foo", "3.141")
				data.Set("bar", "12")
				data.Set("name", "foobar")

				b, err := s.Marshal(data, media.ParseContentType("application/json"))
				require.NoError(t, err)
				require.Equal(t, `{"foo":3.141,"bar":12,"name":"foobar"}`, string(b))
			},
		},
		{
			"allOf ignores any free-form false",
			func(t *testing.T) {
				s := &schema.Ref{Value: schematest.New("",
					schematest.AllOf(
						schematest.New("object", schematest.WithProperty("foo", schematest.New("string")), schematest.WithFreeForm(false)),
						schematest.New("object", schematest.WithProperty("bar", schematest.New("string")), schematest.WithFreeForm(false)),
					))}
				data := map[string]interface{}{"foo": "foo", "bar": "bar", "value": "test"}

				b, err := s.Marshal(data, media.ParseContentType("application/json"))
				require.NoError(t, err)

				require.Equal(t, `{"foo":"foo","bar":"bar","value":"test"}`, string(b))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.f(t)
		})
	}
}

func TestRef_Marshal_Invalid(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Ref
		data   interface{}
		ct     media.ContentType
		exp    string
	}{
		{
			"number",
			&schema.Ref{Value: schematest.New("number")},
			"foo",
			media.ParseContentType("application/json"),
			`could not parse 'foo' as floating number, expected schema type=number`,
		},
		{
			"min array",
			&schema.Ref{Value: schematest.New("array", schematest.WithItems(schematest.New("integer")), schematest.WithMinItems(3))},
			[]interface{}{12, 13},
			media.ParseContentType("application/json"),
			`validation error minItems on [12 13], expected schema type=array minItems=3`,
		},
		{
			"max array",
			&schema.Ref{Value: schematest.New("array", schematest.WithItems(schematest.New("integer")), schematest.WithMaxItems(1))},
			[]interface{}{12, 13},
			media.ParseContentType("application/json"),
			`validation error maxItems on [12 13], expected schema type=array maxItems=1`,
		},
		{
			"map missing required property",
			&schema.Ref{Value: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("value", schematest.New("integer")),
				schematest.WithRequired("value"),
			)},
			map[interface{}]interface{}{"name": "foo"},
			media.ParseContentType("application/json"),
			`missing required field value in {name: foo}, expected: schema type=object properties=[name, value] required=[value]`,
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.schema.Marshal(tc.data, tc.ct)
			require.EqualError(t, err, tc.exp)
		})
	}
}

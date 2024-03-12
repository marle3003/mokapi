package schema_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/json/schema"
	"mokapi/json/schematest"
	"testing"
)

func TestSchema_Validate(t *testing.T) {
	testcases := []struct {
		name   string
		data   string
		schema *schema.Schema
		test   func(t *testing.T, err error)
	}{
		{
			name:   "empty with no schema",
			data:   `""`,
			schema: nil,
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "empty with schema but no type",
			data:   `""`,
			schema: &schema.Schema{},
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "object and no schema",
			data:   `{"foo": 12}`,
			schema: nil,
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "array and no schema",
			data:   `[1, 2, 3, 4]`,
			schema: nil,
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "object with integer ",
			data:   `{"foo": 12}`,
			schema: schematest.New("object", schematest.WithProperty("foo", schematest.New("integer"))),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "object with string",
			data:   `{"foo": "bar"}`,
			schema: schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "object with string format",
			data: `{"foo": "2021-01-20"}`,
			schema: schematest.New("object",
				schematest.WithProperty("foo",
					schematest.New("string", schematest.WithFormat("date")))),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "object containing array",
			data: `{"foo": ["a", "b", "c"]}`,
			schema: schematest.New("object",
				schematest.WithProperty("foo",
					schematest.New("array", schematest.WithItems("string")))),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "null",
			data:   `{ "foo": null }`,
			schema: schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "validation error on <nil>, expected schema type=string")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var v interface{}
			err := json.Unmarshal([]byte(tc.data), &v)
			require.NoError(t, err)
			err = tc.schema.Validate(v)
			tc.test(t, err)
		})
	}
}

func TestSchema_Validate_String(t *testing.T) {
	testcases := []struct {
		name   string
		data   string
		schema *schema.Schema
		test   func(t *testing.T, err error)
	}{
		{
			name:   "not string",
			data:   `12`,
			schema: schematest.New("string"),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "validation error on 12, expected schema type=string")
			},
		},
		{
			name:   "string",
			data:   `"gbRMaRxHkiJBPta"`,
			schema: schematest.New("string"),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "type not defined",
			data:   `"gbRMaRxHkiJBPta"`,
			schema: &schema.Schema{},
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "by pattern",
			data:   `"013-64-5994"`,
			schema: schematest.New("string", schematest.WithPattern("^\\d{3}-\\d{2}-\\d{4}$")),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "not match pattern",
			data:   `"013-64-59943"`,
			schema: schematest.New("string", schematest.WithPattern("^\\d{3}-\\d{2}-\\d{4}$")),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "value '013-64-59943' does not match pattern, expected schema type=string pattern=^\\d{3}-\\d{2}-\\d{4}$")
			},
		},
		{
			name:   "date",
			data:   `"1908-12-07"`,
			schema: schematest.New("string", schematest.WithFormat("date")),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "not a date",
			data:   `"1908-12-7"`,
			schema: schematest.New("string", schematest.WithFormat("date")),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "value '1908-12-7' does not match format 'date' (RFC3339), expected schema type=string format=date")
			},
		},
		{
			name:   "date-time",
			data:   `"1908-12-07T04:14:25Z"`,
			schema: schematest.New("string", schematest.WithFormat("date-time")),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "not a date-time",
			data:   `"1908-12-07 T04:14:25Z"`,
			schema: schematest.New("string", schematest.WithFormat("date-time")),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "value '1908-12-07 T04:14:25Z' does not match format 'date-time' (RFC3339), expected schema type=string format=date-time")
			},
		},
		{
			name:   "password",
			data:   `"H|$9lb{J<+S;"`,
			schema: schematest.New("string", schematest.WithFormat("password")),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "email",
			data:   `"markusmoen@pagac.net"`,
			schema: schematest.New("string", schematest.WithFormat("email")),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "not an email",
			data:   `"markusmoen@@pagac.net"`,
			schema: schematest.New("string", schematest.WithFormat("email")),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "value 'markusmoen@@pagac.net' does not match format 'email', expected schema type=string format=email")
			},
		},
		{
			name:   "uuid",
			data:   `"590c1440-9888-45b0-bd51-a817ee07c3f2"`,
			schema: schematest.New("string", schematest.WithFormat("uuids")),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "not an uuid",
			data:   `"590c1440-9888-45b0-bd51-a817ee07c3f2a"`,
			schema: schematest.New("string", schematest.WithFormat("uuid")),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "value '590c1440-9888-45b0-bd51-a817ee07c3f2a' does not match format 'uuid', expected schema type=string format=uuid")
			},
		},
		{
			name:   "ipv4",
			data:   `"152.23.53.100"`,
			schema: schematest.New("string", schematest.WithFormat("ipv4")),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "not ipv4",
			data:   `"152.23.53.100."`,
			schema: schematest.New("string", schematest.WithFormat("ipv4")),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "value '152.23.53.100.' does not match format 'ipv4', expected schema type=string format=ipv4")
			},
		},
		{
			name:   "ipv6",
			data:   `"8898:ee17:bc35:9064:5866:d019:3b95:7857"`,
			schema: schematest.New("string", schematest.WithFormat("ipv6")),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "not ipv6",
			data:   `"-8898:ee17:bc35:9064:5866:d019:3b95:7857"`,
			schema: schematest.New("string", schematest.WithFormat("ipv6")),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "value '-8898:ee17:bc35:9064:5866:d019:3b95:7857' does not match format 'ipv6', expected schema type=string format=ipv6")
			},
		},
		{
			name:   "not minLength",
			data:   `"foo"`,
			schema: schematest.New("string", schematest.WithMinLength(4)),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "length of 'foo' is too short, expected schema type=string minLength=4")
			},
		},
		{
			name:   "minLength",
			data:   `"foo"`,
			schema: schematest.New("string", schematest.WithMinLength(3)),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "not maxLength",
			data:   `"foo"`,
			schema: schematest.New("string", schematest.WithMaxLength(2)),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "length of 'foo' is too long, expected schema type=string maxLength=2")
			},
		},
		{
			name:   "maxLength",
			data:   `"foo"`,
			schema: schematest.New("string", schematest.WithMaxLength(3)),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "enum",
			data:   `"foo"`,
			schema: schematest.New("string", schematest.WithEnum([]interface{}{"foo"})),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "not in enum",
			data:   `"foo"`,
			schema: schematest.New("string", schematest.WithEnum([]interface{}{"bar"})),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "validation error foo: value does not match one in the enumeration [bar]")
			},
		},
		{
			name:   "nullable string",
			data:   `null`,
			schema: schematest.NewTypes([]string{"string", "null"}),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var v interface{}
			err := json.Unmarshal([]byte(tc.data), &v)
			require.NoError(t, err)
			err = tc.schema.Validate(v)
			tc.test(t, err)
		})
	}
}

func TestSchema_Validate_OneOf(t *testing.T) {
	testcases := []struct {
		name   string
		data   string
		schema *schema.Schema
		test   func(t *testing.T, err error)
	}{
		{
			name: "valid oneOf",
			data: `{"foo": true}`,
			schema: schematest.New("", schematest.OneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("boolean"))),
			)),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "no match",
			data: `{"foo": "bar", "bar": 12}`,
			schema: schematest.New("", schematest.OneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("bar", schematest.New("boolean"))),
			)),
			test: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(t, err, "expected to match one of schema but it matches none")
			},
		},
		{
			name: "two match",
			data: `{"foo": 12}`,
			schema: schematest.New("", schematest.OneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("number"))),
			)),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "validation error {foo: 12}: it is valid for more than one schema, expected one of schema type=object properties=[foo], schema type=object properties=[foo]")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var v interface{}
			err := json.Unmarshal([]byte(tc.data), &v)
			require.NoError(t, err)
			err = tc.schema.Validate(v)
			tc.test(t, err)
		})
	}
}

func TestSchema_Validate_AllOf(t *testing.T) {
	testcases := []struct {
		name   string
		data   string
		schema *schema.Schema
		test   func(t *testing.T, err error)
	}{
		{
			name: "valid",
			data: `{"foo": 12, "bar": true}`,
			schema: schematest.New("object",
				schematest.AllOf(
					schematest.New("object", schematest.WithProperty("foo", schematest.New("integer"))),
					schematest.New("object", schematest.WithProperty("bar", schematest.New("boolean"))),
				)),
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "missing required",
			data: `{"foo": 12}`,
			schema: schematest.New("object",
				schematest.AllOf(
					schematest.New("object", schematest.WithProperty("foo", schematest.New("integer"))),
					schematest.New("object",
						schematest.WithRequired("bar"),
						schematest.WithProperty("bar", schematest.New("boolean"))),
				)),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "validation error {foo: 12}: value does not match part of allOf: missing required field(s) [bar]")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var v interface{}
			err := json.Unmarshal([]byte(tc.data), &v)
			require.NoError(t, err)
			err = tc.schema.Validate(v)
			tc.test(t, err)
		})
	}
}

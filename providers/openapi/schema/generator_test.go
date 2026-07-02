package schema_test

import (
	"encoding/json"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/schema/json/generator"
	jsonSchema "mokapi/schema/json/schema"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func toFloatP(f float64) *float64 { return &f }
func toBoolP(b bool) *bool        { return &b }

func TestGenerator(t *testing.T) {
	testcases := []struct {
		name   string
		exp    interface{}
		schema *schema.Schema
		test   func(t *testing.T, v interface{}, err error)
	}{
		{
			name:   "no schema",
			schema: nil,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.InDelta(t, 971925.852188296, v, 0.000001)
			},
		},
		{
			name:   "empty schema",
			schema: schematest.New(""),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.InDelta(t, 971925.852188296, v, 0.000001)
			},
		},
		{
			name:   "invalid type",
			schema: schematest.New("foobar"),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "unsupported schema: schema type=foobar")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)

			v, err := schema.CreateValue(tc.schema)
			tc.test(t, v, err)
		})
	}
}

func TestGeneratorString(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Schema
		test   func(t *testing.T, v interface{}, err error)
	}{
		{
			name:   "string",
			schema: schematest.New("string"),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "fnsy", v)
			},
		},
		{
			name:   "by pattern",
			schema: schematest.New("string", schematest.WithPattern("^\\d{3}-\\d{2}-\\d{4}$")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "013-64-5994", v)
			},
		},
		{
			name:   "date",
			schema: schematest.New("string", schematest.WithFormat("date")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2033-11-06", v)
			},
		},
		{
			name:   "date-time",
			schema: schematest.New("string", schematest.WithFormat("date-time")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2033-11-06T04:31:13Z", v)
			},
		},
		{
			name:   "password",
			schema: schematest.New("string", schematest.WithFormat("password")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "L*S9@WG!5x_1", v)
			},
		},
		{
			name:   "email",
			schema: schematest.New("string", schematest.WithFormat("email")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "priscilla.thornton@duncan.biz", v)
			},
		},
		{
			name:   "uuid",
			schema: schematest.New("string", schematest.WithFormat("uuid")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "b4ddf623-4ea6-48e5-9292-541f028d1fdb", v)
			},
		},
		{
			name:   "url",
			schema: schematest.New("string", schematest.WithFormat("{url}")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.deputyinnovative.biz/infrastructures", v)
			},
		},
		{
			name:   "hostname",
			schema: schematest.New("string", schematest.WithFormat("hostname")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "strategicinfrastructures.biz", v)
			},
		},
		{
			name:   "ipv4",
			schema: schematest.New("string", schematest.WithFormat("ipv4")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "180.18.181.251", v)
			},
		},
		{
			name:   "ipv6",
			schema: schematest.New("string", schematest.WithFormat("ipv6")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "ddb4:9212:aab5:87fb:4e33:17a4:f7b9:bf8e", v)
			},
		},
		{
			name:   "beername",
			schema: schematest.New("string", schematest.WithFormat("{beername}")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Sierra Nevada Bigfoot Barleywine Style Ale", v)
			},
		},
		{
			name:   "address",
			schema: schematest.New("string", schematest.WithFormat("{zip} {city}")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "81252 Buffalo", v)
			},
		},
		{
			name:   "uri",
			schema: schematest.New("string", schematest.WithFormat("uri")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.deputyinnovative.biz/infrastructures", v)
			},
		},
		{
			name:   "minLength",
			schema: schematest.New("string", schematest.WithMinLength(25)),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "fnsyx7yIkhyaaKAQyByPS<qbftyw5", v)
			},
		},
		{
			name:   "maxLength",
			schema: schematest.New("string", schematest.WithMaxLength(4)),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "fnsy", v)
			},
		},
		{
			name:   "maxLength",
			schema: schematest.New("string", schematest.WithMaxLength(12)),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "fnsyx7yIkhy", v)
			},
		},
		{
			name:   "minLength with maxLength",
			schema: schematest.New("string", schematest.WithMinLength(3), schematest.WithMaxLength(6)),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "fns", v)
			},
		},
		{
			name:   "minLength equals maxLength",
			schema: schematest.New("string", schematest.WithMinLength(4), schematest.WithMaxLength(4)),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "wfgn", v)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)
			generator.Seed(11)

			v, err := schema.CreateValue(tc.schema)
			tc.test(t, v, err)
		})
	}
}

func TestGeneratorBool(t *testing.T) {
	testdata := []struct {
		name   string
		exp    interface{}
		schema *schema.Schema
	}{
		{
			name:   "boolean",
			exp:    false,
			schema: schematest.New("boolean"),
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.Seed(11)

			o, err := schema.CreateValue(data.schema)
			require.NoError(t, err)
			require.Equal(t, data.exp, o)
		})
	}
}

func TestGeneratorInt(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
	}{
		{
			name:   "int32",
			schema: schematest.New("integer", schematest.WithFormat("int32")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(791768), i)
			},
		},
		{
			name:   "int32 min",
			schema: schematest.New("integer", schematest.WithFormat("int32"), schematest.WithMinimum(10)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(770303), i)
			},
		},
		{
			name:   "int32 max",
			schema: schematest.New("integer", schematest.WithFormat("int32"), schematest.WithMaximum(0)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(-229699), i)
			},
		},
		{
			name:   "int32 min max",
			schema: schematest.New("integer", schematest.WithFormat("int32"), schematest.WithMinimum(-5), schematest.WithMaximum(5)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(3), i)
			},
		},
		{
			name:   "int64",
			schema: schematest.New("integer", schematest.WithFormat("int64")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(791768), i)
			},
		},
		{
			name:   "int64 min",
			schema: schematest.New("integer", schematest.WithFormat("int64"), schematest.WithMinimum(10)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(770303), i)
			},
		},
		{
			name:   "int64 max",
			schema: schematest.New("integer", schematest.WithFormat("int64"), schematest.WithMaximum(0)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-229699), i)
			},
		},
		{
			name:   "int64 min max",
			schema: schematest.New("integer", schematest.WithFormat("int64"), schematest.WithMinimum(-5), schematest.WithMaximum(5)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(3), i)
			},
		},
		{
			name:   "int64 min max positive",
			schema: schematest.New("integer", schematest.WithFormat("int64"), schematest.WithMinimum(4), schematest.WithMaximum(10)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(9), i)
			},
		},
		{
			name:   "int64 min max positive exclusive",
			schema: schematest.New("integer", schematest.WithFormat("int64"), schematest.WithExclusiveMinimum(3), schematest.WithExclusiveMaximum(5)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(4), i)
			},
		},
		{
			name: "int64 min max positive exclusive but error",
			schema: &schema.Schema{
				Type:             jsonSchema.Types{"integer"},
				Format:           "int64",
				Minimum:          toFloatP(4),
				Maximum:          toFloatP(5),
				ExclusiveMinimum: jsonSchema.NewUnionTypeB[float64, bool](true),
				ExclusiveMaximum: jsonSchema.NewUnionTypeB[float64, bool](true),
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "invalid minimum '5' and maximum '4' in schema type=integer format=int64 minimum=4 maximum=5 exclusiveMinimum=true exclusiveMaximum=true")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)

			i, err := schema.CreateValue(tc.schema)
			tc.test(t, i, err)
		})
	}
}

func TestGeneratorFloat(t *testing.T) {
	testdata := []struct {
		name   string
		exp    interface{}
		schema *schema.Schema
	}{
		{
			name:   "float",
			exp:    float32(540601.9),
			schema: schematest.New("number", schematest.WithFormat("float")),
		},
		{
			name:   "float min",
			exp:    float32(770303.25),
			schema: schematest.New("number", schematest.WithFormat("float"), schematest.WithMinimum(10)),
		},
		{
			name:   "float max",
			exp:    float32(-229699.06),
			schema: schematest.New("number", schematest.WithFormat("float"), schematest.WithMaximum(0)),
		},
		{
			name:   "float min max",
			exp:    float32(2.7030094),
			schema: schematest.New("number", schematest.WithFormat("float"), schematest.WithMinimum(-5), schematest.WithMaximum(5)),
		},
		{
			name:   "double",
			exp:    540601.8643242136,
			schema: schematest.New("number", schematest.WithFormat("double")),
		},
		{
			name:   "double min",
			exp:    770303.2291527851,
			schema: schematest.New("number", schematest.WithFormat("double"), schematest.WithMinimum(10)),
		},
		{
			name:   "double max",
			exp:    -229699.06783789318,
			schema: schematest.New("number", schematest.WithFormat("double"), schematest.WithMaximum(0)),
		},
		{
			name:   "double min max",
			exp:    2.703009321621068,
			schema: schematest.New("number", schematest.WithFormat("double"), schematest.WithMinimum(-5), schematest.WithMaximum(5)),
		},
		{
			name:   "example",
			exp:    1.0,
			schema: schematest.New("number", schematest.WithFormat("double"), schematest.WithExample(1)),
		},
		{
			name:   "examples",
			exp:    7.0,
			schema: schematest.New("number", schematest.WithFormat("double"), schematest.WithExamples(5, 6, 7)),
		},
		{
			name:   "examples over example",
			exp:    7.0,
			schema: schematest.New("number", schematest.WithFormat("double"), schematest.WithExample(1), schematest.WithExamples(5, 6, 7)),
		},
		{
			name:   "enum",
			exp:    1,
			schema: schematest.New("number", schematest.WithFormat("double"), schematest.WithEnumValues(1, 2, 3, 4)),
		},
		{
			name:   "exclusive minimum",
			exp:    0.25406018643242156,
			schema: schematest.New("number", schematest.WithFormat("double"), schematest.WithExclusiveMinimum(0.1), schematest.WithMaximum(0.3)),
		},
		{
			name:   "exclusive maximum",
			exp:    0.28851504660810456,
			schema: schematest.New("number", schematest.WithFormat("double"), schematest.WithMinimum(0.25), schematest.WithExclusiveMaximum(0.3)),
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.Seed(11)

			o, err := schema.CreateValue(data.schema)
			require.NoError(t, err)
			require.InDelta(t, data.exp, o, 0.000001)
		})
	}
}

func TestGeneratorArray(t *testing.T) {
	testcases := []struct {
		name   string
		exp    interface{}
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
	}{
		{
			name: "int32",
			schema: schematest.New("array",
				schematest.WithItems("integer", schematest.WithFormat("int32"), schematest.WithMinimum(0), schematest.WithMaximum(10)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int32(10), int32(6), int32(2), int32(3), int32(9)}, i)
			},
		},
		{
			name: "min items",
			schema: schematest.New("array", schematest.WithMinItems(5),
				schematest.WithItems("integer", schematest.WithFormat("int32"), schematest.WithMinimum(0), schematest.WithMaximum(10)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int32(10), int32(6), int32(2), int32(3), int32(9), int32(0), int32(2), int32(6), int32(9), int32(7)}, i)
			},
		},
		{
			name: "min & max items",
			schema: schematest.New("array", schematest.WithMinItems(5), schematest.WithMaxItems(10),
				schematest.WithItems("integer", schematest.WithFormat("int32"), schematest.WithMinimum(0), schematest.WithMaximum(10)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int32(10), int32(6), int32(2), int32(3), int32(9), int32(0), int32(2), int32(6), int32(9), int32(7)}, i)
			},
		},
		{
			name: "unique items",
			schema: schematest.New("array", schematest.WithMinItems(5), schematest.WithMaxItems(10), schematest.WithUniqueItems(true),
				schematest.WithItems("integer", schematest.WithFormat("int32"), schematest.WithMinimum(0), schematest.WithMaximum(10)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int32(10), int32(6), int32(2), int32(3), int32(9), int32(0), int32(7), int32(8), int32(1), int32(5)}, i)
			},
		},
		{
			name: "unique and shuffle items",
			schema: schematest.New("array", schematest.WithMinItems(2), schematest.WithMaxItems(5), schematest.WithUniqueItems(true), schematest.WithShuffleItems(),
				schematest.WithItems("integer", schematest.WithFormat("int32"), schematest.WithMinimum(0), schematest.WithMaximum(10)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int32(6), int32(10)}, i)
			},
		},
		{
			name: "enum ignores items config",
			schema: schematest.New("array", schematest.WithMinItems(5), schematest.WithMaxItems(10), schematest.WithUniqueItems(true),
				schematest.WithEnumValues([]interface{}{1, 2, 3}, []interface{}{3, 2, 1}),
				schematest.WithItems("integer", schematest.WithFormat("int32"), schematest.WithMinimum(0), schematest.WithMaximum(3)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{1, 2, 3}, i)
			},
		},
		{
			name: "unique items with error",
			schema: schematest.New("array", schematest.WithMinItems(5), schematest.WithMaxItems(10), schematest.WithUniqueItems(true),
				schematest.WithItems("integer", schematest.WithMinimum(0), schematest.WithMaximum(3)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "failed to generate valid array: reached attempt limit (10) caused by: cannot fill array with unique items")
			},
		},
		{
			name: "unique items with enum",
			schema: schematest.New("array", schematest.WithMinItems(5), schematest.WithMaxItems(10), schematest.WithUniqueItems(true),
				schematest.WithItems("integer",
					schematest.WithFormat("int32"),
					schematest.WithEnumValues(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{10, 1, 2, 3, 4, 5, 6, 7, 8, 9}, i)
			},
		},
		{
			name: "unique items with enum and shuffle",
			schema: schematest.New("array", schematest.WithMinItems(5), schematest.WithMaxItems(10), schematest.WithUniqueItems(true), schematest.WithShuffleItems(),
				schematest.WithItems("integer", schematest.WithFormat("int32"), schematest.WithEnumValues(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{1, 5, 2, 8, 9, 4, 3, 6, 7, 10}, i)
			},
		},
		{
			name:   "items not defined",
			schema: schematest.New("array"),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				a := i.([]interface{})
				require.Equal(t, "nsyx7", a[0])
				require.InDelta(t, 824801.9947984695, a[1], 0.000001)
				require.Equal(t, int64(-342586), a[2])
				require.Equal(t, "hyaaKAQyB", a[3])
				require.InDelta(t, -916030.8296825297, a[4], 0.000001)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(11)

			o, err := schema.CreateValue(tc.schema)
			tc.test(t, o, err)
		})
	}
}

func TestGeneratorObject(t *testing.T) {
	testdata := []struct {
		name   string
		exp    map[string]interface{}
		schema *schema.Schema
	}{
		{
			name: "simple",
			exp:  map[string]interface{}{"id": int32(89589)},
			schema: schematest.New("object",
				schematest.WithProperty("id", schematest.New("integer", schematest.WithFormat("int32"))),
				schematest.WithRequired("id"),
			),
		},
		{
			name: "more fields",
			exp:  map[string]interface{}{"date": "2030-03-07", "id": int32(89589)},
			schema: schematest.New("object",
				schematest.WithProperty("id", schematest.New("integer", schematest.WithFormat("int32"))),
				schematest.WithProperty("date", schematest.New("string", schematest.WithFormat("date"))),
				schematest.WithRequired("id", "date"),
			),
		},
		{
			name: "nested",
			exp:  map[string]interface{}{"nested": map[string]interface{}{"date": "2030-03-07", "id": int32(89589)}},
			schema: schematest.New("object",
				schematest.WithProperty("nested", schematest.New("object",
					schematest.WithProperty("id", schematest.New("integer", schematest.WithFormat("int32"))),
					schematest.WithProperty("date", schematest.New("string", schematest.WithFormat("date"))),
					schematest.WithRequired("id", "date"),
				),
				),
				schematest.WithRequired("nested"),
			),
		},
		{
			name: "dictionary",
			exp:  map[string]interface{}{"body": "1fpidf", "class": "yqD", "doctor": "t6ckaieGDffxcd", "fear": "TI5ydf yByPS<qb", "harm": "WDmJn", "pack": "Paitucts2mXR5eZ", "problem": "Qzy", "trip": "mWmsMMblIz"},
			schema: schematest.New("object",
				schematest.WithAdditionalProperties(schematest.New("string"))),
		},
		{
			name: "with property _metadata",
			exp:  map[string]interface{}{"_metadata": int64(791768)},
			schema: schematest.New("object",
				schematest.WithProperty("_metadata", schematest.New("integer", schematest.WithFormat("int64"))),
				schematest.WithRequired("_metadata"),
			),
		},
		{
			name: "with property address as any",
			exp:  map[string]interface{}{"address": map[string]interface{}{"address": "125 East Routemouth, North Las Vegas, South Dakota 17999", "city": "North Las Vegas", "country": "Isle of Man", "latitude": -79.948308, "longitude": -60.019628, "state": "South Dakota", "street": "125 East Routemouth", "zip": "17999"}},
			schema: schematest.New("object",
				schematest.WithProperty("address", schematest.New("")),
				schematest.WithRequired("address"),
			),
		},
		{
			name: "using XML name",
			exp:  map[string]any{"name": "Leah Martinez"},
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithXml(&schema.Xml{Name: "person"}),
			),
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.Seed(11)
			generator.Seed(11)

			v, err := schema.CreateValue(data.schema)
			require.NoError(t, err)
			require.Equal(t, data.exp, v)
		})
	}
}

func TestGenerator_AnyOf(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "array any of",
			f: func(t *testing.T) {
				s := schematest.New("array",
					schematest.WithMinItems(1),
					schematest.WithMaxItems(3),
					schematest.WithItems("",
						schematest.Any(
							schematest.New("object",
								schematest.WithProperty("foo", schematest.New("string")),
								schematest.WithRequired("foo"),
							),

							schematest.New("object",
								schematest.WithProperty("bar",
									schematest.New("integer",
										schematest.WithMinimum(0),
										schematest.WithMaximum(5))),
								schematest.WithRequired("bar"),
							),
						),
					),
				)
				o, err := schema.CreateValue(s)
				require.NoError(t, err)
				b, err := json.Marshal(o)
				require.NoError(t, err)
				require.Equal(t, `[{"foo":"nsyx7"},{"foo":""},{"foo":"IkhyaaKAQyByP"}]`, string(b))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)
			tc.f(t)
		})
	}
}

func TestGenerator_AllOf(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Schema
		test   func(t *testing.T, result interface{}, err error)
	}{
		{
			name: "all of",
			schema: schematest.New("", schematest.AllOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("string")),
					schematest.WithRequired("foo"),
				),
				schematest.New("object",
					schematest.WithProperty("bar", schematest.New("number")),
					schematest.WithRequired("bar"),
				),
			)),
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				m := result.(map[string]any)
				require.Equal(t, m["foo"], "fnsy")
				require.InDelta(t, 897230.3868030173, m["bar"], 0.000001)
			},
		},
		{
			name: "one is null",
			schema: schematest.NewAllOf(
				nil,
				schematest.New("object",
					schematest.WithProperty("bar", schematest.New("number")),
					schematest.WithRequired("bar"),
				),
			),
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				m := result.(map[string]any)
				require.InDelta(t, 540601.8643242136, m["bar"], 0.000001)
			},
		},
		{
			name: "one reference value is null",
			schema: schematest.NewAllOfRefs(
				nil,
				schematest.New("object",
					schematest.WithProperty("bar", schematest.New("number")),
					schematest.WithRequired("bar"),
				),
			),
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				m := result.(map[string]any)
				require.InDelta(t, 540601.8643242136, m["bar"], 0.000001)
			},
		},
		{
			name: "with integer type",
			schema: schematest.New("", schematest.AllOf(
				schematest.New("integer"),
				schematest.New("object",
					schematest.WithProperty("bar", schematest.New("number")),
					schematest.WithRequired("bar"),
				),
			)),
			test: func(t *testing.T, result interface{}, err error) {
				require.EqualError(t, err, "generate random data for schema failed: no shared types found: integer and object")
				require.Nil(t, result)
			},
		},
		{
			name: "one gets error",
			schema: schematest.New("", schematest.AllOf(
				schematest.New("object", schematest.WithProperty("a",
					schematest.New("array",
						schematest.WithUniqueItems(true),
						schematest.WithMinItems(5),
						schematest.WithItems(
							"integer",
							schematest.WithMinimum(0),
							schematest.WithMaximum(3),
						)),
				),
					schematest.WithRequired("a"),
				),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("number"))),
			)),
			test: func(t *testing.T, result interface{}, err error) {
				require.EqualError(t, err, "failed to generate valid object: reached attempt limit (10) caused by: failed to generate valid array: reached attempt limit (10) caused by: cannot fill array with unique items")
				require.Nil(t, result)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)

			o, err := schema.CreateValue(tc.schema)

			tc.test(t, o, err)
		})
	}
}

func TestGenerator_OneOf(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Schema
		test   func(t *testing.T, result interface{}, err error)
	}{
		{
			name: "one of",
			schema: schematest.New("", schematest.OneOf(
				schematest.New("number", schematest.WithMinimum(10)),
				schematest.New("number", schematest.WithMinimum(0), schematest.WithMaximum(9)),
			)),
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 985963.0664648871, result)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)

			o, err := schema.CreateValue(tc.schema)

			tc.test(t, o, err)
		})
	}
}

func TestGenerator_Recursions(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"recursion depth 1",
			func(t *testing.T) {
				s := schematest.New("object", schematest.And("null"), schematest.WithRequired("foo"))
				props := &schema.Schemas{}
				props.Set("foo", s)
				s.Properties = props

				result, err := schema.CreateValue(s)
				require.NoError(t, err)

				b, err := json.Marshal(result)
				require.NoError(t, err)
				require.Equal(t, `{"foo":null}`, string(b))
			},
		},
		{
			"recursion across two objects depth 1",
			func(t *testing.T) {
				child := schematest.New("object", schematest.WithRequired("foo"))
				s := schematest.New("object",
					schematest.IsNullable(true),
					schematest.WithProperty("bar", child),
					schematest.WithRequired("bar"),
				)
				props := &schema.Schemas{}
				props.Set("foo", s)
				child.Properties = props

				result, err := schema.CreateValue(s)
				require.NoError(t, err)
				require.NotNil(t, result)

				b, err := json.Marshal(result)
				require.NoError(t, err)
				require.Equal(t, `{"bar":{"foo":null}}`, string(b))
			},
		},
		{
			"array",
			func(t *testing.T) {
				obj := schematest.New("object", schematest.And("null"))
				props := &schema.Schemas{}
				props.Set("foo", obj)
				obj.Properties = props
				array := schematest.New("array")
				array.Items = obj
				minItems := 2
				array.MinItems = &minItems

				o, err := schema.CreateValue(array)
				require.NoError(t, err)
				require.NotNil(t, o)
				a := o.([]interface{})
				require.NotNil(t, a[1])
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)
			tc.f(t)
		})
	}

}

func TestGeneratorNullable(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Schema
		seed   int64
		test   func(t *testing.T, exp interface{}, err error)
	}{
		{
			name:   "nullable string",
			schema: schematest.New("string", schematest.IsNullable(true)),
			seed:   49,
			test: func(t *testing.T, exp interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, exp)
			},
		},
		{
			name:   "nullable int",
			schema: schematest.New("integer", schematest.IsNullable(true)),
			seed:   49,
			test: func(t *testing.T, exp interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, exp)
			},
		},
		{
			name:   "nullable number",
			schema: schematest.New("number", schematest.IsNullable(true)),
			seed:   49,
			test: func(t *testing.T, exp interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, exp)
			},
		},
		{
			name: "nullable object",
			schema: schematest.New("object",
				schematest.IsNullable(true),
				schematest.WithProperty("foo", schematest.New("string"))),
			seed: 49,
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, result)
			},
		},
		{
			name: "nullable property",
			schema: schematest.New("object",
				schematest.WithProperty("foo", schematest.New("string", schematest.IsNullable(true))),
				schematest.WithRequired("foo"),
			),
			seed: 49,
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)

				b, err := json.Marshal(result)
				require.NoError(t, err)
				require.Equal(t, `{"foo":null}`, string(b))
			},
		},
		{
			name: "nullable array",
			schema: schematest.New("array",
				schematest.IsNullable(true),
				schematest.WithItems("string")),
			seed: 49,
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, result)
			},
		},
		{
			name: "nullable array item",
			schema: schematest.New("array",
				schematest.WithMinItems(1),
				schematest.WithItems("string", schematest.IsNullable(true))),
			seed: 49,
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				arr := result.([]interface{})
				require.Nil(t, arr[0])
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(tc.seed)

			o, err := schema.CreateValue(tc.schema)
			tc.test(t, o, err)
		})
	}
}

func _TestFindSeed(t *testing.T) {
	i := int64(0)
	for {
		gofakeit.Seed(i)

		o, _ := schema.CreateValue(schematest.New("array",
			schematest.WithItems("string", schematest.IsNullable(true))))

		//require.NotNil(t, o, "seed %v", i)
		if o == nil {
			continue
		}

		for _, v := range o.([]interface{}) {
			if v == nil {
				require.NotNil(t, v, "seed %v", i)
				return
			}
		}

		i++
	}
}

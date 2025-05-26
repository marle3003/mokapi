package schema_test

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/schema/json/generator"
	jsonSchema "mokapi/schema/json/schema"
	"testing"
)

func toFloatP(f float64) *float64 { return &f }
func toIntP(i int) *int           { return &i }
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
				require.Equal(t, "", v)
			},
		},
		{
			name:   "empty schema",
			schema: schematest.New(""),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "", v)
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
		test   func(v interface{}, err error)
	}{
		{
			name:   "string",
			schema: schematest.New("string"),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "XidZuoWq ", v)
			},
		},
		{
			name:   "by pattern",
			schema: schematest.New("string", schematest.WithPattern("^\\d{3}-\\d{2}-\\d{4}$")),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "013-64-5994", v)
			},
		},
		{
			name:   "date",
			schema: schematest.New("string", schematest.WithFormat("date")),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2035-01-24", v)
			},
		},
		{
			name:   "date-time",
			schema: schematest.New("string", schematest.WithFormat("date-time")),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2035-01-24T13:00:35Z", v)
			},
		},
		{
			name:   "password",
			schema: schematest.New("string", schematest.WithFormat("password")),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "sX!54wZ8!69V", v)
			},
		},
		{
			name:   "email",
			schema: schematest.New("string", schematest.WithFormat("email")),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "markusmoen@pagac.net", v)
			},
		},
		{
			name:   "uuid",
			schema: schematest.New("string", schematest.WithFormat("uuid")),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "98173564-6619-4557-888e-65b16bb5def5", v)
			},
		},
		{
			name:   "url",
			schema: schematest.New("string", schematest.WithFormat("{url}")),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "https://www.dynamiciterate.name/target/seamless", v)
			},
		},
		{
			name:   "hostname",
			schema: schematest.New("string", schematest.WithFormat("hostname")),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "centraltarget.biz", v)
			},
		},
		{
			name:   "ipv4",
			schema: schematest.New("string", schematest.WithFormat("ipv4")),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "152.23.53.100", v)
			},
		},
		{
			name:   "ipv6",
			schema: schematest.New("string", schematest.WithFormat("ipv6")),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "8898:ee17:bc35:9064:5866:d019:3b95:7857", v)
			},
		},
		{
			name:   "beername",
			schema: schematest.New("string", schematest.WithFormat("{beername}")),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Duvel", v)
			},
		},
		{
			name:   "address",
			schema: schematest.New("string", schematest.WithFormat("{zip} {city}")),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "13645 Houston", v)
			},
		},
		{
			name:   "uri",
			schema: schematest.New("string", schematest.WithFormat("uri")),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "https://www.dynamiciterate.name/target/seamless", v)
			},
		},
		{
			name:   "minLength",
			schema: schematest.New("string", schematest.WithMinLength(25)),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "XidZuoWq vY5elXhlD4ezlYehCIA0OSwlV", v)
			},
		},
		{
			name:   "maxLength",
			schema: schematest.New("string", schematest.WithMaxLength(4)),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "", v)
			},
		},
		{
			name:   "maxLength",
			schema: schematest.New("string", schematest.WithMaxLength(12)),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "XidZuoWq vY", v)
			},
		},
		{
			name:   "minLength with maxLength",
			schema: schematest.New("string", schematest.WithMinLength(3), schematest.WithMaxLength(6)),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "XidZ", v)
			},
		},
		{
			name:   "minLength equals maxLength",
			schema: schematest.New("string", schematest.WithMinLength(4), schematest.WithMaxLength(4)),
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "sXPO", v)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)
			generator.Seed(11)

			v, err := schema.CreateValue(tc.schema)
			tc.test(v, err)
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
			exp:    true,
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
				require.Equal(t, int32(-1072427943), i)
			},
		},
		{
			name:   "int32 min",
			schema: schematest.New("integer", schematest.WithFormat("int32"), schematest.WithMinimum(10)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(196446369), i)
			},
		},
		{
			name:   "int32 max",
			schema: schematest.New("integer", schematest.WithFormat("int32"), schematest.WithMaximum(0)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(-1951037288), i)
			},
		},
		{
			name:   "int32 min max",
			schema: schematest.New("integer", schematest.WithFormat("int32"), schematest.WithMinimum(-5), schematest.WithMaximum(5)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(-4), i)
			},
		},
		{
			name:   "int64",
			schema: schematest.New("integer", schematest.WithFormat("int64")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-8379641344161477543), i)
			},
		},
		{
			name:   "int64 min",
			schema: schematest.New("integer", schematest.WithFormat("int64"), schematest.WithMinimum(10)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(843730692693298304), i)
			},
		},
		{
			name:   "int64 max",
			schema: schematest.New("integer", schematest.WithFormat("int64"), schematest.WithMaximum(0)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-8379641344161477632), i)
			},
		},
		{
			name:   "int64 min max",
			schema: schematest.New("integer", schematest.WithFormat("int64"), schematest.WithMinimum(-5), schematest.WithMaximum(5)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-4), i)
			},
		},
		{
			name:   "int64 min max positive",
			schema: schematest.New("integer", schematest.WithFormat("int64"), schematest.WithMinimum(4), schematest.WithMaximum(10)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(5), i)
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
			exp:    float32(3.1128167e+37),
			schema: schematest.New("number", schematest.WithFormat("float")),
		},
		{
			name:   "float min",
			exp:    float32(3.1128167e+37),
			schema: schematest.New("number", schematest.WithFormat("float"), schematest.WithMinimum(10)),
		},
		{
			name:   "float max",
			exp:    float32(-3.0915418e+38),
			schema: schematest.New("number", schematest.WithFormat("float"), schematest.WithMaximum(0)),
		},
		{
			name:   "float min max",
			exp:    float32(-4.0852256),
			schema: schematest.New("number", schematest.WithFormat("float"), schematest.WithMinimum(-5), schematest.WithMaximum(5)),
		},
		{
			name:   "double",
			exp:    1.644484108270445e+307,
			schema: schematest.New("number", schematest.WithFormat("double")),
		},
		{
			name:   "double min",
			exp:    1.644484108270445e+307,
			schema: schematest.New("number", schematest.WithFormat("double"), schematest.WithMinimum(10)),
		},
		{
			name:   "double max",
			exp:    -1.6332447240352712e+308,
			schema: schematest.New("number", schematest.WithFormat("double"), schematest.WithMaximum(0)),
		},
		{
			name:   "double min max",
			exp:    -4.085225349989226,
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
			exp:    2,
			schema: schematest.New("number", schematest.WithFormat("double"), schematest.WithEnumValues(1, 2, 3, 4)),
		},
		{
			name:   "exclusive minimum",
			exp:    0.11829549300021638,
			schema: schematest.New("number", schematest.WithFormat("double"), schematest.WithExclusiveMinimum(0.1), schematest.WithMaximum(0.3)),
		},
		{
			name:   "exclusive maximum",
			exp:    0.25457387325005376,
			schema: schematest.New("number", schematest.WithFormat("double"), schematest.WithMinimum(0.25), schematest.WithExclusiveMaximum(0.3)),
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
				require.Equal(t, []interface{}{int32(8), int32(8), int32(6), int32(7), int32(1)}, i)
			},
		},
		{
			name: "min items",
			schema: schematest.New("array", schematest.WithMinItems(5),
				schematest.WithItems("integer", schematest.WithFormat("int32"), schematest.WithMinimum(0), schematest.WithMaximum(10)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int32(1), int32(8), int32(8), int32(6), int32(7)}, i)
			},
		},
		{
			name: "min & max items",
			schema: schematest.New("array", schematest.WithMinItems(5), schematest.WithMaxItems(10),
				schematest.WithItems("integer", schematest.WithFormat("int32"), schematest.WithMinimum(0), schematest.WithMaximum(10)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int32(8), int32(8), int32(6), int32(7), int32(1), int32(8), int32(9), int32(5), int32(3), int32(1)}, i)
			},
		},
		{
			name: "unique items",
			schema: schematest.New("array", schematest.WithMinItems(5), schematest.WithMaxItems(10), schematest.WithUniqueItems(),
				schematest.WithItems("integer", schematest.WithFormat("int32"), schematest.WithMinimum(0), schematest.WithMaximum(10)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int32(8), int32(6), int32(7), int32(1), int32(9), int32(5), int32(3), int32(2), int32(4), int32(10)}, i)
			},
		},
		{
			name: "unique and shuffle items",
			schema: schematest.New("array", schematest.WithMinItems(2), schematest.WithMaxItems(5), schematest.WithUniqueItems(), schematest.WithShuffleItems(),
				schematest.WithItems("integer", schematest.WithFormat("int32"), schematest.WithMinimum(0), schematest.WithMaximum(10)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int32(7), int32(6), int32(8)}, i)
			},
		},
		{
			name: "enum ignores items config",
			schema: schematest.New("array", schematest.WithMinItems(5), schematest.WithMaxItems(10), schematest.WithUniqueItems(),
				schematest.WithEnumValues([]interface{}{1, 2, 3}, []interface{}{3, 2, 1}),
				schematest.WithItems("integer", schematest.WithFormat("int32"), schematest.WithMinimum(0), schematest.WithMaximum(3)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{3, 2, 1}, i)
			},
		},
		{
			name: "unique items with error",
			schema: schematest.New("array", schematest.WithMinItems(5), schematest.WithMaxItems(10), schematest.WithUniqueItems(),
				schematest.WithItems("integer", schematest.WithFormat("int32"), schematest.WithMinimum(0), schematest.WithMaximum(3)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "can not fill array with unique items: schema type=array minItems=5 maxItems=10 unique-items items=schema type=integer format=int32 minimum=0 maximum=3")
			},
		},
		{
			name: "unique items with enum",
			schema: schematest.New("array", schematest.WithMinItems(5), schematest.WithMaxItems(10), schematest.WithUniqueItems(),
				schematest.WithItems("integer",
					schematest.WithFormat("int32"),
					schematest.WithEnumValues(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{7, 8, 9, 10, 1}, i)
			},
		},
		{
			name: "unique items with enum and shuffle",
			schema: schematest.New("array", schematest.WithMinItems(5), schematest.WithMaxItems(10), schematest.WithUniqueItems(), schematest.WithShuffleItems(),
				schematest.WithItems("integer", schematest.WithFormat("int32"), schematest.WithEnumValues(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{1, 8, 9, 10, 7}, i)
			},
		},
		{
			name:   "items not defined",
			schema: schematest.New("array"),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{true, int64(6224634831868504800), "ZuoWq vY5elXhlD", []interface{}{2.3090412168364615e+307, "lYehCIA", map[string]interface{}{"caravan": true, "hail": 2.536044080333601e+307, "mob": int64(-287411453310397474), "scale": true}, false}, false}, i)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.SetGlobalFaker(gofakeit.New(11))
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
			exp:  map[string]interface{}{"id": int32(98266)},
			schema: schematest.New("object",
				schematest.WithProperty("id", schematest.New("integer", schematest.WithFormat("int32"))),
				schematest.WithRequired("id"),
			),
		},
		{
			name: "more fields",
			exp:  map[string]interface{}{"id": int32(98266), "date": "2038-12-28"},
			schema: schematest.New("object",
				schematest.WithProperty("id", schematest.New("integer", schematest.WithFormat("int32"))),
				schematest.WithProperty("date", schematest.New("string", schematest.WithFormat("date"))),
				schematest.WithRequired("id", "date"),
			),
		},
		{
			name: "nested",
			exp:  map[string]interface{}{"nested": map[string]interface{}{"id": int32(98266), "date": "2038-12-28"}},
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
			exp:  map[string]interface{}{"bunch": "Pevuwy", "gang": "", "growth": "NrLJgmr9arW", "hall": "JKqGj", "woman": "x?vY5elXhlD4ez"},
			schema: schematest.New("object",
				schematest.WithAdditionalProperties(schematest.New("string"))),
		},
		{
			name:   "no fields defined",
			exp:    map[string]interface{}{"bunch": int64(8673350504153079445), "child": int64(5224568207835308195), "gang": 5.544677937412537e+307, "growth": int64(-8487131363427706431), "hall": 1.0615919996637124e+308, "shower": true, "uncle": int64(-7273558372469573415), "woman": int64(6892487422858870876)},
			schema: schematest.New("object"),
		},
		{
			name: "with property _metadata",
			exp:  map[string]interface{}{"_metadata": int64(-8379641344161477543)},
			schema: schematest.New("object",
				schematest.WithProperty("_metadata", schematest.New("integer", schematest.WithFormat("int64"))),
				schematest.WithRequired("_metadata"),
			),
		},
		{
			name: "with property address as any",
			exp:  map[string]interface{}{"address": map[string]interface{}{"address": "364 Unionsville, Norfolk, Ohio 99536", "city": "Norfolk", "country": "Lesotho", "latitude": 88.792592, "longitude": 174.504681, "state": "Ohio", "street": "364 Unionsville", "zip": "99536"}},
			schema: schematest.New("object",
				schematest.WithProperty("address", schematest.New("")),
				schematest.WithRequired("address"),
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
				require.Equal(t, `[{"bar":4},{"bar":3}]`, string(b))
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
				require.Equal(t, map[string]interface{}{"foo": "XidZuoWq ", "bar": 5.58584061532191e+307}, result)
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
				require.Equal(t, map[string]interface{}{"bar": 1.644484108270445e+307}, result)
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
				require.Equal(t, map[string]interface{}{"bar": 1.644484108270445e+307}, result)
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
				require.EqualError(t, err, "generate random data for schema failed: all of schema type=integer, schema type=object properties=[bar] required=[bar]: no shared types found")
				require.Nil(t, result)
			},
		},
		{
			name: "one gets error",
			schema: schematest.New("", schematest.AllOf(
				schematest.New("object", schematest.WithProperty("a",
					schematest.New("array",
						schematest.WithUniqueItems(),
						schematest.WithItems(
							"integer",
							schematest.WithMinItems(5),
							schematest.WithMinimum(0),
							schematest.WithMaximum(3),
						)),
				),
					schematest.WithRequired("a"),
				),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("number"))),
			)),
			test: func(t *testing.T, result interface{}, err error) {
				require.EqualError(t, err, "generate random data for schema failed: schema type=object properties=[a] required=[a]: can not fill array with unique items: schema type=array unique-items items=schema type=integer minimum=0 maximum=3 minItems=5")
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
				require.Equal(t, 7.244365552867502, result)
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
			seed:   43,
			test: func(t *testing.T, exp interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, exp)
			},
		},
		{
			name:   "nullable int",
			schema: schematest.New("integer", schematest.IsNullable(true)),
			seed:   43,
			test: func(t *testing.T, exp interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, exp)
			},
		},
		{
			name:   "nullable number",
			schema: schematest.New("number", schematest.IsNullable(true)),
			seed:   43,
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
			seed: 43,
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
			seed: 43,
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
			seed: 9,
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, result)
			},
		},
		{
			name: "nullable array item",
			schema: schematest.New("array",
				schematest.WithItems("string", schematest.IsNullable(true))),
			seed: 20,
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

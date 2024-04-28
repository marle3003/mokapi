package schema_test

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/json/generator"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
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
				require.Equal(t,
					map[string]interface{}{"buckles": 1.414946145709964e+308, "kitchen": true, "problem": 1.410477402203964e+308, "sand": true, "sock": 2.3090412168364615e+307, "thing": "Y5elX", "tribe": int64(9082579350789565885)},
					v)
			},
		},
		{
			name:   "empty schema",
			schema: &schema.Schema{},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{"buckles": 1.414946145709964e+308, "kitchen": true, "problem": 1.410477402203964e+308, "sand": true, "sock": 2.3090412168364615e+307, "thing": "Y5elX", "tribe": int64(9082579350789565885)},
					v)
			},
		},
		{
			name:   "invalid type",
			schema: &schema.Schema{Type: "foobar"},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "unsupported schema: schema type=foobar")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)

			v, err := schema.CreateValue(&schema.Ref{Value: tc.schema})
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
			schema: &schema.Schema{Type: "string"},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "XidZuoWq ", v)
			},
		},
		{
			name:   "by pattern",
			schema: &schema.Schema{Type: "string", Pattern: "^\\d{3}-\\d{2}-\\d{4}$"},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "013-64-5994", v)
			},
		},
		{
			name:   "date",
			schema: &schema.Schema{Type: "string", Format: "date"},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "1915-01-24", v)
			},
		},
		{
			name:   "date-time",
			schema: &schema.Schema{Type: "string", Format: "date-time"},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "1915-01-24T13:00:35Z", v)
			},
		},
		{
			name:   "password",
			schema: &schema.Schema{Type: "string", Format: "password"},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "sX!54wZ8!69V", v)
			},
		},
		{
			name:   "email",
			schema: &schema.Schema{Type: "string", Format: "email"},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "markusmoen@pagac.net", v)
			},
		},
		{
			name:   "uuid",
			schema: &schema.Schema{Type: "string", Format: "uuid"},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "98173564-6619-4557-888e-65b16bb5def5", v)
			},
		},
		{
			name:   "url",
			schema: &schema.Schema{Type: "string", Format: "{url}"},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "https://www.dynamiciterate.name/target/seamless", v)
			},
		},
		{
			name:   "hostname",
			schema: &schema.Schema{Type: "string", Format: "hostname"},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "centraltarget.biz", v)
			},
		},
		{
			name:   "ipv4",
			schema: &schema.Schema{Type: "string", Format: "ipv4"},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "152.23.53.100", v)
			},
		},
		{
			name:   "ipv6",
			schema: &schema.Schema{Type: "string", Format: "ipv6"},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "8898:ee17:bc35:9064:5866:d019:3b95:7857", v)
			},
		},
		{
			name:   "beername",
			schema: &schema.Schema{Type: "string", Format: "{beername}"},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Duvel", v)
			},
		},
		{
			name:   "address",
			schema: &schema.Schema{Type: "string", Format: "{zip} {city}"},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "13645 Houston", v)
			},
		},
		{
			name:   "uri",
			schema: &schema.Schema{Type: "string", Format: "uri"},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "https://www.dynamiciterate.name/target/seamless", v)
			},
		},
		{
			name:   "minLength",
			schema: &schema.Schema{Type: "string", MinLength: toIntP(25)},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "XidZuoWq vY5elXhlD4ezlYe", v)
			},
		},
		{
			name:   "maxLength",
			schema: &schema.Schema{Type: "string", MaxLength: toIntP(4)},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "", v)
			},
		},
		{
			name:   "maxLength",
			schema: &schema.Schema{Type: "string", MaxLength: toIntP(12)},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "XidZuoWq vY", v)
			},
		},
		{
			name:   "minLength with maxLength",
			schema: &schema.Schema{Type: "string", MinLength: toIntP(3), MaxLength: toIntP(6)},
			test: func(v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "XidZ", v)
			},
		},
		{
			name:   "minLength equals maxLength",
			schema: &schema.Schema{Type: "string", MinLength: toIntP(4), MaxLength: toIntP(4)},
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

			v, err := schema.CreateValue(&schema.Ref{Value: tc.schema})
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
			"false",
			true,
			&schema.Schema{Type: "boolean"},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.Seed(11)

			o, err := schema.CreateValue(&schema.Ref{Value: data.schema})
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
			schema: &schema.Schema{Type: "integer", Format: "int32"},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(-1072427943), i)
			},
		},
		{
			name:   "int32 min",
			schema: &schema.Schema{Type: "integer", Format: "int32", Minimum: toFloatP(10)},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(196446384), i)
			},
		},
		{
			name: "int32 max",

			schema: &schema.Schema{Type: "integer", Format: "int32", Maximum: toFloatP(0)},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(-1951037312), i)
			},
		},
		{
			name:   "int32 min max",
			schema: &schema.Schema{Type: "integer", Format: "int32", Minimum: toFloatP(-5), Maximum: toFloatP(5)},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(-4), i)
			},
		},
		{
			name:   "int64",
			schema: &schema.Schema{Type: "integer", Format: "int64"},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-8379641344161477543), i)
			},
		},
		{
			name:   "int64 min",
			schema: &schema.Schema{Type: "integer", Format: "int64", Minimum: toFloatP(10)},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(843730692693298304), i)
			},
		},
		{
			name:   "int64 max",
			schema: &schema.Schema{Type: "integer", Format: "int64", Maximum: toFloatP(0)},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-8379641344161477632), i)
			},
		},
		{
			name:   "int64 min max",
			schema: &schema.Schema{Type: "integer", Format: "int64", Minimum: toFloatP(-5), Maximum: toFloatP(5)},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-4), i)
			},
		},
		{
			name:   "int64 min max positive",
			schema: &schema.Schema{Type: "integer", Format: "int64", Minimum: toFloatP(4), Maximum: toFloatP(10)},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(5), i)
			},
		},
		{
			name: "int64 min max positive exclusive",
			schema: &schema.Schema{
				Type:             "integer",
				Format:           "int64",
				Minimum:          toFloatP(3),
				Maximum:          toFloatP(5),
				ExclusiveMinimum: toBoolP(true),
				ExclusiveMaximum: toBoolP(true),
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(4), i)
			},
		},
		{
			name: "int64 min max positive exclusive but error",
			schema: &schema.Schema{
				Type:             "integer",
				Format:           "int64",
				Minimum:          toFloatP(4),
				Maximum:          toFloatP(5),
				ExclusiveMinimum: toBoolP(true),
				ExclusiveMaximum: toBoolP(true),
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "invalid minimum '5' and maximum '4' in schema type=integer format=int64 exclusiveMinimum=4 exclusiveMaximum=5")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)

			i, err := schema.CreateValue(&schema.Ref{Value: tc.schema})
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
			"float",
			float32(3.1128167e+37),
			&schema.Schema{Type: "number", Format: "float"},
		},
		{
			"float min",
			float32(3.1128167e+37),
			&schema.Schema{Type: "number", Format: "float", Minimum: toFloatP(10)},
		},
		{
			"float max",
			float32(-3.0915418e+38),
			&schema.Schema{Type: "number", Format: "float", Maximum: toFloatP(0)},
		},
		{
			"float min max",
			float32(-4.085225),
			&schema.Schema{Type: "number", Format: "float", Minimum: toFloatP(-5), Maximum: toFloatP(5)},
		},
		{
			"double",
			1.644484108270445e+307,
			&schema.Schema{Type: "number", Format: "double"},
		},
		{
			"double min",
			1.644484108270445e+307,
			&schema.Schema{Type: "number", Format: "double", Minimum: toFloatP(10)},
		},
		{
			"double max",
			-1.6332447240352712e+308,
			&schema.Schema{Type: "number", Format: "double", Maximum: toFloatP(0)},
		},
		{
			"double min max",
			-4.085225349989226,
			&schema.Schema{Type: "number", Format: "double", Minimum: toFloatP(-5), Maximum: toFloatP(5)},
		},
		{
			"example",
			1,
			&schema.Schema{Type: "number", Format: "double", Example: 1},
		},
		{
			"enum",
			2,
			&schema.Schema{Type: "number", Format: "double", Enum: []interface{}{1, 2, 3, 4}},
		},
		{
			"exclusive minimum",
			0.11829549300021638,
			&schema.Schema{Type: "number", Format: "double",
				Minimum: toFloatP(0.1), ExclusiveMinimum: toBoolP(true),
				Maximum: toFloatP(0.3),
			},
		},
		{
			"exclusive maximum",
			0.25457387325005376,
			&schema.Schema{Type: "number", Format: "double",
				Minimum: toFloatP(0.25),
				Maximum: toFloatP(0.3), ExclusiveMaximum: toBoolP(true),
			},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.Seed(11)

			o, err := schema.CreateValue(&schema.Ref{Value: data.schema})
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
			schema: &schema.Schema{Type: "array", Items: &schema.Ref{
				Value: &schema.Schema{
					Type: "integer", Format: "int32", Minimum: toFloatP(0), Maximum: toFloatP(10)}}},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int32(8), int32(8), int32(6), int32(7), int32(1)}, i)
			},
		},
		{
			name: "min items",
			schema: &schema.Schema{Type: "array", MinItems: toIntP(5), Items: &schema.Ref{
				Value: &schema.Schema{
					Type: "integer", Format: "int32", Minimum: toFloatP(0), Maximum: toFloatP(10)}}},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int32(1), int32(8), int32(8), int32(6), int32(7)}, i)
			},
		},
		{
			name: "min & max items",
			schema: &schema.Schema{Type: "array", MinItems: toIntP(5), MaxItems: toIntP(10), Items: &schema.Ref{
				Value: &schema.Schema{
					Type: "integer", Format: "int32", Minimum: toFloatP(0), Maximum: toFloatP(10)}}},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int32(8), int32(8), int32(6), int32(7), int32(1), int32(8), int32(9), int32(5), int32(3), int32(1)}, i)
			},
		},
		{
			name: "unique items",
			schema: &schema.Schema{Type: "array", MinItems: toIntP(5), MaxItems: toIntP(10), UniqueItems: true,
				Items: &schema.Ref{
					Value: &schema.Schema{
						Type: "integer", Format: "int32", Minimum: toFloatP(0), Maximum: toFloatP(10)}}},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int32(8), int32(6), int32(7), int32(1), int32(9), int32(5), int32(3), int32(2), int32(4), int32(10)}, i)
			},
		},
		{
			name: "unique and shuffle items",
			schema: &schema.Schema{Type: "array", MinItems: toIntP(2), MaxItems: toIntP(5), UniqueItems: true,
				Items: &schema.Ref{
					Value: &schema.Schema{
						Type:    "integer",
						Format:  "int32",
						Minimum: toFloatP(0),
						Maximum: toFloatP(10),
					},
				},
				ShuffleItems: true,
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int32(7), int32(6), int32(8)}, i)
			},
		},
		{
			name: "enum ignores items config",
			schema: &schema.Schema{Type: "array", MinItems: toIntP(5), MaxItems: toIntP(10), UniqueItems: true,
				Enum: []interface{}{
					[]interface{}{1, 2, 3},
					[]interface{}{3, 2, 1},
				},
				Items: &schema.Ref{
					Value: &schema.Schema{
						Type: "integer", Format: "int32", Minimum: toFloatP(0), Maximum: toFloatP(3)}}},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{3, 2, 1}, i)
			},
		},
		{
			name: "example",
			schema: &schema.Schema{Type: "array", MinItems: toIntP(5), MaxItems: toIntP(10), UniqueItems: true,
				Example: []interface{}{1, 2, 3},
				Items: &schema.Ref{
					Value: &schema.Schema{
						Type: "integer", Format: "int32", Minimum: toFloatP(0), Maximum: toFloatP(3)}}},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{1, 2, 3}, i)
			},
		},
		{
			name: "unique items with error",
			schema: &schema.Schema{Type: "array", MinItems: toIntP(5), MaxItems: toIntP(10), UniqueItems: true,
				Items: &schema.Ref{
					Value: &schema.Schema{
						Type: "integer", Format: "int32", Minimum: toFloatP(0), Maximum: toFloatP(3)}}},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "can not fill array with unique items: schema type=array minItems=5 maxItems=10 unique-items items=schema type=integer format=int32 minimum=0 maximum=3")
			},
		},
		{
			name: "unique items with enum",
			schema: &schema.Schema{Type: "array", MinItems: toIntP(5), MaxItems: toIntP(10), UniqueItems: true,
				Items: &schema.Ref{
					Value: &schema.Schema{
						Type:   "integer",
						Format: "int32",
						Enum:   []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
					},
				},
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{7, 10, 1, 2, 6, 3, 9, 4, 5, 8}, i)
			},
		},
		{
			name: "unique items with enum and shuffle",
			schema: &schema.Schema{Type: "array", MinItems: toIntP(5), MaxItems: toIntP(10), UniqueItems: true,
				Items: &schema.Ref{
					Value: &schema.Schema{
						Type:   "integer",
						Format: "int32",
						Enum:   []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
					},
				},
				ShuffleItems: true,
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{10, 3, 1, 5, 8, 6, 2, 9, 4, 7}, i)
			},
		},
		{
			name:   "items not defined",
			schema: &schema.Schema{Type: "array"},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"idZ", false, "", []interface{}{}, map[string]interface{}{"shower": 1.3433890851076963e+308}}, i)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.SetGlobalFaker(gofakeit.New(11))
			generator.Seed(11)

			o, err := schema.CreateValue(&schema.Ref{Value: tc.schema})
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
			name:   "simple",
			exp:    map[string]interface{}{"id": 98266},
			schema: schematest.New("object", schematest.WithProperty("id", &schema.Schema{Type: "integer", Format: "int32"})),
		},
		{
			name: "more fields",
			exp:  map[string]interface{}{"id": 98266, "date": "1901-12-28"},
			schema: schematest.New("object",
				schematest.WithProperty("id", schematest.New("integer", schematest.WithFormat("int32"))),
				schematest.WithProperty("date", schematest.New("string", schematest.WithFormat("date")))),
		},
		{
			name: "nested",
			exp:  map[string]interface{}{"nested": map[string]interface{}{"id": 98266, "date": "1901-12-28"}},
			schema: schematest.New("object",
				schematest.WithProperty("nested", schematest.New("object",
					schematest.WithProperty("id", schematest.New("integer", schematest.WithFormat("int32"))),
					schematest.WithProperty("date", schematest.New("string", schematest.WithFormat("date"))),
				),
				),
			),
		},
		{
			name: "dictionary",
			exp:  map[string]interface{}{"effect": "q vY5elXhlD4ez", "gang": "", "problem": "zw", "tribe": "", "way": "1JKqGj", "wisp": "evuwyrNrLJgmr9a"},
			schema: schematest.New("object",
				schematest.WithAdditionalProperties(schematest.New("string"))),
		},
		{
			name:   "no fields defined",
			exp:    map[string]interface{}{"bunch": map[string]interface{}{"shower": 1.3433890851076963e+308}, "gang": []interface{}{false}, "growth": "m", "hall": 1.018301155186648e+308, "woman": []interface{}{}},
			schema: &schema.Schema{Type: "object"},
		},
		{
			name:   "with property _metadata",
			exp:    map[string]interface{}{"_metadata": int64(-8379641344161477543)},
			schema: schematest.New("object", schematest.WithProperty("_metadata", &schema.Schema{Type: "integer", Format: "int64"})),
		},
		{
			name:   "with property address as any",
			exp:    map[string]interface{}{"address": map[string]interface{}{"address": "364 Unionsville, Norfolk, Ohio 99536", "city": "Norfolk", "country": "Lesotho", "latitude": 88.792592, "longitude": 174.504681, "state": "Ohio", "street": "364 Unionsville", "zip": "99536"}},
			schema: schematest.New("object", schematest.WithProperty("address", &schema.Schema{})),
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.Seed(11)
			generator.Seed(11)

			v, err := schema.CreateValue(&schema.Ref{Value: data.schema})
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
								schematest.WithProperty("foo", schematest.New("string"))),
							schematest.New("object",
								schematest.WithProperty("bar",
									schematest.New("integer",
										schematest.WithMinimum(0),
										schematest.WithMaximum(5)))),
						),
					),
				)
				o, err := schema.CreateValue(&schema.Ref{Value: s})
				require.NoError(t, err)
				b, err := json.Marshal(o)
				require.NoError(t, err)
				require.Equal(t, `[{"foo":"idZ"}]`, string(b))
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
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("number"))),
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
				schematest.New("object", schematest.WithProperty("bar", schematest.New("number"))),
			),
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}(map[string]interface{}{"bar": 1.025479772807108e+308, "bunch": map[string]interface{}{"shower": 1.3433890851076963e+308}, "gang": []interface{}{false}, "growth": "m", "hall": 1.018301155186648e+308, "woman": []interface{}{}}), result)
			},
		},
		{
			name: "one reference value is null",
			schema: schematest.NewAllOfRefs(
				&schema.Ref{},
				&schema.Ref{
					Value: schematest.New("object", schematest.WithProperty("bar", schematest.New("number"))),
				},
			),
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"bar": 1.025479772807108e+308, "bunch": map[string]interface{}{"shower": 1.3433890851076963e+308}, "gang": []interface{}{false}, "growth": "m", "hall": 1.018301155186648e+308, "woman": []interface{}{}}, result)
			},
		},
		{
			name: "with integer type",
			schema: schematest.New("", schematest.AllOf(
				schematest.New("integer"),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("number"))),
			)),
			test: func(t *testing.T, result interface{}, err error) {
				require.EqualError(t, err, "allOf expects type of object but got integer")
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
				)),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("number"))),
			)),
			test: func(t *testing.T, result interface{}, err error) {
				require.EqualError(t, err, "generate random data for schema failed: schema type=object properties=[a]: can not fill array with unique items: schema type=array unique-items items=schema type=integer minimum=0 maximum=3 minItems=5")
				require.Nil(t, result)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)

			o, err := schema.CreateValue(&schema.Ref{Value: tc.schema})

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

			o, err := schema.CreateValue(&schema.Ref{Value: tc.schema})

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
				s := schematest.New("object")
				props := &schema.Schemas{}
				props.Set("foo", &schema.Ref{Value: s})
				s.Properties = props

				result, err := schema.CreateValue(&schema.Ref{Value: s})
				require.NoError(t, err)

				b, err := json.Marshal(result)
				require.NoError(t, err)
				require.Equal(t, `{"foo":{"foo":null}}`, string(b))
			},
		},
		{
			"recursion across two objects depth 1",
			func(t *testing.T) {
				child := schematest.New("object")
				s := schematest.New("object", schematest.WithProperty("bar", child))
				props := &schema.Schemas{}
				props.Set("foo", &schema.Ref{Value: s})
				child.Properties = props

				result, err := schema.CreateValue(&schema.Ref{Value: s})
				require.NoError(t, err)
				require.NotNil(t, result)

				b, err := json.Marshal(result)
				require.NoError(t, err)
				require.Equal(t, `{"bar":{"foo":{"bar":null}}}`, string(b))
			},
		},
		{
			"array",
			func(t *testing.T) {
				obj := schematest.New("object")
				props := &schema.Schemas{}
				props.Set("foo", &schema.Ref{Value: obj})
				obj.Properties = props
				array := schematest.New("array")
				array.Items = &schema.Ref{Value: obj}
				minItems := 2
				array.MinItems = &minItems

				o, err := schema.CreateValue(&schema.Ref{Value: array})
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
			seed:   -77,
			test: func(t *testing.T, exp interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, exp)
			},
		},
		{
			name:   "nullable int",
			schema: schematest.New("integer", schematest.IsNullable(true)),
			seed:   -77,
			test: func(t *testing.T, exp interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, exp)
			},
		},
		{
			name:   "nullable number",
			schema: schematest.New("number", schematest.IsNullable(true)),
			seed:   -77,
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
			seed: -77,
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, result)
			},
		},
		{
			name: "nullable property",
			schema: schematest.New("object",
				schematest.WithProperty("foo", schematest.New("string", schematest.IsNullable(true)))),
			seed: -77,
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
			seed: -77,
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, result)
			},
		},
		{
			name: "nullable array item",
			schema: schematest.New("array",
				schematest.WithItems("string", schematest.IsNullable(true))),
			seed: 52,
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				arr := result.([]interface{})
				require.Nil(t, arr[1])
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(tc.seed)

			o, err := schema.CreateValue(&schema.Ref{Value: tc.schema})
			tc.test(t, o, err)
		})
	}
}

func _TestFindSeed(t *testing.T) {
	i := int64(0)
	for {
		gofakeit.Seed(i)

		o, _ := schema.CreateValue(&schema.Ref{Value: schematest.New("array",
			schematest.WithItems("string", schematest.IsNullable(true)))})

		for _, v := range o.([]interface{}) {
			if v == nil {
				require.NotNil(t, v, "seed %v", i)
				return
			}
		}

		i++
	}
}

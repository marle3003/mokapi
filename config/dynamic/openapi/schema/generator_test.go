package schema_test

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/sortedmap"
	"testing"
)

func toFloatP(f float64) *float64 { return &f }
func toIntP(i int) *int           { return &i }
func toBoolP(b bool) *bool        { return &b }

func TestGenerator(t *testing.T) {
	testdata := []struct {
		name   string
		exp    interface{}
		schema *schema.Schema
	}{
		{
			"no schema",
			nil,
			nil,
		},
		{
			"empty schema",
			nil,
			&schema.Schema{},
		},
		{
			"invalid type",
			nil,
			&schema.Schema{Type: "foobar"},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.Seed(11)

			g := schema.NewGenerator()
			o, err := g.New(&schema.Ref{Value: data.schema})
			require.NoError(t, err)
			require.Equal(t, data.exp, o)
		})
	}
}

func TestGeneratorString(t *testing.T) {
	testdata := []struct {
		name   string
		exp    interface{}
		schema *schema.Schema
	}{
		{
			"nil",
			nil,
			schematest.New(""),
		},
		{
			"string",
			"gbRMaRxHkiJBPta",
			&schema.Schema{Type: "string"},
		},
		{
			"by pattern",
			"013-64-5994",
			&schema.Schema{Type: "string", Pattern: "^\\d{3}-\\d{2}-\\d{4}$"},
		},
		{
			"date",
			"1953-01-24",
			&schema.Schema{Type: "string", Format: "date"},
		},
		{
			"date-time",
			"1953-01-24T13:00:35Z",
			&schema.Schema{Type: "string", Format: "date-time"},
		},
		{
			"password",
			"H|$9lb{J<+S;",
			&schema.Schema{Type: "string", Format: "password"},
		},
		{
			"email",
			"markusmoen@pagac.net",
			&schema.Schema{Type: "string", Format: "email"},
		},
		{
			"uuid",
			"98173564-6619-4557-888e-65b16bb5def5",
			&schema.Schema{Type: "string", Format: "uuid"},
		},
		{
			"url",
			"https://www.dynamiciterate.name/target/seamless",
			&schema.Schema{Type: "string", Format: "{url}"},
		},
		{
			"hostname",
			"centraltarget.biz",
			&schema.Schema{Type: "string", Format: "hostname"},
		},
		{
			"ipv4",
			"152.23.53.100",
			&schema.Schema{Type: "string", Format: "ipv4"},
		},
		{
			"ipv6",
			"8898:ee17:bc35:9064:5866:d019:3b95:7857",
			&schema.Schema{Type: "string", Format: "ipv6"},
		},
		{
			"beername",
			"Duvel",
			&schema.Schema{Type: "string", Format: "{beername}"},
		},
		{
			"address",
			"13645 Houston",
			&schema.Schema{Type: "string", Format: "{zip} {city}"},
		},
		{
			"uri",
			"https://www.dynamiciterate.name/target/seamless",
			&schema.Schema{Type: "string", Format: "uri"},
		},
	}

	for _, data := range testdata {
		d := data
		t.Run(d.name, func(t *testing.T) {
			gofakeit.Seed(11)

			g := schema.NewGenerator()
			o, err := g.New(&schema.Ref{Value: d.schema})
			require.NoError(t, err)
			require.Equal(t, d.exp, o)
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

			g := schema.NewGenerator()
			o, err := g.New(&schema.Ref{Value: data.schema})
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
				require.EqualError(t, err, "generating data failed: invalid minimum '5' and maximum '4' in schema type=integer format=int64 minimum=5 maximum=4")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)
			g := schema.NewGenerator()
			i, err := g.New(&schema.Ref{Value: tc.schema})
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
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.Seed(11)

			g := schema.NewGenerator()
			o, err := g.New(&schema.Ref{Value: data.schema})
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
				require.EqualError(t, err, "generating data failed: can not fill array with unique items for schema type=array minItems=5 maxItems=10 unique-items items=schema type=integer format=int32 minimum=0 maximum=3")
			},
		},
		{
			name: "unique items with enum",
			schema: &schema.Schema{Type: "array", MinItems: toIntP(5), MaxItems: toIntP(10), UniqueItems: true,
				Items: &schema.Ref{
					Value: &schema.Schema{
						Type:   "integer",
						Format: "int32",
						Enum:   []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9},
					},
				},
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{7, 3, 8, 6, 5}, i)
			},
		},
		{
			name: "unique items with enum and shuffle",
			schema: &schema.Schema{Type: "array", MinItems: toIntP(5), MaxItems: toIntP(10), UniqueItems: true,
				Items: &schema.Ref{
					Value: &schema.Schema{
						Type:   "integer",
						Format: "int32",
						Enum:   []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9},
					},
				},
				ShuffleItems: true,
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{5, 3, 8, 6, 7}, i)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.SetGlobalFaker(gofakeit.New(11))

			g := schema.NewGeneratorWithSeed(11)
			o, err := g.New(&schema.Ref{Value: tc.schema})
			tc.test(t, o, err)
		})
	}
}

func TestGeneratorObject(t *testing.T) {
	testdata := []struct {
		name   string
		exp    interface{}
		schema *schema.Schema
	}{
		{
			name: "simple",
			exp: map[string]interface{}{
				"id": int64(-8379641344161477543),
			},
			schema: schematest.New("object", schematest.WithProperty("id", &schema.Schema{Type: "integer", Format: "int64"})),
		},
		{
			name: "more fields",
			exp:  map[string]interface{}{"id": int32(-1072427943), "date": "1992-12-28"},
			schema: schematest.New("object",
				schematest.WithProperty("id", schematest.New("integer", schematest.WithFormat("int32"))),
				schematest.WithProperty("date", schematest.New("string", schematest.WithFormat("date")))),
		},
		{
			name: "nested",
			exp: map[string]interface{}{
				"nested": map[string]interface{}{
					"id":   int32(-1072427943),
					"date": "1992-12-28",
				},
			},
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
			exp: map[string]interface{}{
				"his":          "aYkWwfoRLOPxLIo",
				"most":         "VSeDjRRGUnsAxdB",
				"ocean":        "bRMaRxHkiJBPtap",
				"our":          "JdnSMKgtlxwnqhq",
				"straightaway": "aHGyyvqqdHueUxc",
				"theirs":       "qanPAKaXSMQFpZy",
			},
			schema: schematest.New("object",
				schematest.WithAdditionalProperties(schematest.New("string"))),
		},
		{
			"no fields defined",
			map[string]interface{}{},
			&schema.Schema{Type: "object"},
		},
		//{
		//	"example",
		//	struct {
		//		foo string
		//	}{foo: "bar"},
		//	&openapi.Schema{Type: "object", Example: map[string]interface{}{"foo": "bar"}},
		//},
	}

	var toMap func(m *sortedmap.LinkedHashMap[string, interface{}]) map[string]interface{}
	toMap = func(m *sortedmap.LinkedHashMap[string, interface{}]) map[string]interface{} {
		r := make(map[string]interface{})
		for it := m.Iter(); it.Next(); {
			v := it.Value()
			if vm, ok := v.(*sortedmap.LinkedHashMap[string, interface{}]); ok {
				r[it.Key()] = toMap(vm)
			} else {
				r[it.Key()] = it.Value()
			}
		}
		return r
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.Seed(11)

			g := schema.NewGenerator()
			o, err := g.New(&schema.Ref{Value: data.schema})
			require.NoError(t, err)
			require.Equal(t, data.exp, toMap(o.(*sortedmap.LinkedHashMap[string, interface{}])))
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
					schematest.WithItems(schematest.New("",
						schematest.Any(
							schematest.New("object",
								schematest.WithProperty("foo", schematest.New("string"))),
							schematest.New("object",
								schematest.WithProperty("bar",
									schematest.New("integer",
										schematest.WithMinimum(0),
										schematest.WithMaximum(5)))),
						),
					)),
				)
				g := schema.NewGenerator()
				o, err := g.New(&schema.Ref{Value: s})
				require.NoError(t, err)
				a, ok := o.([]interface{})
				require.True(t, ok, "should be an array")
				require.Len(t, a, 1)
				m := a[0].(*sortedmap.LinkedHashMap[string, interface{}])
				require.Equal(t, "RMaRxHkiJBPtapW", m.Get("foo"))
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
				m, ok := result.(*sortedmap.LinkedHashMap[string, interface{}])
				require.True(t, ok, "should be a sorted map")
				require.Equal(t, 2, m.Len())
				require.Equal(t, "gbRMaRxHkiJBPta", m.Get("foo"))
				require.Equal(t, 2.2451747541855905e+307, m.Get("bar"))
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
				m, ok := result.(*sortedmap.LinkedHashMap[string, interface{}])
				require.True(t, ok, "should be a sorted map")
				require.Equal(t, 1, m.Len())
				require.Equal(t, 1.644484108270445e+307, m.Get("bar"))
			},
		},
		{
			name: "reference value is null",
			schema: schematest.NewAllOfRefs(
				&schema.Ref{},
				&schema.Ref{
					Value: schematest.New("object", schematest.WithProperty("bar", schematest.New("number"))),
				},
			),
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				m, ok := result.(*sortedmap.LinkedHashMap[string, interface{}])
				require.True(t, ok, "should be a sorted map")
				require.Equal(t, 1, m.Len())
				require.Equal(t, 1.644484108270445e+307, m.Get("bar"))
			},
		},
		{
			name: "with integer type",
			schema: schematest.New("", schematest.AllOf(
				schematest.New("integer"),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("number"))),
			)),
			test: func(t *testing.T, result interface{}, err error) {
				require.EqualError(t, err, "generating data failed: allOf expects type of object but got integer")
				require.Nil(t, result)
			},
		},
		{
			name: "one is not object",
			schema: schematest.New("", schematest.AllOf(
				schematest.New("number"),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("number"))),
			)),
			test: func(t *testing.T, result interface{}, err error) {
				require.EqualError(t, err, "generating data failed: allOf expects type of object but got number")
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
							schematest.New("integer",
								schematest.WithMinItems(5),
								schematest.WithMinimum(0),
								schematest.WithMaximum(3),
							))),
				)),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("number"))),
			)),
			test: func(t *testing.T, result interface{}, err error) {
				require.EqualError(t, err, "generating data failed: allOf expects to be valid against all of subschemas: can not fill array with unique items for schema type=array unique-items items=schema type=integer minimum=0 maximum=3 minItems=5")
				require.Nil(t, result)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)

			g := schema.NewGenerator()
			o, err := g.New(&schema.Ref{Value: tc.schema})

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
				g := schema.NewGenerator()
				o, err := g.New(&schema.Ref{Value: s})
				require.NoError(t, err)
				require.NotNil(t, o)
				m := o.(*sortedmap.LinkedHashMap[string, interface{}])
				foo := m.Get("foo").(*sortedmap.LinkedHashMap[string, interface{}])
				require.Nil(t, foo.Get("foo"))
			},
		},
		{
			"recursion across to objects depth 1",
			func(t *testing.T) {
				child := schematest.New("object")
				s := schematest.New("object", schematest.WithProperty("bar", child))
				props := &schema.Schemas{}
				props.Set("foo", &schema.Ref{Value: s})
				child.Properties = props
				g := schema.NewGenerator()
				o, err := g.New(&schema.Ref{Value: s})
				require.NoError(t, err)
				require.NotNil(t, o)
				m := o.(*sortedmap.LinkedHashMap[string, interface{}])
				bar := m.Get("bar").(*sortedmap.LinkedHashMap[string, interface{}])
				foo := bar.Get("foo").(*sortedmap.LinkedHashMap[string, interface{}])
				require.Nil(t, foo.Get("foo"))
			},
		},
		{
			"array",
			func(t *testing.T) {
				obj := schematest.New("object")
				props := &schema.Schemas{}
				props.Set("foo", &schema.Ref{Value: obj})
				obj.Properties = props
				array := schematest.New("array", schematest.WithItems(obj))
				minItems := 2
				array.MinItems = &minItems
				g := schema.NewGenerator()
				o, err := g.New(&schema.Ref{Value: array})
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

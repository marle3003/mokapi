package openapi_test

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/test"
	"testing"
)

func toFloatP(f float64) *float64 { return &f }
func toIntP(i int) *int           { return &i }

func TestGenerator(t *testing.T) {
	testdata := []struct {
		name   string
		exp    interface{}
		schema *openapi.Schema
	}{
		{
			"no schema",
			nil,
			nil,
		},
		{
			"empty schema",
			nil,
			&openapi.Schema{},
		},
		{
			"invalid type",
			nil,
			&openapi.Schema{Type: "foobar"},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.Seed(11)

			g := openapi.NewGenerator()
			o := g.New(&openapi.SchemaRef{Value: data.schema})
			test.Equals(t, data.exp, o)
		})
	}
}

func TestGeneratorString(t *testing.T) {
	testdata := []struct {
		name   string
		exp    interface{}
		schema *openapi.Schema
	}{
		{
			"nil",
			nil,
			openapitest.NewSchema(""),
		},
		{
			"string",
			"gbRMaRxHkiJBPta",
			&openapi.Schema{Type: "string"},
		},
		{
			"by pattern",
			"013-64-5994",
			&openapi.Schema{Type: "string", Pattern: "^\\d{3}-\\d{2}-\\d{4}$"},
		},
		{
			"date",
			"2008-12-07",
			&openapi.Schema{Type: "string", Format: "date"},
		},
		{
			"date-time",
			"2008-12-07T04:14:25Z",
			&openapi.Schema{Type: "string", Format: "date-time"},
		},
		{
			"password",
			"H|$9lb{J<+S;",
			&openapi.Schema{Type: "string", Format: "password"},
		},
		{
			"email",
			"markusmoen@pagac.net",
			&openapi.Schema{Type: "string", Format: "email"},
		},
		{
			"uuid",
			"590c1440-9888-45b0-bd51-a817ee07c3f2",
			&openapi.Schema{Type: "string", Format: "uuid"},
		},
		{
			"url",
			"http://www.principalproductize.biz/target",
			&openapi.Schema{Type: "string", Format: "{url}"},
		},
		{
			"hostname",
			"centraltarget.biz",
			&openapi.Schema{Type: "string", Format: "hostname"},
		},
		{
			"ipv4",
			"152.23.53.100",
			&openapi.Schema{Type: "string", Format: "ipv4"},
		},
		{
			"ipv6",
			"8898:ee17:bc35:9064:5866:d019:3b95:7857",
			&openapi.Schema{Type: "string", Format: "ipv6"},
		},
		{
			"beername",
			"Duvel",
			&openapi.Schema{Type: "string", Format: "{beername}"},
		},
		{
			"address",
			"13645 Mrazhaven",
			&openapi.Schema{Type: "string", Format: "{zip} {city}"},
		},
	}

	for _, data := range testdata {
		d := data
		t.Run(d.name, func(t *testing.T) {
			gofakeit.Seed(11)

			g := openapi.NewGenerator()
			o := g.New(&openapi.SchemaRef{Value: d.schema})
			test.Equals(t, d.exp, o)
		})
	}
}

func TestGeneratorBool(t *testing.T) {
	testdata := []struct {
		name   string
		exp    interface{}
		schema *openapi.Schema
	}{
		{
			"false",
			false,
			&openapi.Schema{Type: "boolean"},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.Seed(11)

			g := openapi.NewGenerator()
			o := g.New(&openapi.SchemaRef{Value: data.schema})
			test.Equals(t, data.exp, o)
		})
	}
}

func TestGeneratorInt(t *testing.T) {
	testdata := []struct {
		name   string
		exp    interface{}
		schema *openapi.Schema
	}{
		{
			"int32",
			int32(-1072427943),
			&openapi.Schema{Type: "integer", Format: "int32"},
		},
		{
			"int32 min",
			int32(196446384),
			&openapi.Schema{Type: "integer", Format: "int32", Minimum: toFloatP(10)},
		},
		{
			"int32 max",
			int32(-1951037312),
			&openapi.Schema{Type: "integer", Format: "int32", Maximum: toFloatP(0)},
		},
		{
			"int32 min max",
			int32(-4),
			&openapi.Schema{Type: "integer", Format: "int32", Minimum: toFloatP(-5), Maximum: toFloatP(5)},
		},
		{
			"int64",
			int64(-8379641344161477543),
			&openapi.Schema{Type: "integer", Format: "int64"},
		},
		{
			"int64 min",
			int64(843730692693298304),
			&openapi.Schema{Type: "integer", Format: "int64", Minimum: toFloatP(10)},
		},
		{
			"int64 max",
			int64(-8379641344161477632),
			&openapi.Schema{Type: "integer", Format: "int64", Maximum: toFloatP(0)},
		},
		{
			"int64 min max",
			int64(-4),
			&openapi.Schema{Type: "integer", Format: "int64", Minimum: toFloatP(-5), Maximum: toFloatP(5)},
		},
		{
			"int64 min max positive",
			int64(5),
			&openapi.Schema{Type: "integer", Format: "int64", Minimum: toFloatP(4), Maximum: toFloatP(10)},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.Seed(11)

			g := openapi.NewGenerator()
			o := g.New(&openapi.SchemaRef{Value: data.schema})
			test.Equals(t, data.exp, o)
		})
	}
}

func TestGeneratorFloat(t *testing.T) {
	testdata := []struct {
		name   string
		exp    interface{}
		schema *openapi.Schema
	}{
		{
			"float",
			float32(3.1128167e+37),
			&openapi.Schema{Type: "number", Format: "float"},
		},
		{
			"float min",
			float32(3.1128167e+37),
			&openapi.Schema{Type: "number", Format: "float", Minimum: toFloatP(10)},
		},
		{
			"float max",
			float32(-3.0915418e+38),
			&openapi.Schema{Type: "number", Format: "float", Maximum: toFloatP(0)},
		},
		{
			"float min max",
			float32(-4.085225),
			&openapi.Schema{Type: "number", Format: "float", Minimum: toFloatP(-5), Maximum: toFloatP(5)},
		},
		{
			"double",
			1.644484108270445e+307,
			&openapi.Schema{Type: "number", Format: "double"},
		},
		{
			"double min",
			1.644484108270445e+307,
			&openapi.Schema{Type: "number", Format: "double", Minimum: toFloatP(10)},
		},
		{
			"double max",
			-1.6332447240352712e+308,
			&openapi.Schema{Type: "number", Format: "double", Maximum: toFloatP(0)},
		},
		{
			"double min max",
			-4.085225349989226,
			&openapi.Schema{Type: "number", Format: "double", Minimum: toFloatP(-5), Maximum: toFloatP(5)},
		},
		{
			"example",
			1,
			&openapi.Schema{Type: "number", Format: "double", Example: 1},
		},
		{
			"enum",
			1,
			&openapi.Schema{Type: "number", Format: "double", Enum: []interface{}{1, 2, 3, 4}},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.Seed(11)

			g := openapi.NewGenerator()
			o := g.New(&openapi.SchemaRef{Value: data.schema})
			test.Equals(t, data.exp, o)
		})
	}
}

func TestGeneratorArray(t *testing.T) {
	testdata := []struct {
		name   string
		exp    interface{}
		schema *openapi.Schema
	}{
		{
			"simple",
			[]interface{}{},
			&openapi.Schema{Type: "array", Items: &openapi.SchemaRef{
				Value: &openapi.Schema{
					Type: "integer", Format: "int32", Minimum: toFloatP(0), Maximum: toFloatP(10)}}},
		},
		{
			"min items",
			[]interface{}{int32(1), int32(8), int32(8), int32(6), int32(7)},
			&openapi.Schema{Type: "array", MinItems: toIntP(5), Items: &openapi.SchemaRef{
				Value: &openapi.Schema{
					Type: "integer", Format: "int32", Minimum: toFloatP(0), Maximum: toFloatP(10)}}},
		},
		{
			"min items",
			[]interface{}{int32(8), int32(8), int32(6), int32(7), int32(1), int32(8), int32(9), int32(5), int32(3), int32(1)},
			&openapi.Schema{Type: "array", MinItems: toIntP(5), MaxItems: toIntP(10), Items: &openapi.SchemaRef{
				Value: &openapi.Schema{
					Type: "integer", Format: "int32", Minimum: toFloatP(0), Maximum: toFloatP(10)}}},
		},
		{
			"unique items",
			[]interface{}{int32(8), int32(6), int32(7), int32(1), int32(9), int32(5), int32(3), int32(2), int32(4), int32(10)},
			&openapi.Schema{Type: "array", MinItems: toIntP(5), MaxItems: toIntP(10), UniqueItems: true,
				Items: &openapi.SchemaRef{
					Value: &openapi.Schema{
						Type: "integer", Format: "int32", Minimum: toFloatP(0), Maximum: toFloatP(10)}}},
		},
		{
			"enum ignores items config",
			[]interface{}{1, 2, 3},
			&openapi.Schema{Type: "array", MinItems: toIntP(5), MaxItems: toIntP(10), UniqueItems: true,
				Enum: []interface{}{
					[]interface{}{1, 2, 3},
					[]interface{}{3, 2, 1},
				},
				Items: &openapi.SchemaRef{
					Value: &openapi.Schema{
						Type: "integer", Format: "int32", Minimum: toFloatP(0), Maximum: toFloatP(3)}}},
		},
		{
			"example",
			[]interface{}{1, 2, 3},
			&openapi.Schema{Type: "array", MinItems: toIntP(5), MaxItems: toIntP(10), UniqueItems: true,
				Example: []interface{}{1, 2, 3},
				Items: &openapi.SchemaRef{
					Value: &openapi.Schema{
						Type: "integer", Format: "int32", Minimum: toFloatP(0), Maximum: toFloatP(3)}}},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.SetGlobalFaker(gofakeit.New(11))

			g := openapi.NewGenerator()
			o := g.New(&openapi.SchemaRef{Value: data.schema})
			test.Equals(t, data.exp, o)
		})
	}

	t.Run("unique items panic", func(t *testing.T) {
		defer func() {
			r := recover()
			test.Equals(t, "can not fill array with unique items", r)
		}()

		gofakeit.Seed(11)
		g := openapi.NewGenerator()
		g.New(&openapi.SchemaRef{Value: &openapi.Schema{Type: "array", MinItems: toIntP(5), MaxItems: toIntP(10), UniqueItems: true,
			Items: &openapi.SchemaRef{
				Value: &openapi.Schema{
					Type: "integer", Format: "int32", Minimum: toFloatP(0), Maximum: toFloatP(3)}}},
		})
	})
}

func TestGeneratorObject(t *testing.T) {
	testdata := []struct {
		name   string
		exp    interface{}
		schema *openapi.Schema
	}{
		{
			"simple",
			&struct {
				Id int64
			}{Id: int64(-8379641344161477543)},
			openapitest.NewSchema("object", openapitest.WithProperty("id", &openapi.Schema{Type: "integer", Format: "int64"})),
		},
		// unable to test because of unordered map which generates random values despite fixed seed
		//{
		//	"more fields",
		//	map[string]interface{}{"id": 5624956352167149568, "date": "1943-5-23"},
		//	&openapi.Schema{Type: "object", Properties: &openapi.Schemas{
		//		Value: map[string]*openapi.SchemaRef{
		//			"id":   {Value: &openapi.Schema{Type: "integer", Format: "int32"}},
		//			"date": {Value: &openapi.Schema{Type: "string", Format: "date"}},
		//		},
		//	}},
		//},
		//{
		//	"nested",
		//	map[string]interface{}{
		//		"nested": map[string]interface{}{
		//			"id":   3225897846662234112,
		//			"date": "1943-5-23",
		//		},
		//	},
		//	&openapi.Schema{Type: "object", Properties: &openapi.Schemas{
		//		Value: map[string]*openapi.SchemaRef{
		//			"nested": {Value: &openapi.Schema{Type: "object", Properties: &openapi.Schemas{
		//				Value: map[string]*openapi.SchemaRef{
		//					"id":   {Value: &openapi.Schema{Type: "integer", Format: "int32"}},
		//					"date": {Value: &openapi.Schema{Type: "string", Format: "date"}},
		//				},
		//			}}},
		//		},
		//	}},
		//},
		{
			"no fields defined",
			&struct{}{},
			&openapi.Schema{Type: "object"},
		},
		//{
		//	"example",
		//	struct {
		//		foo string
		//	}{foo: "bar"},
		//	&openapi.Schema{Type: "object", Example: map[string]interface{}{"foo": "bar"}},
		//},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			gofakeit.Seed(11)

			g := openapi.NewGenerator()
			o := g.New(&openapi.SchemaRef{Value: data.schema})
			test.Equals(t, data.exp, o)
		})
	}
}

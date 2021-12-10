package openapi_test

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/test"
	"testing"
)

func TestGenerator(t *testing.T) {
	testdata := []struct {
		name   string
		exp    interface{}
		schema *openapi.Schema
	}{
		{
			"nil",
			nil,
			openapitest.NewSchema(),
		},
		{
			"string",
			"hrUKPttUEzpTnEU",
			&openapi.Schema{Type: "string"},
		},
		{
			"float",
			float32(1.2693497e+38),
			&openapi.Schema{Type: "number", Format: "float"},
		},
		//{
		//	"float min",
		//	float32(2.4240552e+38),
		//	&openapi.Schema{Type: "number", Format: "float", Minimum: new(float64(1.0))},
		//},
		//{
		//	"float max",
		//	float32(2.4240552e+38),
		//	&openapi.Schema{Type: "number", Format: "float"},
		//},
	}

	gofakeit.SetGlobalFaker(gofakeit.New(42))

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			g := openapi.NewGenerator()
			o := g.New(&openapi.SchemaRef{Value: data.schema})
			test.Equals(t, data.exp, o)
		})
	}
}

package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/json/schematest"
	"testing"
)

func toFloat64P(f float64) *float64 { return &f }
func toIntP(i int) *int             { return &i }

func TestNumber(t *testing.T) {
	testcases := []struct {
		name    string
		request *Request
		test    func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "id",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "id",
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 37727, v)
			},
		},
		{
			name: "id with max",
			request: &Request{
				Path: Path{
					&PathElement{Name: "id", Schema: schematest.NewRef("integer", schematest.WithMaximum(10000))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 7727, v)
			},
		},
		{
			name: "id with min & max",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "id",
						Schema: schematest.NewRef("integer",
							schematest.WithMinimum(10),
							schematest.WithMaximum(20),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 18, v)
			},
		},
		{
			name: "ids",
			request: &Request{
				Path: Path{
					&PathElement{
						Name:   "ids",
						Schema: schematest.NewRef("array"),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{83580, 80588}, v)
			},
		},
		{
			name: "year no schema",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "year",
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 1926, v)
			},
		},
		{
			name: "year",
			request: &Request{
				Path: Path{
					&PathElement{
						Name:   "year",
						Schema: schematest.NewRef("integer"),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 1926, v)
			},
		},
		{
			name: "year min",
			request: &Request{
				Path: Path{
					&PathElement{
						Name:   "year",
						Schema: schematest.NewRef("integer", schematest.WithMinimum(1990)),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 2196, v)
			},
		},
		{
			name: "year min max",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "year",
						Schema: schematest.NewRef("integer",
							schematest.WithMinimum(1990),
							schematest.WithMaximum(2049),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 2016, v)
			},
		},
		{
			name: "quantity",
			request: &Request{
				Path: Path{
					&PathElement{
						Name:   "quantity",
						Schema: schematest.NewRef("integer"),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 79, v)
			},
		},
		{
			name: "quantity min max",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "year",
						Schema: schematest.NewRef("integer",
							schematest.WithMinimum(0),
							schematest.WithMaximum(50),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 23, v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.request)
			tc.test(t, v, err)
		})
	}
}

func TestInt32(t *testing.T) {
	testcases := []struct {
		name    string
		request *Request
		test    func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "int32",
			request: &Request{
				Path: Path{
					&PathElement{
						Name:   "size",
						Schema: schematest.NewRef("integer", schematest.WithFormat("int32")),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(271629950), v)
			},
		},
		{
			name: "int32 with max=15",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "size",
						Schema: schematest.NewRef("integer",
							schematest.WithFormat("int32"),
							schematest.WithMaximum(15),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(-838395520), v)
			},
		},
		{
			name: "int32 with min=0, max=15",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "size",
						Schema: schematest.NewRef("integer",
							schematest.WithFormat("int32"),
							schematest.WithMinimum(0),
							schematest.WithMaximum(15),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(9), v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.request)
			tc.test(t, v, err)
		})
	}
}

func TestInteger(t *testing.T) {
	testcases := []struct {
		name    string
		request *Request
		test    func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "integer",
			request: &Request{
				Path: Path{
					&PathElement{
						Name:   "size",
						Schema: schematest.NewRef("integer"),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-3600881594791838082), v)
			},
		},
		{
			name: "integer with max=15",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "size",
						Schema: schematest.NewRef("integer",
							schematest.WithMaximum(15),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-3600881594791837696), v)
			},
		},
		{
			name: "integer with min=0, max=15",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "size",
						Schema: schematest.NewRef("integer",
							schematest.WithMinimum(0),
							schematest.WithMaximum(15),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(9), v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.request)
			tc.test(t, v, err)
		})
	}
}

func TestFloat32(t *testing.T) {
	testcases := []struct {
		name    string
		request *Request
		test    func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "float",
			request: &Request{
				Path: Path{
					&PathElement{
						Name:   "size",
						Schema: schematest.NewRef("number", schematest.WithFormat("float")),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, float32(2.0743327e+38), v)
			},
		},
		{
			name: "float with max=15",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "size",
						Schema: schematest.NewRef("number",
							schematest.WithFormat("float"),
							schematest.WithMaximum(15),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, float32(-1.3284907e+38), v)
			},
		},
		{
			name: "float with min=0, max=15",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "size",
						Schema: schematest.NewRef("number",
							schematest.WithFormat("float"),
							schematest.WithMinimum(0),
							schematest.WithMaximum(15),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, float32(9.143875), v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.request)
			tc.test(t, v, err)
		})
	}
}

func TestFloat(t *testing.T) {
	testcases := []struct {
		name    string
		request *Request
		test    func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "number",
			request: &Request{
				Path: Path{
					&PathElement{
						Name:   "size",
						Schema: schematest.NewRef("number"),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 1.0958586976799703e+308, v)
			},
		},
		{
			name: "number with max=15",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "size",
						Schema: schematest.NewRef("number",
							schematest.WithMaximum(15),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, -7.018344371823454e+307, v)
			},
		},
		{
			name: "number with min=0, max=15",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "size",
						Schema: schematest.NewRef("number",
							schematest.WithMinimum(0),
							schematest.WithMaximum(15),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 9.143874528095433, v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.request)
			tc.test(t, v, err)
		})
	}
}

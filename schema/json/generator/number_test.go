package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestNumber(t *testing.T) {
	testcases := []struct {
		name    string
		request *Request
		test    func(t *testing.T, v interface{}, err error)
	}{}

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
						Schema: schematest.New("integer", schematest.WithFormat("int32")),
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
						Schema: schematest.New("integer",
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
						Schema: schematest.New("integer",
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
		{
			name: "int32 exclusive",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "size",
						Schema: schematest.New("integer",
							schematest.WithFormat("int32"),
							schematest.WithExclusiveMinimum(0),
							schematest.WithExclusiveMaximum(15),
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
						Schema: schematest.New("integer"),
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
						Schema: schematest.New("integer",
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
						Schema: schematest.New("integer",
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
						Schema: schematest.New("number", schematest.WithFormat("float")),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, float32(2.0743327e+38), v)
			},
		},
		// random number is different between win/linux to macos
		//{
		//	name: "float with max=15",
		//	request: &Request{
		//		Path: Path{
		//			&PathElement{
		//				Name: "size",
		//				Schema: schematest.NewRef("number",
		//					schematest.WithFormat("float"),
		//					schematest.WithMaximum(15),
		//				),
		//			},
		//		},
		//	},
		//	test: func(t *testing.T, v interface{}, err error) {
		//		require.NoError(t, err)
		//		require.Equal(t, float32(-1.3284907e+38), v)
		//	},
		//},
		{
			name: "float with min=0, max=15",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "size",
						Schema: schematest.New("number",
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
						Schema: schematest.New("number"),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 1.0958586976799703e+308, v)
			},
		},
		// random number is different between win/linux to macos
		//{
		//	name: "number with max=15",
		//	request: &Request{
		//		Path: Path{
		//			&PathElement{
		//				Name: "size",
		//				Schema: schematest.NewRef("number",
		//					schematest.WithMaximum(15),
		//				),
		//			},
		//		},
		//	},
		//	test: func(t *testing.T, v interface{}, err error) {
		//		require.NoError(t, err)
		//		require.Equal(t, -7.018344371823454e+307, v)
		//	},
		//},
		{
			name: "number with min=0, max=15",
			request: &Request{
				Path: Path{
					&PathElement{
						Name: "size",
						Schema: schematest.New("number",
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

package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
	"testing"
)

func TestExamples(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "examples",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: &schema.Ref{
							Value: &schema.Schema{
								Examples: []interface{}{
									"Anything", 4035,
								},
							},
						},
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Anything", v)
			},
		},
		{
			name: "invalid example",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: &schema.Ref{
							Value: schematest.New("integer", schematest.WithExample("foo")),
						},
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-3600881594791838082), v)
			},
		},
		{
			name: "invalid examples",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: &schema.Ref{
							Value: schematest.New("integer", schematest.WithExamples("foo", "bar", "test")),
						},
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-3652171958352792229), v)
			},
		},
		{
			name: "invalid examples but last is valid",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: &schema.Ref{
							Value: schematest.New("integer", schematest.WithExamples("foo", "bar", "test", int64(1234))),
						},
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(1234), v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)
			Seed(1234567)

			v, err := New(tc.req)
			tc.test(t, v, err)
		})
	}
}

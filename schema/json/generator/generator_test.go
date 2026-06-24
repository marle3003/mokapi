package generator_test

import (
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestGenerator(t *testing.T) {
	testcases := []struct {
		name string
		s    *schema.Schema
		test func(t *testing.T, v any, err error)
	}{
		{
			name: "generator should use $id when path is empty",
			s:    schematest.New("object", schematest.WithId("/schemas/address")),
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{
					"address":   "2824 North Walkborough, Plano, California 22554",
					"city":      "Plano",
					"country":   "Brazil",
					"latitude":  17.057939,
					"longitude": 30.125984,
					"state":     "California",
					"street":    "2824 North Walkborough",
					"zip":       "22554",
				}, v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(1234567)

			v, err := generator.New(generator.NewRequest(nil, tc.s, nil))
			tc.test(t, v, err)
		})
	}
}

func TestFakeNotDefined(t *testing.T) {
	gofakeit.Seed(1234567)

	root := generator.FindByName(generator.RootName)
	require.NotNil(t, root)

	children := root.Children
	defer func() {
		root.Children = children
	}()

	root.Children = []*generator.Node{
		{
			Name: "test",
		},
	}

	v, err := generator.New(generator.NewRequest(nil, nil, nil))
	require.NoError(t, err)
	require.Equal(t, -170715.30581115812, v)
}

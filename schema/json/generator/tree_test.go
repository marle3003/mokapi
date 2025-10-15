package generator_test

import (
	"mokapi/schema/json/generator"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindByName(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "root",
			test: func(t *testing.T) {
				n := generator.FindByName("root")
				require.NotNil(t, n)
				require.Equal(t, "root", n.Name)
			},
		},
		{
			name: "street",
			test: func(t *testing.T) {
				n := generator.FindByName("street")
				require.NotNil(t, n)
				require.Equal(t, "street", n.Name)
				v, err := n.Fake(generator.NewRequest(nil, nil, nil))
				require.NoError(t, err)
				require.Equal(t, "75220 Forestborough", v)
			},
		},
		{
			name: "absolute /house/number",
			test: func(t *testing.T) {
				n := generator.FindByName("/house/number")
				require.NotNil(t, n)
				require.Equal(t, "number", n.Name)
				v, err := n.Fake(generator.NewRequest(nil, nil, nil))
				require.NoError(t, err)
				require.Equal(t, "375", v)
			},
		},
		{
			name: "absolute /house/number2",
			test: func(t *testing.T) {
				n := generator.FindByName("/house/number2")
				require.Nil(t, n)
			},
		},
		{
			name: "relative middle/name",
			test: func(t *testing.T) {
				n := generator.FindByName("middle/name")
				require.NotNil(t, n)
				require.Equal(t, "name", n.Name)
				v, err := n.Fake(generator.NewRequest(nil, nil, nil))
				require.NoError(t, err)
				require.Equal(t, "Noel", v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(1234578)
			tc.test(t)
		})
	}
}

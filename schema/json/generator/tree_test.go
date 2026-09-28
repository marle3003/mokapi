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
				require.Equal(t, "295 Estatestown", v)
			},
		},
		{
			name: "absolute /street",
			test: func(t *testing.T) {
				n := generator.FindByName("/street")
				require.NotNil(t, n)
				require.Equal(t, "street", n.Name)
				v, err := n.Fake(generator.NewRequest(nil, nil, nil))
				require.NoError(t, err)
				require.Equal(t, "295 Estatestown", v)
			},
		},
		{
			name: "absolute /address/line1",
			test: func(t *testing.T) {
				n := generator.FindByName("/address/line1")
				require.NotNil(t, n)
				require.Equal(t, "line1", n.Name)
				v, err := n.Fake(generator.NewRequest(nil, nil, nil))
				require.NoError(t, err)
				require.Equal(t, "Grayson Anderson", v)
			},
		},
		{
			name: "absolute /address/line10",
			test: func(t *testing.T) {
				n := generator.FindByName("/house/number2")
				require.Nil(t, n)
			},
		},
		{
			name: "relative contact/phone",
			test: func(t *testing.T) {
				n := generator.FindByName("contact/phone")
				require.NotNil(t, n)
				require.Equal(t, "phone", n.Name)
				v, err := n.Fake(generator.NewRequest(nil, nil, nil))
				require.NoError(t, err)
				require.Equal(t, "+1062951049628", v)
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

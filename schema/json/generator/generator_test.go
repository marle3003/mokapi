package generator_test

import (
	"mokapi/schema/json/generator"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
)

func TestFakeNotDefined(t *testing.T) {
	gofakeit.Seed(1234567)

	root := generator.FindByName(generator.RootName)
	root.Children = []*generator.Node{
		{
			Name: "test",
		},
	}

	v, err := generator.New(generator.NewRequest(nil, nil, nil))
	require.NoError(t, err)
	require.Equal(t, int64(337128), v)
}

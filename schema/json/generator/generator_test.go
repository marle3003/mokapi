package generator_test

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/generator"
	"testing"
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
	require.Equal(t, int64(-3652171958352792229), v)
}

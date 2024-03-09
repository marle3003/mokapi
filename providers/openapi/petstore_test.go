package openapi

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestPetStore(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, c *Config)
	}{
		{
			name: "title",
			test: func(t *testing.T, c *Config) {
				require.Equal(t, "Swagger Petstore - OpenAPI 3.0", c.Info.Name)
			},
		},
		{
			name: "schema Pet",
			test: func(t *testing.T, c *Config) {
				pet := c.Components.Schemas.Get("Pet")
				require.Equal(t, "object", pet.Value.Type)
				require.Equal(t, 6, pet.Value.Properties.Len())
				id := pet.Value.Properties.Get("id")
				require.Equal(t, "integer", id.Value.Type)
			},
		},
	}

	var c *Config
	b, err := os.ReadFile("../../examples/swagger/petstore-3.0.json")
	require.NoError(t, err)
	err = json.Unmarshal(b, &c)
	require.NoError(t, err)

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t, c)
		})
	}
}

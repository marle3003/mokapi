package openapi_test

import (
	"encoding/json"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// An operation with "security: []" explicitly has no security requirement, which is different
// from an operation without a security key (it inherits the global requirement). The model
// tells them apart by nil (absent) and empty (explicitly none).
func TestOperation_EmptySecurityIsKept(t *testing.T) {
	t.Run("yaml empty list", func(t *testing.T) {
		o := openapi.Operation{}
		require.NoError(t, yaml.Unmarshal([]byte(`security: []`), &o))
		require.NotNil(t, o.Security)
		require.Empty(t, o.Security)
	})
	t.Run("yaml without key", func(t *testing.T) {
		o := openapi.Operation{}
		require.NoError(t, yaml.Unmarshal([]byte(`summary: foo`), &o))
		require.Nil(t, o.Security)
	})
	t.Run("json empty list", func(t *testing.T) {
		o := openapi.Operation{}
		require.NoError(t, json.Unmarshal([]byte(`{ "security": [] }`), &o))
		require.NotNil(t, o.Security)
		require.Empty(t, o.Security)
	})
	t.Run("json without key", func(t *testing.T) {
		o := openapi.Operation{}
		require.NoError(t, json.Unmarshal([]byte(`{ "summary": "foo" }`), &o))
		require.Nil(t, o.Security)
	})
}

func TestConfig_Patch_OperationEmptySecurity(t *testing.T) {
	t.Run("new operation is a clone of the patch", func(t *testing.T) {
		result := openapitest.NewConfig("1.0", openapitest.WithPath("/foo"))
		result.Patch(openapitest.NewConfig("1.0", openapitest.WithPath("/foo",
			openapitest.WithOperation("get", openapitest.WithEmptySecurity()))))

		o := result.Paths.Lookup("/foo").Value.Get
		require.NotNil(t, o.Security)
		require.Empty(t, o.Security)
	})
	t.Run("existing operation without security", func(t *testing.T) {
		result := openapitest.NewConfig("1.0", openapitest.WithPath("/foo", openapitest.WithOperation("get")))
		result.Patch(openapitest.NewConfig("1.0", openapitest.WithPath("/foo",
			openapitest.WithOperation("get", openapitest.WithEmptySecurity()))))

		o := result.Paths.Lookup("/foo").Value.Get
		require.NotNil(t, o.Security, "patch with security: [] must keep 'no security'")
	})
}

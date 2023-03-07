package openapi

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi/parameter"
	"net/http"
	"testing"
)

func TestContent_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name string
		s    string
		fn   func(t *testing.T, c Content)
	}{
		{
			"empty",
			"{}",
			func(t *testing.T, c Content) {
				require.Len(t, c, 0)
			},
		},
		{
			"with ref",
			`{
"application/json": {
  "schema": {
    "$ref": "#/components/schemas/Foo"
  }
}
}`,
			func(t *testing.T, c Content) {
				require.Len(t, c, 1)
				require.Contains(t, c, "application/json")
				ct := c["application/json"].ContentType
				require.Equal(t, "application", ct.Type)
				require.Equal(t, "json", ct.Subtype)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := Content{}
			err := json.Unmarshal([]byte(tc.s), &c)
			require.NoError(t, err)
			tc.fn(t, c)
		})
	}
}

func TestConfig_UnmarshalJSON(t *testing.T) {
	s := `{
	"openapi": "3.0.0",
	"info": {
		"version": "1.0.0",
		"title": "Pet Store"
	},
	"servers": [
		{
			"url": "/foo"
		}
	],
	"paths": {
		"/pets": {
			"summary": "foo summary",
			"get": {
				"summary": "foo get",
				"operationId": "listPets",
				"tags": ["pets"],
				"parameters": [
					{
						"name": "limit",
						"in": "query",
						"description": "How many items to return",
						"required": false,
						"schema": {"type": "integer", "maximum": 100, "format": "int32" }
					}
				],
				"responses": {
					"200": {
						"description": "a paged array of pets",
						"headers": {
							"next": {
								"description": "A link to the next page of responses",
								"schema": {"type":"string"}
							}
						},
						"content": {
							"application/json": {"schema": {"$ref": "#/components/schemas/Pets"}}
						}
					}
				}
			}
		}
	}
}`
	c := &Config{}
	err := json.Unmarshal([]byte(s), &c)
	require.NoError(t, err)
	require.Equal(t, "3.0.0", c.OpenApi)
	require.Equal(t, "1.0.0", c.Info.Version)
	require.Equal(t, "Pet Store", c.Info.Name)
	require.Equal(t, "/foo", c.Servers[0].Url)
	require.Contains(t, c.Paths.Value, "/pets")
	pets := c.Paths.Value["/pets"]
	require.Equal(t, "foo summary", pets.Value.Summary)
	require.Equal(t, "foo get", pets.Value.Get.Summary)
	require.Equal(t, "listPets", pets.Value.Get.OperationId)
	require.Equal(t, "pets", pets.Value.Get.Tags[0])
	// parameter
	require.Equal(t, "limit", pets.Value.Get.Parameters[0].Value.Name)
	require.Equal(t, parameter.Query, pets.Value.Get.Parameters[0].Value.Type)
	require.Equal(t, "How many items to return", pets.Value.Get.Parameters[0].Value.Description)
	require.False(t, pets.Value.Get.Parameters[0].Value.Required)
	require.Equal(t, "integer", pets.Value.Get.Parameters[0].Value.Schema.Value.Type)
	require.Equal(t, float64(100), *pets.Value.Get.Parameters[0].Value.Schema.Value.Maximum)
	require.Equal(t, "int32", pets.Value.Get.Parameters[0].Value.Schema.Value.Format)
	// responses
	response := pets.Value.Get.Responses.GetResponse(http.StatusOK)
	require.Equal(t, "a paged array of pets", response.Description)
	require.Equal(t, "A link to the next page of responses", response.Headers["next"].Value.Description)
	require.Equal(t, "string", response.Headers["next"].Value.Schema.Value.Type)
	require.Equal(t, "#/components/schemas/Pets", response.Content["application/json"].Schema.Ref)
}

func TestResponses_UnmarshalJSON(t *testing.T) {
	s := `{"200": {
"description": "Success"
}}`
	r := &Responses{}
	err := json.Unmarshal([]byte(s), &r)
	require.NoError(t, err)
	require.Equal(t, "Success", r.GetResponse(200).Description)
}

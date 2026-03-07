package openapi_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/openapi"
	json "mokapi/schema/json/schema"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConfig_DynamicSchema(t *testing.T) {
	testdata := []struct {
		name string
		data string
		test func(t *testing.T, c *openapi.Config)
	}{
		{
			name: "dynamic reference using $id",
			data: `
openapi: "3.0.0"
info:
  title: "Dynamic Schema"
paths:
  /foo:
    get:
      responses:
        '200':
          content:
            'application/json':
              schema:
                $id: /foo
                $ref: '#/components/schemas/Response'
                $defs:
                  content:
                    $dynamicAnchor: T
                    type: object
  /bar:
    get:
      responses:
        '200':
          content:
            'application/json':
              schema:
                $id: /bar
                $ref: '#/components/schemas/Response'
                $defs:
                  content:
                    $dynamicAnchor: T
                    type: array
components:
  schemas:
    Response:
      $defs:
        content:
          $dynamicAnchor: T
          not: true
      type: object
      properties:
        content:
          $dynamicRef: '#T'
        error:
          type: object
`,
			test: func(t *testing.T, c *openapi.Config) {
				require.NotNil(t, c)
				foo := c.Paths["/foo"].Value.Get.Responses.GetResponse(http.StatusOK).Content["application/json"].Schema
				require.NotNil(t, foo)
				require.NotNil(t, foo.Properties)
				require.Equal(t, json.Types{"object"}, foo.Properties.Get("content").Type)

				bar := c.Paths["/bar"].Value.Get.Responses.GetResponse(http.StatusOK).Content["application/json"].Schema
				require.NotNil(t, bar)
				require.NotNil(t, bar.Properties)
				require.Equal(t, json.Types{"array"}, bar.Properties.Get("content").Type)

			},
		},
		{
			name: "dynamic reference without $id",
			data: `
openapi: "3.0.0"
info:
  title: "Dynamic Schema"
paths:
  /foo:
    get:
      responses:
        '200':
          content:
            'application/json':
              schema:
                $ref: '#/components/schemas/Response'
                $defs:
                  content:
                    $dynamicAnchor: T
                    type: object
  /bar:
    get:
      responses:
        '200':
          content:
            'application/json':
              schema:
                $id: /bar
                $ref: '#/components/schemas/Response'
                $defs:
                  content:
                    $dynamicAnchor: T
                    type: array
components:
  schemas:
    Response:
      $defs:
        content:
          $dynamicAnchor: T
          not: true
      type: object
      properties:
        content:
          $dynamicRef: '#T'
        error:
          type: object
`,
			test: func(t *testing.T, c *openapi.Config) {
				require.NotNil(t, c)
				s := c.Paths["/foo"].Value.Get.Responses.GetResponse(http.StatusOK).Content["application/json"].Schema
				require.NotNil(t, s)
				require.NotNil(t, s.Properties)
				require.Equal(t, json.Types{"object"}, s.Properties.Get("content").Type)

				bar := c.Paths["/bar"].Value.Get.Responses.GetResponse(http.StatusOK).Content["application/json"].Schema
				require.NotNil(t, bar)
				require.NotNil(t, bar.Properties)
				require.Equal(t, json.Types{"array"}, bar.Properties.Get("content").Type)
			},
		},
	}

	t.Parallel()
	for _, tc := range testdata {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := &openapi.Config{}
			err := yaml.Unmarshal([]byte(tc.data), c)
			require.NoError(t, err)
			err = c.Parse(&dynamic.Config{Data: c}, &dynamictest.Reader{})
			require.NoError(t, err)
			tc.test(t, c)
		})
	}
}

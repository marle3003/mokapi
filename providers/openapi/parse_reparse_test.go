package openapi_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/openapi"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

const specWithAnchor = `openapi: 3.1.0
info:
  title: anchors
  version: "1.0"
paths:
  /pets:
    get:
      parameters:
        - name: kind
          in: query
          schema:
            $anchor: kind
            type: string
      responses:
        "200":
          description: ok
`

// The ConfigWatcher parses the same *Config again when the file changes.
func TestParse_ReparseWithSchemaAnchor(t *testing.T) {
	u, err := url.Parse("file:///anchors.yml")
	require.NoError(t, err)
	c := dynamictest.NewConfig(&openapi.Config{},
		dynamictest.WithInfo(dynamictest.NewConfigInfo(dynamictest.WithUrl(u.String()))),
		dynamictest.WithRaw([]byte(specWithAnchor)))

	require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))
	op := c.Data().(*openapi.Config).Paths.Lookup("/pets").Value.Get
	require.Empty(t, op.Errors)

	// Operation errors are logged and stored on the operation, Parse itself returns nil.
	require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))
	op = c.Data().(*openapi.Config).Paths.Lookup("/pets").Value.Get
	require.Empty(t, op.Errors, "same content parsed again must not fail")
}

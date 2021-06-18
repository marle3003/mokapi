package openapi

import (
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/test"
	"testing"
)

type testReader struct {
}

func (tr *testReader) Read(path string, c dynamic.Config, h dynamic.ChangeEventHandler) error {
	return nil
}

func TestSchemaRef(t *testing.T) {
	s := `mokapi: 1.0
components:
  schemas:
    A:
      type: object
      properties:
        b:
          $ref: '#/components/schemas/C'
    C:
      type: string`
	var c *Config
	err := yaml.Unmarshal([]byte(s), &c)
	test.Ok(t, err)

	tr := &testReader{}

	r := refResolver{reader: tr, path: "", config: c, eh: dynamic.NewEmptyEventHandler(c)}
	err = r.resolveConfig()
	test.Ok(t, err)

	test.Assert(t, len(c.Components.Schemas.Value) == 2, "unexpected length of schemas")
	test.Assert(t, c.Components.Schemas.Value["A"].Value.Type == "object", "unexpected type")
	test.Assert(t, c.Components.Schemas.Value["A"].Value.Properties.Value["b"].Ref == "#/components/schemas/C", "unexpected ref")
	test.Assert(t, c.Components.Schemas.Value["A"].Value.Properties.Value["b"].Value != nil, "reference was not resolved")
	test.Assert(t, c.Components.Schemas.Value["A"].Value.Properties.Value["b"].Value == c.Components.Schemas.Value["C"].Value, "reference was not correct resolved")
}

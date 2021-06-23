package openapi

import (
	"mokapi/test"
	"testing"
)

var testExternalData = []testDataEntry{
	{
		"External SchemaRef",
		externalSchemaRef,
		func(t *testing.T, c *Config) {
			test.Assert(t, c.Components.Schemas.Value["A"].Value.Properties.Value["b"].Value != nil, "\"reference was not resolved")
			test.Assert(t, c.Components.Schemas.Value["A"].Value.Properties.Value["b"].Value.Type == "string", "expected type of string")
		},
		&testReader{data: map[string]string{
			"test.yml": externalSchemaRef2,
		}},
	},

	{
		"External Relative SchemaRef",
		externalRelativeSchemaRef,
		func(t *testing.T, c *Config) {
			test.Assert(t, c.Components.Schemas.Value["A"].Value.Properties.Value["b"].Value != nil, "\"reference was not resolved")
			test.Assert(t, c.Components.Schemas.Value["A"].Value.Properties.Value["b"].Value.Type == "string", "expected type of string")
		},
		&testReader{data: map[string]string{
			"test.yml": externalRelativeSchemaRef2,
		}},
	},

	{
		"External Schemas",
		externalSchemasRef,
		func(t *testing.T, c *Config) {
			test.Assert(t, c.Components.Schemas.Value["A"].Value != nil, "reference A was not resolved")
			test.Assert(t, c.Components.Schemas.Value["C"].Value != nil, "reference C was not resolved")
		},
		&testReader{data: map[string]string{
			"test.yml": externalSchemasRef2,
		}},
	},

	{
		"External Endpoint",
		externalEndpointRef,
		func(t *testing.T, c *Config) {
			test.Assert(t, c.EndPoints["/test"] != nil, "reference was not resolved")
			test.Equals(t, "Returns a list of users.", c.EndPoints["/test"].Value.Get.Summary)
		},
		&testReader{data: map[string]string{
			"test.yml": externalEndpointRef2,
		}},
	},
}

const externalSchemaRef = `
openapi: 3.0.0
components:
  schemas:
    A:
      type: object
      properties:
        b:
          $ref: 'test.yml#/components/schemas/C'
`
const externalSchemaRef2 = `
components:
  schemas:
    C:
      type: string
`

const externalRelativeSchemaRef = `
openapi: 3.0.0
components:
  schemas:
    A:
      type: object
      properties:
        b:
          $ref: 'test.yml#/C'
`
const externalRelativeSchemaRef2 = `
C:
  type: string
`

const externalSchemasRef = `
openapi: 3.0.0
components:
  schemas:
    $ref: 'test.yml'
`
const externalSchemasRef2 = `
A:
  type: string
C:
  type: string
`

const externalEndpointRef = `
openapi: 3.0.0
paths:
  /test:
    $ref: 'test.yml'
`
const externalEndpointRef2 = `
get:
  summary: Returns a list of users.
`

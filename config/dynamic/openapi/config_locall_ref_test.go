package openapi

import (
	"mokapi/test"
	"testing"
)

var testLocalData = []testDataEntry{
	{
		"Local SchemaRef",
		localSchemaRef,
		func(t *testing.T, c *Config) {
			test.Assert(t, len(c.Components.Schemas.Value) == 2, "unexpected length of schemas")
			test.Assert(t, c.Components.Schemas.Value["A"].Value.Type == "object", "unexpected type")
			test.Assert(t, c.Components.Schemas.Value["A"].Value.Properties.Value["b"].Ref == "#/components/schemas/C", "unexpected ref")
			test.Assert(t, c.Components.Schemas.Value["A"].Value.Properties.Value["b"].Value != nil, "reference was not resolved")
			test.Assert(t, c.Components.Schemas.Value["A"].Value.Properties.Value["b"].Value == c.Components.Schemas.Value["C"].Value, "reference was not correct resolved")
		},
		&testReader{},
	},

	{
		"Local ResponseRef",
		localResponseRef,
		func(t *testing.T, c *Config) {
			r := c.EndPoints["/test"].Value.Get.Responses[InternalServerError]
			test.Assert(t, r != nil, "reference nil")
			test.Assert(t, r.Value.Content["application/json"].Schema.Value.Type == "string", "expected type string")
		},
		&testReader{},
	},

	{
		"Local ParameterRef",
		localParameterRef,
		func(t *testing.T, c *Config) {
			params := c.EndPoints["/test"].Value.Get.Parameters
			test.Assert(t, len(params) == 1, "expected length of parameters 1, got %v", len(params))
			test.Assert(t, params[0].Value.Name == "session", "expected name of parameter session, got %q", params[0].Value.Name)
		},
		&testReader{},
	},

	{
		"Local ExampleRef",
		localExampleRef,
		func(t *testing.T, c *Config) {
			r := c.EndPoints["/test"].Value.Get.Responses[OK]
			examples := r.Value.Content["application/json"].Examples
			testObject := examples["testObject"].Value
			s := r.Value.Content["application/json"].Schema
			test.Equals(t, map[string]interface{}{"id": 1}, testObject.Value)
			test.Assert(t, s.Value.Type == "object", "expected type of object")
		},
		&testReader{},
	},

	{
		"Local RequestBodies",
		localRequestBodiesRef,
		func(t *testing.T, c *Config) {
			b := c.EndPoints["/test"].Value.Get.RequestBody
			s := b.Value.Content["application/json"].Schema
			test.Assert(t, b.Value != nil, "reference not resolved")
			test.Assert(t, s.Value.Type == "string", "expected type of string")
		},
		&testReader{},
	},

	{
		"Local Header",
		localHeadersRef,
		func(t *testing.T, c *Config) {
			r := c.EndPoints["/test"].Value.Get.Responses[OK]
			h := r.Value.Headers["X-Test"]
			test.Assert(t, h.Value != nil, "reference not resolved")
			test.Assert(t, h.Value.Schema.Value.Type == "string", "expected type of string")
		},
		&testReader{},
	},
}

const localSchemaRef = `
openapi: 3.0.0
components:
  schemas:
    A:
      type: object
      properties:
        b:
          $ref: '#/components/schemas/C'
    C:
      type: string`

const localResponseRef = `
openapi: 3.0.0
paths:
  /test:
    get:
      responses:
        '500':
          $ref: '#/components/responses/Error'
components:
  schemas:
    ErrorMessage:
      type: string
  responses:
    Error:
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorMessage'`

const localParameterRef = `
openapi: 3.0.0
paths:
  /test:
    get:
      parameters:
        - $ref: '#/components/parameters/sessionParam'
components:
  parameters:
    sessionParam:
      in: cookie
      name: session
      schema:
        type: string`

const localExampleRef = `
openapi: 3.0.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MyObject'
              examples:
                testObject:
                  $ref: '#/components/examples/TestObject'
components:
  schemas:
    MyObject:
      type: object
      parameter:
        id:
          type: integer
  examples:
    TestObject:
      value:
        id: 1
      summary: A sample object`

const localRequestBodiesRef = `
openapi: 3.0.0
paths:
  /test:
    get:
      requestBody:
        $ref: '#/components/requestBodies/TestBody'
components:
  schemas:
    TestBody:
      type: string
  requestBodies:
    TestBody:
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/TestBody'
`

const localHeadersRef = `
openapi: 3.0.0
paths:
  /test:
    get:
      responses:
        '200':
          headers:
            X-Test:
              $ref: '#/components/headers/TestHeader'
components:
  headers:
    TestHeader:
      schema:
        type: string
`

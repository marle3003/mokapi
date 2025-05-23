package openapi_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/media"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/parameter"
	"mokapi/providers/openapi/schema/schematest"
	"net/http"
	"net/url"
	"testing"
)

func TestResolve(t *testing.T) {
	t.Run("empty should not error", func(t *testing.T) {
		reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) { return nil, nil })
		config := &openapi.Config{}
		err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
		require.NoError(t, err)
	})
}

func TestConfig_Validate(t *testing.T) {
	testdata := []struct {
		name string
		data string
		test func(t *testing.T, err error)
	}{
		{
			name: "simple valid config 3",
			data: `
openapi: 3
info:
  title: foo
`,
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "simple valid config 3.0",
			data: `
openapi: 3.0
info:
  title: foo
`,
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "simple valid config 3.1",
			data: `
openapi: 3.1
info:
  title: foo
`,
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "simple valid config 2",
			data: `
openapi: 2
info:
  title: foo
`,
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "not supported version: 2.0.0")
			},
		},
		{
			name: "missing title in 2.0",
			data: `
openapi: 2.0
info:
`,
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "not supported version: 2.0.0\nan openapi title is required")
			},
		},
		{
			name: "missing title in 3.0",
			data: `
openapi: 3.0.3
info:
`,
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "an openapi title is required")
			},
		},
		{
			name: "simple valid config 3.f",
			data: `
openapi: 3.f
info:
  title: foo
`,
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "no version",
			data: `
openapi: ""
info:
  title: foo
`,
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "no OpenApi version defined")
			},
		},
		{
			name: "path has no prefix '/'",
			data: `
openapi: "3.1"
info:
  title: foo
paths:
  foo: {}
`,
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "should only have path names that start with `/`: 'foo'")
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
			tc.test(t, c.Validate())
		})
	}
}

func TestResponses(t *testing.T) {
	testdata := []struct {
		name    string
		content string
		fn      func(t *testing.T, c *openapi.Responses)
	}{
		{
			name: "default httpstatus",
			content: `
default: {}
`,
			fn: func(t *testing.T, res *openapi.Responses) {
				r := res.GetResponse(200)
				require.NotNil(t, r)
			},
		},
		{
			name: "httpstatus",
			content: `
401: {}
`,
			fn: func(t *testing.T, res *openapi.Responses) {
				r := res.GetResponse(401)
				require.NotNil(t, r)
			},
		},
		{
			name: "not defined",
			content: `
401: {}
`,
			fn: func(t *testing.T, res *openapi.Responses) {
				r := res.GetResponse(200)
				require.Nil(t, r)
			},
		},
		{
			name: "not defined using default",
			content: `
401: {}
default: {description: default}
`,
			fn: func(t *testing.T, res *openapi.Responses) {
				r := res.GetResponse(200)
				require.NotNil(t, r)
				require.Equal(t, "default", r.Description)
			},
		},
	}
	t.Parallel()
	for _, data := range testdata {
		d := data
		t.Run(d.name, func(t *testing.T) {
			res := &openapi.Responses{}
			err := yaml.Unmarshal([]byte(d.content), res)
			require.NoError(t, err)
			data.fn(t, res)
		})
	}
}

func TestConfig(t *testing.T) {
	testdata := []struct {
		Name    string
		Content string
		f       func(t *testing.T, c *openapi.Config)
	}{
		{
			Name: "component schema",
			Content: `
openapi: 3.0.0
components:
  schemas:
    Foo:
      type: string
`,
			f: func(t *testing.T, c *openapi.Config) {
				require.Equal(t, 1, c.Components.Schemas.Len())
				foo := c.Components.Schemas.Get("Foo")
				require.Equal(t, "string", foo.Type.String())
			},
		},
		{
			Name: "Responses",
			Content: `
openapi: 3.0.0
info:
  title: foo
paths:
  /foo:
    get:
      responses:
        '204':
          description: no content
        '200':
          description: ok
          content:
            application/xml:
              schema:
                $ref: '#/components/schemas/Foo'
components:
  schemas:
    Foo:
      type: string
`,
			f: func(t *testing.T, c *openapi.Config) {
				require.NoError(t, c.Validate())
				require.Len(t, c.Paths, 1)
				exp := []string{"204", "200"}
				keys := c.Paths["/foo"].Value.Get.Responses.Keys()
				require.Equal(t, exp, keys)
				r := c.Paths["/foo"].Value.Get.Responses.GetResponse(http.StatusOK)
				content := r.Content["application/xml"]
				require.NotNil(t, content.Schema, "ref resolved")
			},
		},
	}

	for _, data := range testdata {
		d := data
		t.Run(d.Name, func(t *testing.T) {
			c := &openapi.Config{}
			err := yaml.Unmarshal([]byte(d.Content), c)
			require.NoError(t, err)
			err = c.Parse(&dynamic.Config{Data: c}, nil)
			require.NoError(t, err)
			d.f(t, c)
		})
	}
}

func TestConfig_PetStore_PetSchema(t *testing.T) {
	config := &openapi.Config{}
	err := yaml.Unmarshal([]byte(petstore), &config)
	require.NoError(t, err)
	err = config.Parse(&dynamic.Config{Data: config}, nil)
	require.NoError(t, err)

	require.Nil(t, config.Components.Schemas.Get("foo"))

	pet := config.Components.Schemas.Get("Pet")
	require.Equal(t, []string{"name", "photoUrls"}, pet.Required)
	require.Equal(t, "object", pet.Type.String())
	require.Equal(t, "Pet", pet.Xml.Name)

	// id
	id := pet.Properties.Get("id")
	require.Equal(t, "integer", id.Type.String())
	require.Equal(t, "int64", id.Format)

	// category
	category := pet.Properties.Get("category")
	require.Equal(t, "#/components/schemas/Category", category.Ref)
	require.NotNil(t, category, "ref resolved")
	require.Equal(t, "object", category.Type.String())

	// name
	name := pet.Properties.Get("name")
	require.Equal(t, "string", name.Type.String())
	require.Equal(t, "doggie", name.Example.Value)

	// photoUrls
	photoUrls := pet.Properties.Get("photoUrls")
	require.Equal(t, "array", photoUrls.Type.String())
	require.Equal(t, "string", photoUrls.Items.Type.String())
	require.Equal(t, "photoUrl", photoUrls.Xml.Name)
	require.True(t, photoUrls.Xml.Wrapped)

	// tags
	tags := pet.Properties.Get("tags")
	require.Equal(t, "array", tags.Type.String())
	require.Equal(t, "#/components/schemas/Tag", tags.Items.Ref)
	require.Equal(t, "object", tags.Items.Type.String())
	require.Equal(t, "tag", tags.Xml.Name)
	require.True(t, tags.Xml.Wrapped)

	// status
	status := pet.Properties.Get("status")
	require.Equal(t, "string", status.Type.String())
	require.Equal(t, "pet status in the store", status.Description)
	require.Equal(t, []interface{}{"available", "pending", "sold"}, status.Enum)

}

func TestConfig_PetStore_Path(t *testing.T) {
	config := &openapi.Config{}
	err := yaml.Unmarshal([]byte(petstore), &config)
	require.NoError(t, err)
	err = config.Parse(&dynamic.Config{Data: config}, nil)
	require.NoError(t, err)

	endpoint := config.Paths["/pet"]
	require.NotNil(t, endpoint.Value.Put, "put is defined")
	require.NotNil(t, endpoint.Value.Post, "post is defined")
	put := endpoint.Value.Put
	require.Equal(t, "Update an existing pet", put.Summary)
	require.Equal(t, "updatePet", put.OperationId)
	body := put.RequestBody.Value
	require.Equal(t, "Pet object that needs to be added to the store", body.Description)
	require.Len(t, body.Content, 2)
	require.Equal(t, "#/components/schemas/Pet", body.Content["application/json"].Schema.Ref)
	require.Equal(t, "#/components/schemas/Pet", body.Content["application/xml"].Schema.Ref)
	require.NotNil(t, body.Content["application/json"].Schema, "ref resolved")
	require.NotNil(t, body.Content["application/xml"].Schema, "ref resolved")
	require.Equal(t, body.Content["application/json"], body.GetMedia(media.ParseContentType("application/json")))
	require.Nil(t, body.GetMedia(media.ParseContentType("foo/bar")))

	schema := body.Content["application/json"].Schema
	require.Equal(t, []string{"name", "photoUrls"}, schema.Required)

	require.True(t, put.Responses.Len() == 3)
	r := put.Responses.GetResponse(http.StatusBadRequest)
	require.Len(t, r.Content, 0)

}

func TestPetStore_Response(t *testing.T) {
	config := &openapi.Config{}
	err := yaml.Unmarshal([]byte(petstore), &config)
	require.NoError(t, err)
	err = config.Parse(&dynamic.Config{Data: config}, nil)
	require.NoError(t, err)

	endpoint := config.Paths["/pet/{petId}"]
	r := endpoint.Value.Get.Responses.GetResponse(http.StatusOK)
	require.NotNil(t, r, "response exists")
	ct := media.ParseContentType("application/json")
	m := r.GetContent(ct)
	require.Equal(t, r.Content[ct.String()], m)

	m = r.GetContent(media.ParseContentType("foo/bar"))
	require.Nil(t, m)
}

func TestPetStore_Paramters(t *testing.T) {
	config := &openapi.Config{}
	err := yaml.Unmarshal([]byte(petstore), &config)
	require.NoError(t, err)
	err = config.Parse(&dynamic.Config{Data: config}, nil)
	require.NoError(t, err)

	endpoint := config.Paths["/pet/{petId}"]
	params := endpoint.Value.Delete.Parameters
	require.Len(t, params, 2)
	require.Equal(t, "api_key", params[0].Value.Name)
	require.Equal(t, parameter.Header, params[0].Value.Type)
	require.Equal(t, "string", params[0].Value.Schema.Type.String())

	require.Equal(t, "petId", params[1].Value.Name)
	require.Equal(t, parameter.Path, params[1].Value.Type)
	require.Equal(t, "Pet id to delete", params[1].Value.Description)
	require.True(t, params[1].Value.Required)
	require.Equal(t, "integer", params[1].Value.Schema.Type.String())
	require.Equal(t, "int64", params[1].Value.Schema.Format)
}

func TestHttpStatus_IsSuccess(t *testing.T) {
	for _, d := range []int{
		200,
		201,
		202,
		203,
		204,
		205,
		206,
		207,
		208,
		226,
	} {
		t.Run(fmt.Sprintf("%v", d), func(t *testing.T) {
			require.True(t, openapi.IsHttpStatusSuccess(d))
		})
	}
	require.False(t, openapi.IsHttpStatusSuccess(http.StatusBadGateway))
}

func TestConfig_Patch(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "patch add path on empty",
			configs: []*openapi.Config{
				{},
				openapitest.NewConfig("1.0", openapitest.WithPath("/foo", openapitest.NewPath())),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths, 1)
				require.Contains(t, result.Paths, "/foo")
			},
		},
		{
			name: "patch add component schema on empty",
			configs: []*openapi.Config{
				{},
				openapitest.NewConfig("1.0", openapitest.WithComponentSchema("Foo", schematest.New("string"))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, 1, result.Components.Schemas.Len())
				require.NotNil(t, result.Components.Schemas.Get("Foo"))
				require.Equal(t, "string", result.Components.Schemas.Get("Foo").Type.String())
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := tc.configs[0]
			for _, p := range tc.configs[1:] {
				c.Patch(p)
			}
			tc.test(t, c)
		})
	}
}

const petstore = `
openapi: 3.0.1
info:
  title: Swagger Petstore
  description: 'This is a sample server Petstore server.  You can find out more about     Swagger
    at [http://swagger.io](http://swagger.io) or on [irc.freenode.net, #swagger](http://swagger.io/irc/).      For
    this sample, you can use the api key "special-key"" to test the authorization     filters.'
  termsOfService: http://swagger.io/terms/
  contact:
    email: apiteam@swagger.io
  license:
    name: Apache 2.0
    url: http://www.apache.org/licenses/LICENSE-2.0.html
  version: 1.0.0
externalDocs:
  description: Find out more about Swagger
  url: http://swagger.io
servers:
  - url: https://127.0.0.1
  - url: http://127.0.0.1
tags:
  - name: pet
    description: Everything about your Pets
    externalDocs:
      description: Find out more
      url: http://swagger.io
  - name: store
    description: Access to Petstore orders
  - name: user
    description: Operations about user
    externalDocs:
      description: Find out more about our store
      url: http://swagger.io
paths:
  /pet:
    put:
      tags:
        - pet
      summary: Update an existing pet
      operationId: updatePet
      requestBody:
        description: Pet object that needs to be added to the store
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Pet'
          application/xml:
            schema:
              $ref: '#/components/schemas/Pet'
        required: true
      responses:
        400:
          description: Invalid ID supplied
          content: {}
        404:
          description: Pet not found
          content: {}
        405:
          description: Validation exception
          content: {}
      security:
        - petstore_auth:
            - write:pets
            - read:pets
      x-codegen-request-body-name: body
    post:
      tags:
        - pet
      summary: Add a new pet to the store
      operationId: addPet
      requestBody:
        description: Pet object that needs to be added to the store
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Pet'
          application/xml:
            schema:
              $ref: '#/components/schemas/Pet'
        required: true
      responses:
        405:
          description: Invalid input
          content: {}
      security:
        - petstore_auth:
            - write:pets
            - read:pets
      x-codegen-request-body-name: body
  /pet/findByStatus:
    get:
      tags:
        - pet
      summary: Finds Pets by status
      description: Multiple status values can be provided with comma separated strings
      operationId: findPetsByStatus
      parameters:
        - name: status
          in: query
          description: Status values that need to be considered for filter
          required: true
          style: form
          explode: true
          schema:
            type: array
            items:
              type: string
              default: available
              enum:
                - available
                - pending
                - sold
      responses:
        200:
          description: successful operation
          content:
            application/xml:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Pet'
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Pet'
        400:
          description: Invalid status value
          content: {}
      security:
        - petstore_auth:
            - write:pets
            - read:pets
  /pet/findByTags:
    get:
      tags:
        - pet
      summary: Finds Pets by tags
      description: Muliple tags can be provided with comma separated strings. Use         tag1,
        tag2, tag3 for testing.
      operationId: findPetsByTags
      parameters:
        - name: tags
          in: query
          description: Tags to filter by
          required: true
          style: form
          explode: true
          schema:
            type: array
            items:
              type: string
      responses:
        200:
          description: successful operation
          content:
            application/xml:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Pet'
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Pet'
        400:
          description: Invalid tag value
          content: {}
      deprecated: true
      security:
        - petstore_auth:
            - write:pets
            - read:pets
  /pet/{petId}:
    get:
      tags:
        - pet
      summary: Find pet by ID
      description: Returns a single pet
      operationId: getPetById
      parameters:
        - name: petId
          in: path
          description: ID of pet to return
          required: true
          schema:
            type: integer
            format: int64
      responses:
        200:
          description: successful operation
          content:
            application/xml:
              schema:
                $ref: '#/components/schemas/Pet'
            application/json:
              schema:
                $ref: '#/components/schemas/Pet'
        400:
          description: Invalid ID supplied
          content: {}
        404:
          description: Pet not found
          content: {}
      security:
        - api_key: []
    post:
      tags:
        - pet
      summary: Updates a pet in the store with form data
      operationId: updatePetWithForm
      parameters:
        - name: petId
          in: path
          description: ID of pet that needs to be updated
          required: true
          schema:
            type: integer
            format: int64
      requestBody:
        content:
          application/x-www-form-urlencoded:
            schema:
              properties:
                name:
                  type: string
                  description: Updated name of the pet
                status:
                  type: string
                  description: Updated status of the pet
      responses:
        405:
          description: Invalid input
          content: {}
      security:
        - petstore_auth:
            - write:pets
            - read:pets
    delete:
      tags:
        - pet
      summary: Deletes a pet
      operationId: deletePet
      parameters:
        - name: api_key
          in: header
          schema:
            type: string
        - name: petId
          in: path
          description: Pet id to delete
          required: true
          schema:
            type: integer
            format: int64
      responses:
        400:
          description: Invalid ID supplied
          content: {}
        404:
          description: Pet not found
          content: {}
      security:
        - petstore_auth:
            - write:pets
            - read:pets
  /pet/{petId}/uploadImage:
    post:
      tags:
        - pet
      summary: uploads an image
      operationId: uploadFile
      parameters:
        - name: petId
          in: path
          description: ID of pet to update
          required: true
          schema:
            type: integer
            format: int64
      requestBody:
        content:
          multipart/form-data:
            schema:
              properties:
                additionalMetadata:
                  type: string
                  description: Additional data to pass to server
                file:
                  type: string
                  description: file to upload
                  format: binary
      responses:
        200:
          description: successful operation
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ApiResponse'
      security:
        - petstore_auth:
            - write:pets
            - read:pets
  /store/inventory:
    get:
      tags:
        - store
      summary: Returns pet inventories by status
      description: Returns a map of status codes to quantities
      operationId: getInventory
      responses:
        200:
          description: successful operation
          content:
            application/json:
              schema:
                type: object
                additionalProperties:
                  type: integer
                  format: int32
      security:
        - api_key: []
  /store/order:
    post:
      tags:
        - store
      summary: Place an order for a pet
      operationId: placeOrder
      requestBody:
        description: order placed for purchasing the pet
        content:
          '*/*':
            schema:
              $ref: '#/components/schemas/Order'
        required: true
      responses:
        200:
          description: successful operation
          content:
            application/xml:
              schema:
                $ref: '#/components/schemas/Order'
            application/json:
              schema:
                $ref: '#/components/schemas/Order'
        400:
          description: Invalid Order
          content: {}
      x-codegen-request-body-name: body
  /store/order/{orderId}:
    get:
      tags:
        - store
      summary: Find purchase order by ID
      description: For valid response try integer IDs with value >= 1 and <= 10.         Other
        values will generated exceptions
      operationId: getOrderById
      parameters:
        - name: orderId
          in: path
          description: ID of pet that needs to be fetched
          required: true
          schema:
            maximum: 10.0
            minimum: 1.0
            type: integer
            format: int64
      responses:
        200:
          description: successful operation
          content:
            application/xml:
              schema:
                $ref: '#/components/schemas/Order'
            application/json:
              schema:
                $ref: '#/components/schemas/Order'
        400:
          description: Invalid ID supplied
          content: {}
        404:
          description: Order not found
          content: {}
    delete:
      tags:
        - store
      summary: Delete purchase order by ID
      description: For valid response try integer IDs with positive integer value.         Negative
        or non-integer values will generate API errors
      operationId: deleteOrder
      parameters:
        - name: orderId
          in: path
          description: ID of the order that needs to be deleted
          required: true
          schema:
            minimum: 1.0
            type: integer
            format: int64
      responses:
        400:
          description: Invalid ID supplied
          content: {}
        404:
          description: Order not found
          content: {}
  /user:
    post:
      tags:
        - user
      summary: Create user
      description: This can only be done by the logged in user.
      operationId: createUser
      requestBody:
        description: Created user object
        content:
          '*/*':
            schema:
              $ref: '#/components/schemas/User'
        required: true
      responses:
        default:
          description: successful operation
          content: {}
      x-codegen-request-body-name: body
  /user/createWithArray:
    post:
      tags:
        - user
      summary: Creates list of users with given input array
      operationId: createUsersWithArrayInput
      requestBody:
        description: List of user object
        content:
          '*/*':
            schema:
              type: array
              items:
                $ref: '#/components/schemas/User'
        required: true
      responses:
        default:
          description: successful operation
          content: {}
      x-codegen-request-body-name: body
  /user/createWithList:
    post:
      tags:
        - user
      summary: Creates list of users with given input array
      operationId: createUsersWithListInput
      requestBody:
        description: List of user object
        content:
          '*/*':
            schema:
              type: array
              items:
                $ref: '#/components/schemas/User'
        required: true
      responses:
        default:
          description: successful operation
          content: {}
      x-codegen-request-body-name: body
  /user/login:
    get:
      tags:
        - user
      summary: Logs user into the system
      operationId: loginUser
      parameters:
        - name: username
          in: query
          description: The user name for login
          required: true
          schema:
            type: string
        - name: password
          in: query
          description: The password for login in clear text
          required: true
          schema:
            type: string
      responses:
        200:
          description: successful operation
          headers:
            X-Rate-Limit:
              description: calls per hour allowed by the user
              schema:
                type: integer
                format: int32
            X-Expires-After:
              description: date in UTC when token expires
              schema:
                type: string
                format: date-time
          content:
            application/xml:
              schema:
                type: string
            application/json:
              schema:
                type: string
        400:
          description: Invalid username/password supplied
          content: {}
  /user/logout:
    get:
      tags:
        - user
      summary: Logs out current logged in user session
      operationId: logoutUser
      responses:
        default:
          description: successful operation
          content: {}
  /user/{username}:
    get:
      tags:
        - user
      summary: Get user by user name
      operationId: getUserByName
      parameters:
        - name: username
          in: path
          description: 'The name that needs to be fetched. Use user1 for testing. '
          required: true
          schema:
            type: string
      responses:
        200:
          description: successful operation
          content:
            application/xml:
              schema:
                $ref: '#/components/schemas/User'
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        400:
          description: Invalid username supplied
          content: {}
        404:
          description: User not found
          content: {}
    put:
      tags:
        - user
      summary: Updated user
      description: This can only be done by the logged in user.
      operationId: updateUser
      parameters:
        - name: username
          in: path
          description: name that need to be updated
          required: true
          schema:
            type: string
      requestBody:
        description: Updated user object
        content:
          '*/*':
            schema:
              $ref: '#/components/schemas/User'
        required: true
      responses:
        400:
          description: Invalid user supplied
          content: {}
        404:
          description: User not found
          content: {}
      x-codegen-request-body-name: body
    delete:
      tags:
        - user
      summary: Delete user
      description: This can only be done by the logged in user.
      operationId: deleteUser
      parameters:
        - name: username
          in: path
          description: The name that needs to be deleted
          required: true
          schema:
            type: string
      responses:
        400:
          description: Invalid username supplied
          content: {}
        404:
          description: User not found
          content: {}
components:
  schemas:
    Order:
      type: object
      properties:
        id:
          type: integer
          format: int64
        petId:
          type: integer
          format: int64
        quantity:
          type: integer
          format: int32
        shipDate:
          type: string
          format: date-time
        status:
          type: string
          description: Order Status
          enum:
            - placed
            - approved
            - delivered
        complete:
          type: boolean
          default: false
      xml:
        name: Order
    Category:
      type: object
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
      xml:
        name: Category
    User:
      type: object
      properties:
        id:
          type: integer
          format: int64
        username:
          type: string
        firstName:
          type: string
        lastName:
          type: string
        email:
          type: string
        password:
          type: string
        phone:
          type: string
        userStatus:
          type: integer
          description: User Status
          format: int32
      xml:
        name: User
    Tag:
      type: object
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
      xml:
        name: Tag
    Pet:
      required:
        - name
        - photoUrls
      type: object
      properties:
        id:
          type: integer
          format: int64
        category:
          $ref: '#/components/schemas/Category'
        name:
          type: string
          example: doggie
        photoUrls:
          type: array
          xml:
            name: photoUrl
            wrapped: true
          items:
            type: string
        tags:
          type: array
          xml:
            name: tag
            wrapped: true
          items:
            $ref: '#/components/schemas/Tag'
        status:
          type: string
          description: pet status in the store
          enum:
            - available
            - pending
            - sold
      xml:
        name: Pet
    ApiResponse:
      type: object
      properties:
        code:
          type: integer
          format: int32
        type:
          type: string
        message:
          type: string
  securitySchemes:
    petstore_auth:
      type: oauth2
      flows:
        implicit:
          authorizationUrl: http://petstore.swagger.io/oauth/dialog
          scopes:
            write:pets: modify pets in your account
            read:pets: read your pets
    api_key:
      type: apiKey
      name: api_key
      in: header

`

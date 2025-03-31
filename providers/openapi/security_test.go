package openapi_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/providers/openapi"
	"testing"
)

func TestConfig_Security_Yaml(t *testing.T) {
	testdata := []struct {
		Name    string
		Content string
		f       func(t *testing.T, c *openapi.Config)
	}{
		{
			Name: "bearer scheme",
			Content: `
openapi: 3.0.0
components:
  securitySchemes:
    foo:
      type: http
      scheme: bearer
`,
			f: func(t *testing.T, c *openapi.Config) {
				require.Len(t, c.Components.SecuritySchemes, 1)
				scheme := c.Components.SecuritySchemes["foo"]
				require.IsType(t, &openapi.HttpSecurityScheme{}, scheme)
				http := scheme.(*openapi.HttpSecurityScheme)
				require.Equal(t, "http", http.Type)
				require.Equal(t, "bearer", http.Scheme)
			},
		},
		{
			Name: "unknown scheme",
			Content: `
openapi: 3.0.0
components:
  securitySchemes:
    foo:
      type: foo
`,
			f: func(t *testing.T, c *openapi.Config) {
				require.Len(t, c.Components.SecuritySchemes, 1)
				scheme := c.Components.SecuritySchemes["foo"]
				require.IsType(t, &openapi.NotSupportedSecurityScheme{}, scheme)
			},
		},
		{
			Name: "basic scheme",
			Content: `
openapi: 3.0.0
components:
  securitySchemes:
    foo:
      type: http
      scheme: basic
`,
			f: func(t *testing.T, c *openapi.Config) {
				require.Len(t, c.Components.SecuritySchemes, 1)
				scheme := c.Components.SecuritySchemes["foo"]
				require.IsType(t, &openapi.HttpSecurityScheme{}, scheme)
				http := scheme.(*openapi.HttpSecurityScheme)
				require.Equal(t, "http", http.Type)
				require.Equal(t, "basic", http.Scheme)
			},
		},
		{
			Name: "api key scheme",
			Content: `
openapi: 3.0.0
components:
  securitySchemes:
    foo:
      type: apiKey
      in: header
      name: X-API-KEY
`,
			f: func(t *testing.T, c *openapi.Config) {
				require.Len(t, c.Components.SecuritySchemes, 1)
				scheme := c.Components.SecuritySchemes["foo"]
				require.IsType(t, &openapi.ApiKeySecurityScheme{}, scheme)
				apiKey := scheme.(*openapi.ApiKeySecurityScheme)
				require.Equal(t, "apiKey", apiKey.Type)
				require.Equal(t, "header", apiKey.In)
				require.Equal(t, "X-API-KEY", apiKey.Name)
			},
		},
		{
			Name: "oauth2 scheme",
			Content: `
openapi: 3.0.0
components:
  securitySchemes:
    foo:
      type: oauth2
      description: foobar
      flows:
        implicit:
          authorizationUrl: https://example.com/oauth2
          scopes:
            read_pets: read your pets
            write_pets: modify pets in your account
`,
			f: func(t *testing.T, c *openapi.Config) {
				require.Len(t, c.Components.SecuritySchemes, 1)
				scheme := c.Components.SecuritySchemes["foo"]
				require.IsType(t, &openapi.OAuth2SecurityScheme{}, scheme)
				oauth2 := scheme.(*openapi.OAuth2SecurityScheme)
				require.Equal(t, "oauth2", oauth2.Type)
				require.Equal(t, "foobar", oauth2.Description)
				require.Len(t, oauth2.Flows, 1)
				require.Contains(t, oauth2.Flows, "implicit")
				require.Equal(t, "https://example.com/oauth2", oauth2.Flows["implicit"].AuthorizationUrl)
				require.Equal(t, map[string]string{"write_pets": "modify pets in your account", "read_pets": "read your pets"}, oauth2.Flows["implicit"].Scopes)
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

func TestConfig_Security_Json(t *testing.T) {
	testdata := []struct {
		Name    string
		Content string
		f       func(t *testing.T, c *openapi.Config)
	}{
		{
			Name: "bearer scheme",
			Content: `{
"openapi": "3.0.0",
"components": {
  "securitySchemes": {
    "foo": {
      "type": "http",
      "scheme": "bearer"
    }
  }
}
}`,
			f: func(t *testing.T, c *openapi.Config) {
				require.Len(t, c.Components.SecuritySchemes, 1)
				scheme := c.Components.SecuritySchemes["foo"]
				require.IsType(t, &openapi.HttpSecurityScheme{}, scheme)
				http := scheme.(*openapi.HttpSecurityScheme)
				require.Equal(t, "http", http.Type)
				require.Equal(t, "bearer", http.Scheme)
			},
		},
		{
			Name: "unknown scheme",
			Content: `{
"openapi": "3.0.0",
"components": {
  "securitySchemes": {
    "foo": {
      "type": "foo"
    }
  }
}
}`,
			f: func(t *testing.T, c *openapi.Config) {
				require.Len(t, c.Components.SecuritySchemes, 1)
				scheme := c.Components.SecuritySchemes["foo"]
				require.IsType(t, &openapi.NotSupportedSecurityScheme{}, scheme)
			},
		},
		{
			Name: "basic scheme",
			Content: `{
"openapi": "3.0.0",
"components": {
  "securitySchemes": {
    "foo": {
      "type": "http",
      "scheme": "basic"
    }
  }
}
}`,
			f: func(t *testing.T, c *openapi.Config) {
				require.Len(t, c.Components.SecuritySchemes, 1)
				scheme := c.Components.SecuritySchemes["foo"]
				require.IsType(t, &openapi.HttpSecurityScheme{}, scheme)
				http := scheme.(*openapi.HttpSecurityScheme)
				require.Equal(t, "http", http.Type)
				require.Equal(t, "basic", http.Scheme)
			},
		},
		{
			Name: "api key scheme",
			Content: `{
"openapi": "3.0.0",
"components": {
  "securitySchemes": {
    "foo": {
      "type": "apiKey",
      "in": "header",
      "name": "X-API-KEY"
    }
  }
}
}`,
			f: func(t *testing.T, c *openapi.Config) {
				require.Len(t, c.Components.SecuritySchemes, 1)
				scheme := c.Components.SecuritySchemes["foo"]
				require.IsType(t, &openapi.ApiKeySecurityScheme{}, scheme)
				apiKey := scheme.(*openapi.ApiKeySecurityScheme)
				require.Equal(t, "apiKey", apiKey.Type)
				require.Equal(t, "header", apiKey.In)
				require.Equal(t, "X-API-KEY", apiKey.Name)
			},
		},
		{
			Name: "no error when using oauth",
			Content: `{
"openapi": "3.0.0",
"components": {
  "securitySchemes": {
    "foo": {
      "type": "oauth2",
      "description": "foobar",
      "flows": {
        "implicit": {
          "authorizationUrl": "https://example.com/oauth2",
          "scopes": {
            "read_pets": "read your pets",
            "write_pets": "modify pets in your account"
         }
       }
     }
    }
  }
}
}`,
			f: func(t *testing.T, c *openapi.Config) {
				require.Len(t, c.Components.SecuritySchemes, 1)
				scheme := c.Components.SecuritySchemes["foo"]
				require.IsType(t, &openapi.OAuth2SecurityScheme{}, scheme)
				oauth2 := scheme.(*openapi.OAuth2SecurityScheme)
				require.Equal(t, "oauth2", oauth2.Type)
				require.Equal(t, "foobar", oauth2.Description)
				require.Len(t, oauth2.Flows, 1)
				require.Contains(t, oauth2.Flows, "implicit")
				require.Equal(t, "https://example.com/oauth2", oauth2.Flows["implicit"].AuthorizationUrl)
				require.Equal(t, map[string]string{"write_pets": "modify pets in your account", "read_pets": "read your pets"}, oauth2.Flows["implicit"].Scopes)
			},
		},
	}

	for _, data := range testdata {
		d := data
		t.Run(d.Name, func(t *testing.T) {
			c := &openapi.Config{}
			err := json.Unmarshal([]byte(d.Content), c)
			require.NoError(t, err)
			err = c.Parse(&dynamic.Config{Data: c}, nil)
			require.NoError(t, err)
			d.f(t, c)
		})
	}
}

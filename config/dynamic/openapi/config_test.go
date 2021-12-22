package openapi

import (
	"gopkg.in/yaml.v3"
	"mokapi/test"
	"testing"
)

func TestConfig(t *testing.T) {
	testdata := []struct {
		Name    string
		Content string
		f       func(t *testing.T, c *Config)
	}{
		{
			Name: "Responses",
			Content: `
openapi: 3.0.0
paths:
  /foo:
    get:
      responses:
        '204':
          description: no content
        '200':
          description: ok
`,
			f: func(t *testing.T, c *Config) {
				test.Equals(t, 1, len(c.EndPoints))
				exp := []interface{}{HttpStatus(204), HttpStatus(200)}
				keys := c.EndPoints["/foo"].Value.Get.Responses.Keys()
				test.Equals(t, exp, keys)
			},
		},
	}

	for _, data := range testdata {
		d := data
		t.Run(d.Name, func(t *testing.T) {
			c := &Config{}
			err := yaml.Unmarshal([]byte(d.Content), c)
			test.Ok(t, err)
			d.f(t, c)
		})
	}
}

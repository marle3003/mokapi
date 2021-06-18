package encoding

import (
	"fmt"
	"mokapi/config/dynamic/openapi"
	"mokapi/models/media"
	"mokapi/test"
	"testing"
)

func TestParse(t *testing.T) {
	data := []struct {
		s      string
		schema *openapi.SchemaRef
		e      interface{}
	}{
		{
			`{"test": 12}`,
			nil,
			map[string]interface{}{"test": 12.0},
		},
		{
			`{"test": 12}`,
			&openapi.SchemaRef{
				Value: &openapi.Schema{
					Type: "object",
					Properties: &openapi.Schemas{
						Value: map[string]*openapi.SchemaRef{
							"test": {
								Value: &openapi.Schema{
									Type: "integer",
								},
							},
						},
					},
				},
			},
			map[string]interface{}{"test": int64(12)},
		},
		{
			`{"test": "hello world"}`,
			&openapi.SchemaRef{
				Value: &openapi.Schema{
					Type: "object",
					Properties: &openapi.Schemas{
						Value: map[string]*openapi.SchemaRef{
							"test": {
								Value: &openapi.Schema{
									Type: "string",
								},
							},
						},
					},
				},
			},
			map[string]interface{}{"test": "hello world"},
		},
		{
			`{"test": "2021-01-20"}`,
			&openapi.SchemaRef{
				Value: &openapi.Schema{
					Type: "object",
					Properties: &openapi.Schemas{
						Value: map[string]*openapi.SchemaRef{
							"test": {
								Value: &openapi.Schema{
									Type:   "string",
									Format: "date",
								},
							},
						},
					},
				},
			},
			map[string]interface{}{"test": "2021-01-20"},
		},
		{
			`{"test": "2021-01-20T12:41:45.14Z"}`,
			&openapi.SchemaRef{
				Value: &openapi.Schema{
					Type: "object",
					Properties: &openapi.Schemas{
						Value: map[string]*openapi.SchemaRef{
							"test": {
								Value: &openapi.Schema{
									Type:   "string",
									Format: "date-time",
								},
							},
						},
					},
				},
			},
			map[string]interface{}{"test": "2021-01-20T12:41:45.14Z"},
		},
		{
			`{"test": true}`,
			&openapi.SchemaRef{
				Value: &openapi.Schema{
					Type: "object",
					Properties: &openapi.Schemas{
						Value: map[string]*openapi.SchemaRef{
							"test": {
								Value: &openapi.Schema{
									Type: "boolean",
								},
							},
						},
					},
				},
			},
			map[string]interface{}{"test": true},
		},
		{
			`{"test": ""}`,
			&openapi.SchemaRef{
				Value: &openapi.Schema{
					Type: "object",
					Properties: &openapi.Schemas{
						Value: map[string]*openapi.SchemaRef{
							"test": {
								Value: &openapi.Schema{
									Type: "boolean",
								},
							},
						},
					},
				},
			},
			map[string]interface{}{"test": false},
		},
		{
			`{"test": ["a", "b", "c"]}`,
			&openapi.SchemaRef{
				Value: &openapi.Schema{
					Type: "object",
					Properties: &openapi.Schemas{
						Value: map[string]*openapi.SchemaRef{
							"test": {
								Value: &openapi.Schema{
									Type: "array",
									Items: &openapi.SchemaRef{
										Value: &openapi.Schema{
											Type: "string",
										},
									},
								},
							},
						},
					},
				},
			},
			map[string]interface{}{"test": []interface{}{"a", "b", "c"}},
		},
		{
			`{"test": 12, "test2": true}`,
			&openapi.SchemaRef{
				Value: &openapi.Schema{
					AnyOf: []*openapi.SchemaRef{
						{
							Value: &openapi.Schema{
								Type: "object",
								Properties: &openapi.Schemas{
									Value: map[string]*openapi.SchemaRef{
										"test": {
											Value: &openapi.Schema{
												Type: "integer",
											},
										},
									},
								},
							},
						},
						{
							Value: &openapi.Schema{
								Type: "object",
								Properties: &openapi.Schemas{
									Value: map[string]*openapi.SchemaRef{
										"test2": {
											Value: &openapi.Schema{
												Type: "boolean",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			map[string]interface{}{"test": int64(12), "test2": true},
		},
		{
			`"hello world"`,
			&openapi.SchemaRef{
				Value: &openapi.Schema{
					AnyOf: []*openapi.SchemaRef{
						{
							Value: &openapi.Schema{
								Type: "object",
								Properties: &openapi.Schemas{
									Value: map[string]*openapi.SchemaRef{
										"test": {
											Value: &openapi.Schema{
												Type: "integer",
											},
										},
									},
								},
							},
						},
						{
							Value: &openapi.Schema{
								Type: "string",
							},
						},
					},
				},
			},
			"hello world",
		},
	}

	for i, d := range data {
		t.Logf("parse %v: %v", i, d.s)
		i, err := Parse([]byte(d.s), media.ParseContentType("application/json"), d.schema)
		test.Ok(t, err)
		test.Equals(t, d.e, i)
	}
}

func TestParseOneOf(t *testing.T) {
	data := []struct {
		s      string
		schema *openapi.SchemaRef
		e      interface{}
		err    error
	}{
		{
			`{"test2": true}`,
			&openapi.SchemaRef{
				Value: &openapi.Schema{
					OneOf: []*openapi.SchemaRef{
						{
							Value: &openapi.Schema{
								Type: "object",
								Properties: &openapi.Schemas{
									Value: map[string]*openapi.SchemaRef{
										"test": {
											Value: &openapi.Schema{
												Type: "integer",
											},
										},
									},
								},
							},
						},
						{
							Value: &openapi.Schema{
								Type: "object",
								Properties: &openapi.Schemas{
									Value: map[string]*openapi.SchemaRef{
										"test2": {
											Value: &openapi.Schema{
												Type: "boolean",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			map[string]interface{}{"test2": true},
			nil,
		},
		{
			`{"test": 12, "test2": true}`,
			&openapi.SchemaRef{
				Value: &openapi.Schema{
					OneOf: []*openapi.SchemaRef{
						{
							Value: &openapi.Schema{
								Type: "object",
								Properties: &openapi.Schemas{
									Value: map[string]*openapi.SchemaRef{
										"test": {
											Value: &openapi.Schema{
												Type: "integer",
											},
										},
									},
								},
							},
						},
						{
							Value: &openapi.Schema{
								Type: "object",
								Properties: &openapi.Schemas{
									Value: map[string]*openapi.SchemaRef{
										"test2": {
											Value: &openapi.Schema{
												Type: "boolean",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			nil,
			fmt.Errorf("oneOf: given data is valid against more as one schema"),
		},
		{
			`"hello world"`,
			&openapi.SchemaRef{
				Value: &openapi.Schema{
					OneOf: []*openapi.SchemaRef{
						{
							Value: &openapi.Schema{
								Type: "integer",
							},
						},
						{
							Value: &openapi.Schema{
								Type: "string",
							},
						},
					},
				},
			},
			"hello world",
			nil,
		},
	}

	for i, d := range data {
		t.Logf("parse %v: %v", i, d.s)
		i, err := Parse([]byte(d.s), media.ParseContentType("application/json"), d.schema)
		test.Equals(t, d.err, err)
		test.Equals(t, d.e, i)
	}
}

func TestParseAllOf(t *testing.T) {
	data := []struct {
		s      string
		schema *openapi.SchemaRef
		e      interface{}
	}{
		{
			`{"test": 12, "test2": true}`,
			&openapi.SchemaRef{
				Value: &openapi.Schema{
					AllOf: []*openapi.SchemaRef{
						{
							Value: &openapi.Schema{
								Type: "object",
								Properties: &openapi.Schemas{
									Value: map[string]*openapi.SchemaRef{
										"test": {
											Value: &openapi.Schema{
												Type: "integer",
											},
										},
									},
								},
							},
						},
						{
							Value: &openapi.Schema{
								Type: "object",
								Properties: &openapi.Schemas{
									Value: map[string]*openapi.SchemaRef{
										"test2": {
											Value: &openapi.Schema{
												Type: "boolean",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			map[string]interface{}{"test": int64(12), "test2": true},
		},
	}

	for i, d := range data {
		t.Logf("parse %v: %v", i, d.s)
		i, err := Parse([]byte(d.s), media.ParseContentType("application/json"), d.schema)
		test.Ok(t, err)
		test.Equals(t, d.e, i)
	}
}

package web

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/config/dynamic/openapi"
	"mokapi/test"
	"net/http"
	"net/url"
	"testing"
)

func TestParseParam(t *testing.T) {
	type assert func(RequestParameters) error
	data := []struct {
		request *http.Request
		route   string
		params  openapi.Parameters
		assert  assert
	}{
		{
			&http.Request{URL: &url.URL{Path: "/api/v1/channels/test/messages/123456743/test"}},
			"/api/v1/channels/{channel}/messages/{id}/test",
			openapi.Parameters{
				&openapi.ParameterRef{Value: &openapi.Parameter{
					Name: "channel",
					Type: openapi.PathParameter,
					Schema: &openapi.SchemaRef{
						Value: &openapi.Schema{Type: "string"},
					},
					Required: true,
				}},
				&openapi.ParameterRef{Value: &openapi.Parameter{
					Name: "id",
					Type: openapi.PathParameter,
					Schema: &openapi.SchemaRef{
						Value: &openapi.Schema{Type: "integer", Format: "int64"},
					},
					Required: true,
				}},
			},
			func(p RequestParameters) error {
				if channel, ok := p[openapi.PathParameter]["channel"]; !ok {
					return errors.New("expected channel parameter in path")
				} else if channel.Value != "test" {
					return errors.Errorf("got %v; expected id value test", channel)
				}
				if id, ok := p[openapi.PathParameter]["id"]; !ok {
					return errors.New("GET: expected id parameter in path")
				} else if id.Value != int64(123456743) {
					return errors.Errorf("got %v; expected id value 123456743", id)
				}
				return nil
			},
		},
		{
			&http.Request{URL: &url.URL{Path: "/api/v1/search", RawQuery: "limit=10"}},
			"/api/v1/search",
			openapi.Parameters{
				&openapi.ParameterRef{
					Value: &openapi.Parameter{
						Name: "limit",
						Type: openapi.QueryParameter,
						Schema: &openapi.SchemaRef{
							Value: &openapi.Schema{Type: "integer"},
						},
					},
				},
			},
			func(p RequestParameters) error {
				if limit, ok := p[openapi.QueryParameter]["limit"]; !ok {
					return errors.New("expected limit parameter in query")
				} else if limit.Value != 10 {
					t.Errorf("got %v; expected limit value 10", limit)
				}
				return nil
			},
		},
	}

	for _, d := range data {
		p, err := parseParams(d.params, d.route, d.request)
		if err != nil {
			t.Errorf("%v: %v", d.request.URL, err)
		}
		err = d.assert(p)
		if err != nil {
			t.Errorf("%v: %v", d.request.URL, err)
		}
	}
}

func TestParse(t *testing.T) {
	testdata := []struct {
		s      string
		schema *openapi.SchemaRef
		e      interface{}
	}{
		{
			"foobar",
			&openapi.SchemaRef{Value: &openapi.Schema{Type: "string"}},
			"foobar",
		},
		{
			"123",
			&openapi.SchemaRef{Value: &openapi.Schema{Type: "integer"}},
			123,
		},
		{
			"123123123123",
			&openapi.SchemaRef{Value: &openapi.Schema{Type: "integer", Format: "int64"}},
			int64(123123123123),
		},
		{
			"123.123",
			&openapi.SchemaRef{Value: &openapi.Schema{Type: "number"}},
			float32(123.123),
		},
		{
			"123.123",
			&openapi.SchemaRef{Value: &openapi.Schema{Type: "number", Format: "double"}},
			123.123,
		},
		{
			"true",
			&openapi.SchemaRef{Value: &openapi.Schema{Type: "boolean"}},
			true,
		},
		{
			"false",
			&openapi.SchemaRef{Value: &openapi.Schema{Type: "boolean"}},
			false,
		},
		{
			"12",
			&openapi.SchemaRef{
				Value: &openapi.Schema{
					AnyOf: []*openapi.SchemaRef{
						{Value: &openapi.Schema{Type: "integer"}},
						{Value: &openapi.Schema{Type: "string"}},
					}}},
			12,
		},
	}

	t.Parallel()
	for i, data := range testdata {
		d := data
		t.Run(fmt.Sprintf("parse %v: %v", i, d.s), func(t *testing.T) {
			t.Parallel()
			i, err := parse(d.s, d.schema)
			test.Ok(t, err)
			test.Equals(t, d.e, i)
		})
	}
}

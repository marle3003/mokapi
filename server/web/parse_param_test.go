package web

import (
	"github.com/pkg/errors"
	"mokapi/models"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestParseParam(t *testing.T) {
	type assert func(RequestParameters) error
	data := []struct {
		request *http.Request
		route   string
		params  []*models.Parameter
		assert  assert
	}{
		{
			&http.Request{URL: &url.URL{Path: "/api/v1/channels/test/messages/123456743/test"}},
			"/api/v1/channels/{channel}/messages/{id}/test",
			[]*models.Parameter{
				{
					Name:     "channel",
					Location: models.PathParameter,
					Schema:   &models.Schema{Type: "string"},
					Required: true,
				},
				{
					Name:     "id",
					Location: models.PathParameter,
					Schema:   &models.Schema{Type: "integer", Format: "int64"},
					Required: true,
				},
			},
			func(p RequestParameters) error {
				if channel, ok := p[models.PathParameter]["channel"]; !ok {
					return errors.New("expected channel parameter in path")
				} else if channel != "test" {
					return errors.Errorf("got %v; expected id value test", channel)
				}
				if id, ok := p[models.PathParameter]["id"]; !ok {
					return errors.New("GET: expected id parameter in path")
				} else if id != int64(123456743) {
					return errors.Errorf("got %v; expected id value 123456743", id)
				}
				return nil
			},
		},
		{
			&http.Request{URL: &url.URL{Path: "/api/v1/search", RawQuery: "limit=10"}},
			"/api/v1/search",
			[]*models.Parameter{
				{
					Name:     "limit",
					Location: models.QueryParameter,
					Schema:   &models.Schema{Type: "integer"},
				},
			},
			func(p RequestParameters) error {
				if limit, ok := p[models.QueryParameter]["limit"]; !ok {
					return errors.New("expected limit parameter in query")
				} else if limit != 10 {
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
	data := []struct {
		s      string
		schema *models.Schema
		e      interface{}
	}{
		{
			"foobar",
			&models.Schema{Type: "string"},
			"foobar",
		},
		{
			"123",
			&models.Schema{Type: "integer"},
			123,
		},
		{
			"123123123123",
			&models.Schema{Type: "integer", Format: "int64"},
			int64(123123123123),
		},
		{
			"123.123",
			&models.Schema{Type: "number"},
			float32(123.123),
		},
		{
			"123.123",
			&models.Schema{Type: "number", Format: "double"},
			123.123,
		},
		{
			"true",
			&models.Schema{Type: "boolean"},
			true,
		},
		{
			"false",
			&models.Schema{Type: "boolean"},
			false,
		},
	}

	for _, d := range data {
		i, err := parse(d.s, d.schema)
		if err != nil {
			t.Errorf("parse %v: %v", d.schema.Type, err)
		} else if !reflect.DeepEqual(i, d.e) {
			t.Errorf("schema %q, format %q: got %v; expected %v", d.schema.Type, d.schema.Format, i, d.e)
		}
	}
}

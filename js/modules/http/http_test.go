package http

import (
	"github.com/dop251/goja"
	"mokapi/config/dynamic/openapi"
	"mokapi/js/common"
	"mokapi/test"
	"testing"
)

func TestOn(t *testing.T) {
	t.Parallel()

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		rt := goja.New()
		rt.SetFieldNameMapper(goja.TagFieldNameMapper("js", true))
		http := New()
		err := rt.Set("http", common.Map(rt, http))
		test.Ok(t, err)

		_, err = rt.RunString("http.on({}, function() {return true;});")
		test.Ok(t, err)
		b := http.Listeners[0].Listener(nil, nil)
		test.Equals(t, true, b)
	})
	t.Run("method parameter", func(t *testing.T) {
		t.Parallel()
		rt := goja.New()
		rt.SetFieldNameMapper(goja.TagFieldNameMapper("js", true))
		http := New()
		err := rt.Set("http", common.Map(rt, http))
		test.Ok(t, err)

		_, err = rt.RunString("http.on({method: 'GET'}, function() {});")
		test.Ok(t, err)
		test.Equals(t, "GET", http.Listeners[0].Event.Method)
	})
	t.Run("url parameter", func(t *testing.T) {
		t.Parallel()
		rt := goja.New()
		rt.SetFieldNameMapper(goja.TagFieldNameMapper("js", true))
		http := New()
		err := rt.Set("http", common.Map(rt, http))
		test.Ok(t, err)

		_, err = rt.RunString("http.on({url: '/abc'}, function() {});")
		test.Ok(t, err)
		test.Equals(t, "/abc", http.Listeners[0].Event.Url)
	})

	t.Run("request method", func(t *testing.T) {
		t.Parallel()
		rt := goja.New()
		rt.SetFieldNameMapper(goja.TagFieldNameMapper("js", true))
		http := New()
		err := rt.Set("http", common.Map(rt, http))
		test.Ok(t, err)

		_, err = rt.RunString("http.on({}, function(request) {return request.method === 'GET'});")
		test.Ok(t, err)

		b := http.Listeners[0].Listener(&openapi.EventRequest{Method: "GET"}, nil)
		test.Equals(t, true, b)
	})

	t.Run("request header", func(t *testing.T) {
		t.Parallel()
		rt := goja.New()
		rt.SetFieldNameMapper(goja.TagFieldNameMapper("js", true))
		http := New()
		err := rt.Set("http", common.Map(rt, http))
		test.Ok(t, err)

		_, err = rt.RunString("http.on({}, function(request) {return request.header.Accept === 'application/json' && request.header['Accept'] === 'application/json'});")
		test.Ok(t, err)

		b := http.Listeners[0].Listener(&openapi.EventRequest{Header: map[string]interface{}{"Accept": "application/json"}}, nil)
		test.Equals(t, true, b)
	})
}

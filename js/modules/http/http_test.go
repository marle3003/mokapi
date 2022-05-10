package http

import (
	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi"
	"mokapi/js/common"
	"testing"
)

func TestOn(t *testing.T) {
	t.Parallel()

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		rt := goja.New()
		rt.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
		http := New()
		err := rt.Set("http", common.Map(rt, http))
		require.NoError(t, err)

		_, err = rt.RunString("http.on({}, function() {return true;});")
		require.NoError(t, err)
		b := http.Listeners[0].Listener(nil, nil)
		require.True(t, b)
	})
	t.Run("method parameter", func(t *testing.T) {
		t.Parallel()
		rt := goja.New()
		rt.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
		http := New()
		err := rt.Set("http", common.Map(rt, http))
		require.NoError(t, err)

		_, err = rt.RunString("http.on({method: 'GET'}, function() {});")
		require.NoError(t, err)
		require.Equal(t, "GET", http.Listeners[0].Event.Method)
	})
	t.Run("url parameter", func(t *testing.T) {
		t.Parallel()
		rt := goja.New()
		rt.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
		http := New()
		err := rt.Set("http", common.Map(rt, http))
		require.NoError(t, err)

		_, err = rt.RunString("http.on({url: '/abc'}, function() {});")
		require.NoError(t, err)
		require.Equal(t, "/abc", http.Listeners[0].Event.Url)
	})

	t.Run("request method", func(t *testing.T) {
		t.Parallel()
		rt := goja.New()
		rt.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
		http := New()
		err := rt.Set("http", common.Map(rt, http))
		require.NoError(t, err)

		_, err = rt.RunString("http.on({}, function(request) {return request.method === 'GET'});")
		require.NoError(t, err)

		b := http.Listeners[0].Listener(&openapi.EventRequest{Method: "GET"}, nil)
		require.True(t, b)
	})

	t.Run("request header", func(t *testing.T) {
		t.Parallel()
		rt := goja.New()
		rt.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
		http := New()
		err := rt.Set("http", common.Map(rt, http))
		require.NoError(t, err)

		_, err = rt.RunString("http.on({}, function(request) {return request.header.Accept === 'application/json' && request.header['Accept'] === 'application/json'});")
		require.NoError(t, err)

		b := http.Listeners[0].Listener(&openapi.EventRequest{Header: map[string]interface{}{"Accept": "application/json"}}, nil)
		require.True(t, b)
	})
}

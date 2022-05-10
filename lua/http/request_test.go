package http

import (
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
	"net/http"
	"testing"
)

type testClient struct {
	req    *http.Request
	doFunc func(request *http.Request) (*http.Response, error)
}

func (c *testClient) Do(request *http.Request) (*http.Response, error) {
	c.req = request
	if c.doFunc != nil {
		return c.doFunc(request)
	}
	return &http.Response{}, nil
}

func TestGet(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		client := &testClient{}
		state := lua.NewState()
		mod := New()
		mod.client = client
		state.PreloadModule("http", mod.Loader)
		err := state.DoString(`
			http = require("http")
			http.get("http://localhost/foo")`,
		)
		require.NoError(t, err)
		require.Equal(t, "GET", client.req.Method)
		require.Equal(t, "http://localhost/foo", client.req.URL.String())
	})
	t.Run("header", func(t *testing.T) {
		client := &testClient{}
		state := lua.NewState()
		mod := New()
		mod.client = client
		state.PreloadModule("http", mod.Loader)
		err := state.DoString(`
			http = require("http")
			http.get("http://localhost/foo", {headers = {foo = "bar"}})`,
		)
		require.NoError(t, err)
		require.Equal(t, "http://localhost/foo", client.req.URL.String())
		require.Equal(t, "bar", client.req.Header.Get("foo"))
	})
	t.Run("headerWithArray", func(t *testing.T) {
		client := &testClient{}
		state := lua.NewState()
		mod := New()
		mod.client = client
		state.PreloadModule("http", mod.Loader)
		err := state.DoString(`
			http = require("http")
			http.get("http://localhost/foo", {headers = {foo = {"hello", "world"}}})`,
		)
		require.NoError(t, err)
		require.Equal(t, "http://localhost/foo", client.req.URL.String())
		require.Equal(t, []string{"hello", "world"}, client.req.Header.Values("foo"))
	})
}

func TestPost(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		client := &testClient{}
		state := lua.NewState()
		mod := New()
		mod.client = client
		state.PreloadModule("http", mod.Loader)
		err := state.DoString(`
			http = require("http")
			http.post("http://localhost/foo")`,
		)
		require.NoError(t, err)
		require.Equal(t, "POST", client.req.Method)
		require.Equal(t, "http://localhost/foo", client.req.URL.String())
	})
	t.Run("contenttype", func(t *testing.T) {
		client := &testClient{}
		state := lua.NewState()
		mod := New()
		mod.client = client
		state.PreloadModule("http", mod.Loader)
		err := state.DoString(`
			http = require("http")
			http.post("http://localhost/foo", "body", {headers = {['Content-Type'] = "application/json"}})`,
		)
		require.NoError(t, err)
		require.Equal(t, "POST", client.req.Method)
		require.Equal(t, "application/json", client.req.Header.Get("Content-Type"))
		require.Equal(t, "http://localhost/foo", client.req.URL.String())
	})
}

package http

import (
	lua "github.com/yuin/gopher-lua"
	"mokapi/test"
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
		state.PreloadModule("http", New(client).Loader)
		err := state.DoString(`
			http = require("http")
			http.get("http://localhost/foo")`,
		)
		test.Ok(t, err)
		test.Equals(t, "http://localhost/foo", client.req.URL.String())
	})
	t.Run("header", func(t *testing.T) {
		client := &testClient{}
		state := lua.NewState()
		state.PreloadModule("http", New(client).Loader)
		err := state.DoString(`
			http = require("http")
			http.get("http://localhost/foo", {headers = {foo = "bar"}})`,
		)
		test.Ok(t, err)
		test.Equals(t, "http://localhost/foo", client.req.URL.String())
		test.Equals(t, "bar", client.req.Header.Get("foo"))
	})
	t.Run("headerWithArray", func(t *testing.T) {
		client := &testClient{}
		state := lua.NewState()
		state.PreloadModule("http", New(client).Loader)
		err := state.DoString(`
			http = require("http")
			http.get("http://localhost/foo", {headers = {foo = {"hello", "world"}}})`,
		)
		test.Ok(t, err)
		test.Equals(t, "http://localhost/foo", client.req.URL.String())
		test.Equals(t, []string{"hello", "world"}, client.req.Header.Values("foo"))
	})
}

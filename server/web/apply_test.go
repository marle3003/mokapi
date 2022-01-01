package web_test

import (
	logtest "github.com/sirupsen/logrus/hooks/test"
	"mokapi/config/dynamic/openapi"
	"mokapi/server/web"
	"mokapi/test"
	"testing"
)

func TestApply(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, b *web.Binding, hook *logtest.Hook)
	}{
		{"nil config",
			func(t *testing.T, b *web.Binding, _ *logtest.Hook) {
				err := b.Apply(nil)
				test.EqualError(t, "unexpected parameter type <nil> in http binding", err)
			}},
		{"no server specified",
			func(t *testing.T, b *web.Binding, hook *logtest.Hook) {
				c := &openapi.Config{Info: openapi.Info{Name: "foo"}}
				err := b.Apply(c)
				test.Ok(t, err)
				test.Equals(t, nil, hook.LastEntry())
			}},
		{"no server specified",
			func(t *testing.T, b *web.Binding, hook *logtest.Hook) {
				c := &openapi.Config{Info: openapi.Info{Name: "foo"}}
				err := b.Apply(c)
				test.Ok(t, err)
				test.Equals(t, nil, hook.LastEntry())
			}},
		{"different port",
			func(t *testing.T, b *web.Binding, hook *logtest.Hook) {
				c := &openapi.Config{Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://:5000"}}}
				err := b.Apply(c)
				test.Ok(t, err)
				test.Equals(t, nil, hook.LastEntry())
			}},
		{"different port https://localhost",
			func(t *testing.T, b *web.Binding, hook *logtest.Hook) {
				c := &openapi.Config{Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "https://localhost"}}}
				err := b.Apply(c)
				test.Ok(t, err)
				test.Equals(t, nil, hook.LastEntry())
			}},
		{"add new host http://:80",
			func(t *testing.T, b *web.Binding, hook *logtest.Hook) {
				c := &openapi.Config{Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://:80"}}}
				err := b.Apply(c)
				test.Ok(t, err)
				test.Equals(t, "Adding service foo on binding :80 on path /", hook.LastEntry().Message)
			}},
		{"add new host http://localhost:80",
			func(t *testing.T, b *web.Binding, hook *logtest.Hook) {
				c := &openapi.Config{Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://localhost:80"}}}
				err := b.Apply(c)
				test.Ok(t, err)
				test.Equals(t, "Adding service foo on binding :80 on path /", hook.LastEntry().Message)
			}},
		{"add new host http://localhost",
			func(t *testing.T, b *web.Binding, hook *logtest.Hook) {
				c := &openapi.Config{Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://localhost"}}}
				err := b.Apply(c)
				test.Ok(t, err)
				test.Equals(t, "Adding service foo on binding :80 on path /", hook.LastEntry().Message)
			}},
		{"invalid port format",
			func(t *testing.T, b *web.Binding, hook *logtest.Hook) {
				c := &openapi.Config{Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://localhost:foo"}}}
				err := b.Apply(c)
				test.Ok(t, err)
				test.Equals(t, "API foo: parse \"http://localhost:foo\": invalid port \":foo\" after host", hook.LastEntry().Message)
			}},
		{"invalid url format",
			func(t *testing.T, b *web.Binding, hook *logtest.Hook) {
				c := &openapi.Config{Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "$://"}}}
				err := b.Apply(c)
				test.Ok(t, err)
				test.Equals(t, "API foo: parse \"$://\": first path segment in URL cannot contain colon", hook.LastEntry().Message)
			}},
		{"empty url",
			func(t *testing.T, b *web.Binding, hook *logtest.Hook) {
				c := &openapi.Config{Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: ""}}}
				err := b.Apply(c)
				test.Ok(t, err)
				test.Equals(t, nil, hook.LastEntry())
			}},
		{"add on same path",
			func(t *testing.T, b *web.Binding, hook *logtest.Hook) {
				c := &openapi.Config{Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "/foo"}}}
				err := b.Apply(c)
				test.Ok(t, err)
				c = &openapi.Config{Info: openapi.Info{Name: "bar"}, Servers: []*openapi.Server{{Url: "/foo"}}}
				err = b.Apply(c)
				test.Error(t, err)
			}},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			hook := test.NewNullLogger()

			b := web.NewBinding(":80")

			data.fn(t, b, hook)
		})

	}
}

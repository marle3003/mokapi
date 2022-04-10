package common_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/script"
	"net/url"
	"os"
	"testing"
)

func TestConfig_Parse(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "default",
			f: func(t *testing.T) {
				c := common.NewConfig(mustUrl("foo.txt"))
				c.Raw = []byte("name: foobar")

				err := c.Parse(&testReader{})
				require.NoError(t, err)
				require.Equal(t, "name: foobar", c.Data)
			},
		},
		{
			name: "opaque",
			f: func(t *testing.T) {
				u := mustUrl("foo.yml")
				u.Opaque = "foo.txt"
				c := common.NewConfig(u)
				c.Raw = []byte("name: foobar")

				err := c.Parse(&testReader{})
				require.NoError(t, err)
				require.Equal(t, "name: foobar", c.Data)
			},
		},
		{
			name: "yaml",
			f: func(t *testing.T) {
				data := &struct {
					Name string
				}{}
				c := common.NewConfig(mustUrl("foo.yml"), common.WithData(data))
				c.Raw = []byte("name: foobar")

				err := c.Parse(&testReader{})
				require.NoError(t, err)
				require.Equal(t, "foobar", data.Name)
			},
		},
		{
			name: "yaml error",
			f: func(t *testing.T) {
				data := &struct {
					Flag bool
				}{}
				c := common.NewConfig(mustUrl("foo.yml"), common.WithData(data))
				c.Raw = []byte("flag: foobar")

				err := c.Parse(&testReader{})
				require.NoError(t, err)
				require.Equal(t, "flag: foobar", c.Data)
			},
		},
		{
			name: "openapi",
			f: func(t *testing.T) {
				c := common.NewConfig(mustUrl("foo.yml"))
				c.Raw = []byte("openapi: 3.0")

				err := c.Parse(&testReader{})
				require.NoError(t, err)
				require.IsType(t, &openapi.Config{}, c.Data)
				o := c.Data.(*openapi.Config)
				require.Equal(t, "3.0", o.OpenApi)
			},
		},
		{
			name: "json",
			f: func(t *testing.T) {
				data := &struct {
					Name string
				}{}
				c := common.NewConfig(mustUrl("foo.json"), common.WithData(data))
				c.Raw = []byte("{\"name\":\"foobar\"}")

				err := c.Parse(&testReader{})
				require.NoError(t, err)
				require.Equal(t, "foobar", data.Name)
			},
		},
		{
			name: "json error",
			f: func(t *testing.T) {
				data := &struct {
					Name string
				}{}
				c := common.NewConfig(mustUrl("foo.json"), common.WithData(data))
				c.Raw = []byte("{\"name\"=\"foobar\"}")

				err := c.Parse(&testReader{})
				require.NoError(t, err)
				require.Equal(t, "{\"name\"=\"foobar\"}", c.Data)
			},
		},
		{
			name: "openapi json",
			f: func(t *testing.T) {
				c := common.NewConfig(mustUrl("foo.json"))
				c.Raw = []byte("{\"openapi\": \"3.0\"}")

				err := c.Parse(&testReader{})
				require.NoError(t, err)
				require.IsType(t, &openapi.Config{}, c.Data)
				o := c.Data.(*openapi.Config)
				require.Equal(t, "3.0", o.OpenApi)
			},
		},
		{
			name: "lua",
			f: func(t *testing.T) {
				c := common.NewConfig(mustUrl("foo.lua"))
				c.Raw = []byte("print(\"Hello World\")")

				err := c.Parse(&testReader{})
				require.NoError(t, err)
				require.IsType(t, &script.Script{}, c.Data)
				s := c.Data.(*script.Script)
				require.Equal(t, "foo.lua", s.Filename)
				require.Equal(t, "print(\"Hello World\")", s.Code)
			},
		},
		{
			name: "lua update",
			f: func(t *testing.T) {
				s := script.New("foo.lua", []byte(""))
				c := common.NewConfig(mustUrl("foo.lua"), common.WithData(s))
				c.Raw = []byte("print(\"Hello World\")")

				err := c.Parse(&testReader{})
				require.NoError(t, err)
				require.IsType(t, &script.Script{}, c.Data)
				s1 := c.Data.(*script.Script)
				require.True(t, s == s1)
				require.Equal(t, "foo.lua", s1.Filename)
				require.Equal(t, "print(\"Hello World\")", s1.Code)
			},
		},
		{
			name: "template",
			f: func(t *testing.T) {
				c := common.NewConfig(mustUrl("foo.tmpl"))
				c.Raw = []byte("the user is {{ env \"TEST_USER\" }}")
				os.Setenv("TEST_USER", "foo")
				defer os.Unsetenv("TEST_USER")

				err := c.Parse(&testReader{})
				require.NoError(t, err)
				require.Equal(t, "the user is foo", c.Data)
			},
		},
		{
			name: "template extract username",
			f: func(t *testing.T) {
				c := common.NewConfig(mustUrl("foo.tmpl"))
				c.Raw = []byte("the user is {{ env \"TEST_USER\" | extractUsername }}")
				os.Setenv("TEST_USER", "foo\\bar")
				defer os.Unsetenv("TEST_USER")

				err := c.Parse(&testReader{})
				require.NoError(t, err)
				require.Equal(t, "the user is bar", c.Data)
			},
		},
		{
			name: "template err",
			f: func(t *testing.T) {
				c := common.NewConfig(mustUrl("foo.tmpl"))
				c.Raw = []byte("the user is {{ env \"TEST_USER\" | foo }}")
				os.Setenv("TEST_USER", "foo\\bar")
				defer os.Unsetenv("TEST_USER")

				err := c.Parse(&testReader{})
				require.EqualError(t, err, "unable to render template foo.tmpl: template: :1: function \"foo\" not defined")
				require.Nil(t, c.Data)
			},
		},
		{
			name: "parse err",
			f: func(t *testing.T) {
				c := common.NewConfig(mustUrl("foo.tmpl"))
				c.Raw = []byte("the user is {{ env \"TEST_USER\" | foo }}")
				os.Setenv("TEST_USER", "foo\\bar")
				defer os.Unsetenv("TEST_USER")

				err := c.Parse(&testReader{})
				require.EqualError(t, err, "unable to render template foo.tmpl: template: :1: function \"foo\" not defined")
				require.Nil(t, c.Data)
			},
		},
		{
			name: "parser test",
			f: func(t *testing.T) {
				data := &testParser{parse: func(file *common.Config, reader common.Reader) error {
					return fmt.Errorf("TEST ERROR")
				}}
				c := common.NewConfig(mustUrl("foo.yml"), common.WithData(data))
				c.Raw = []byte("name: foobar")

				err := c.Parse(&testReader{})
				require.EqualError(t, err, "parsing file foo.yml: TEST ERROR")
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.f(t)
		})
	}
}

type testReader struct {
}

func (r *testReader) Read(u *url.URL, opts ...common.ConfigOptions) (*common.Config, error) {
	return &common.Config{}, nil
}

type testParser struct {
	Name  string
	parse func(file *common.Config, reader common.Reader) error
}

func (p *testParser) Parse(file *common.Config, reader common.Reader) error {
	if p.parse != nil {
		return p.parse(file, reader)
	}
	return nil
}

func mustUrl(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

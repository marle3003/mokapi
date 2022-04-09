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

func TestFile_Parse(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "default",
			f: func(t *testing.T) {
				f := common.NewFile(mustUrl("foo.txt"))

				c := &common.Config{Url: f.Url, Data: []byte("name: foobar")}
				err := f.Parse(c, &testReader{})
				require.NoError(t, err)
				require.Equal(t, "name: foobar", f.Data)
			},
		},
		{
			name: "opaque",
			f: func(t *testing.T) {
				u := mustUrl("foo.yml")
				u.Opaque = "foo.txt"
				f := common.NewFile(u)

				c := &common.Config{Url: f.Url, Data: []byte("name: foobar")}
				err := f.Parse(c, &testReader{})
				require.NoError(t, err)
				require.Equal(t, "name: foobar", f.Data)
			},
		},
		{
			name: "yaml",
			f: func(t *testing.T) {
				data := &struct {
					Name string
				}{}
				f := common.NewFile(mustUrl("foo.yml"), common.WithData(data))

				c := &common.Config{Url: f.Url, Data: []byte("name: foobar")}
				err := f.Parse(c, &testReader{})
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
				f := common.NewFile(mustUrl("foo.yml"), common.WithData(data))

				c := &common.Config{Url: f.Url, Data: []byte("flag: foobar")}
				err := f.Parse(c, &testReader{})
				require.NoError(t, err)
				require.Equal(t, "flag: foobar", f.Data)
			},
		},
		{
			name: "openapi",
			f: func(t *testing.T) {
				f := common.NewFile(mustUrl("foo.yml"))

				c := &common.Config{Url: f.Url, Data: []byte("openapi: 3.0")}
				err := f.Parse(c, &testReader{})
				require.NoError(t, err)
				require.IsType(t, &openapi.Config{}, f.Data)
				o := f.Data.(*openapi.Config)
				require.Equal(t, "3.0", o.OpenApi)
			},
		},
		{
			name: "json",
			f: func(t *testing.T) {
				data := &struct {
					Name string
				}{}
				f := common.NewFile(mustUrl("foo.json"), common.WithData(data))

				c := &common.Config{Url: f.Url, Data: []byte("{\"name\":\"foobar\"}")}
				err := f.Parse(c, &testReader{})
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
				f := common.NewFile(mustUrl("foo.json"), common.WithData(data))

				c := &common.Config{Url: f.Url, Data: []byte("{\"name\"=\"foobar\"}")}
				err := f.Parse(c, &testReader{})
				require.NoError(t, err)
				require.Equal(t, "{\"name\"=\"foobar\"}", f.Data)
			},
		},
		{
			name: "openapi json",
			f: func(t *testing.T) {
				f := common.NewFile(mustUrl("foo.json"))

				c := &common.Config{Url: f.Url, Data: []byte("{\"openapi\": \"3.0\"}")}
				err := f.Parse(c, &testReader{})
				require.NoError(t, err)
				require.IsType(t, &openapi.Config{}, f.Data)
				o := f.Data.(*openapi.Config)
				require.Equal(t, "3.0", o.OpenApi)
			},
		},
		{
			name: "lua",
			f: func(t *testing.T) {
				f := common.NewFile(mustUrl("foo.lua"))

				c := &common.Config{Url: f.Url, Data: []byte("print(\"Hello World\")")}
				err := f.Parse(c, &testReader{})
				require.NoError(t, err)
				require.IsType(t, &script.Script{}, f.Data)
				s := f.Data.(*script.Script)
				require.Equal(t, "foo.lua", s.Filename)
				require.Equal(t, "print(\"Hello World\")", s.Code)
			},
		},
		{
			name: "lua update",
			f: func(t *testing.T) {
				s := script.New("foo.lua", []byte(""))
				f := common.NewFile(mustUrl("foo.lua"), common.WithData(s))

				c := &common.Config{Url: f.Url, Data: []byte("print(\"Hello World\")")}
				err := f.Parse(c, &testReader{})
				require.NoError(t, err)
				require.IsType(t, &script.Script{}, f.Data)
				s1 := f.Data.(*script.Script)
				require.True(t, s == s1)
				require.Equal(t, "foo.lua", s1.Filename)
				require.Equal(t, "print(\"Hello World\")", s1.Code)
			},
		},
		{
			name: "template",
			f: func(t *testing.T) {
				f := common.NewFile(mustUrl("foo.tmpl"))
				os.Setenv("TEST_USER", "foo")
				defer os.Unsetenv("TEST_USER")
				c := &common.Config{Url: f.Url, Data: []byte("the user is {{ env \"TEST_USER\" }}")}
				err := f.Parse(c, &testReader{})
				require.NoError(t, err)
				require.Equal(t, "the user is foo", f.Data)
			},
		},
		{
			name: "template extract username",
			f: func(t *testing.T) {
				f := common.NewFile(mustUrl("foo.tmpl"))
				os.Setenv("TEST_USER", "foo\\bar")
				defer os.Unsetenv("TEST_USER")
				c := &common.Config{Url: f.Url, Data: []byte("the user is {{ env \"TEST_USER\" | extractUsername }}")}
				err := f.Parse(c, &testReader{})
				require.NoError(t, err)
				require.Equal(t, "the user is bar", f.Data)
			},
		},
		{
			name: "template err",
			f: func(t *testing.T) {
				f := common.NewFile(mustUrl("foo.tmpl"))
				os.Setenv("TEST_USER", "foo\\bar")
				defer os.Unsetenv("TEST_USER")
				c := &common.Config{Url: f.Url, Data: []byte("the user is {{ env \"TEST_USER\" | foo }}")}
				err := f.Parse(c, &testReader{})
				require.EqualError(t, err, "unable to render template foo.tmpl: template: :1: function \"foo\" not defined")
				require.Nil(t, f.Data)
			},
		},
		{
			name: "parse err",
			f: func(t *testing.T) {
				f := common.NewFile(mustUrl("foo.tmpl"))
				os.Setenv("TEST_USER", "foo\\bar")
				defer os.Unsetenv("TEST_USER")
				c := &common.Config{Url: f.Url, Data: []byte("the user is {{ env \"TEST_USER\" | foo }}")}
				err := f.Parse(c, &testReader{})
				require.EqualError(t, err, "unable to render template foo.tmpl: template: :1: function \"foo\" not defined")
				require.Nil(t, f.Data)
			},
		},
		{
			name: "parser test",
			f: func(t *testing.T) {
				data := &testParser{parse: func(file *common.File, reader common.Reader) error {
					return fmt.Errorf("TEST ERROR")
				}}
				f := common.NewFile(mustUrl("foo.yml"), common.WithData(data))

				c := &common.Config{Url: f.Url, Data: []byte("name: foobar")}
				err := f.Parse(c, &testReader{})
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

func (r *testReader) Read(u *url.URL, opts ...common.FileOptions) (*common.File, error) {
	return &common.File{}, nil
}

type testParser struct {
	Name  string
	parse func(file *common.File, reader common.Reader) error
}

func (p *testParser) Parse(file *common.File, reader common.Reader) error {
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

package dynamic_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/dynamic/script"
	"net/url"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	mustUrl := func(s string) *url.URL {
		u, err := url.Parse(s)
		if err != nil {
			panic(err)
		}
		return u
	}

	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "text",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.txt")},
					Raw:  []byte(`Hello World`),
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "Hello World", c.Data)
			},
		},
		{
			name: "unknown json",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.json")},
					Raw:  []byte(`{"name": "foo"}`),
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Nil(t, c.Data)
			},
		},
		{
			name: "json error",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.json")},
					Raw:  []byte(`{"name": "foo"`),
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "unexpected end of JSON input")
				require.Nil(t, c.Data)
			},
		},
		{
			name: "json structure error",
			test: func(t *testing.T) {
				d := &struct {
					Info struct{}
				}{}

				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.json")},
					Raw:  []byte(`{ "openapi": "3.0", "info": []}`),
					Data: d,
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "structural error at info: expected object but received an array at line 1, column 29")
				require.Nil(t, c.Data)
			},
		},
		{
			name: "known json",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.json")},
					Raw:  []byte(`{"user": "foo"}`),
					Data: &data{},
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "foo", c.Data.(*data).User)
			},
		},
		{
			name: "unknown yaml",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.yaml")},
					Raw:  []byte(`user: foo`),
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Nil(t, c.Data)
			},
		},
		{
			name: "error yaml",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.yaml")},
					Raw:  []byte(`user: 'foo`),
					Data: &data{},
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "yaml: found unexpected end of stream")
			},
		},
		{
			name: "known yaml",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.yaml")},
					Raw:  []byte(`user: foo`),
					Data: &data{},
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "foo", c.Data.(*data).User)
			},
		},
		{
			name: "lua",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.lua")},
					Raw:  []byte(`print("Hello World")`),
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.IsType(t, &script.Script{}, c.Data)
				require.Equal(t, `print("Hello World")`, c.Data.(*script.Script).Code)
			},
		},
		{
			name: "javascript",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.js")},
					Raw:  []byte(`console.log('Hello World')`),
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.IsType(t, &script.Script{}, c.Data)
				require.Equal(t, `console.log('Hello World')`, c.Data.(*script.Script).Code)
			},
		},
		{
			name: "update javascript",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.js")},
					Raw:  []byte(`console.log('Hello World')`),
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.IsType(t, &script.Script{}, c.Data)
				require.Equal(t, `console.log('Hello World')`, c.Data.(*script.Script).Code)
			},
		},
		{
			name: "template and json",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.json.tmpl")},
					Raw:  []byte(`{"user": "{{ env "TEST_USER1" }}"}`),
					Data: &data{},
				}

				_ = os.Setenv("TEST_USER1", "foo")
				defer func() {
					_ = os.Unsetenv("TEST_USER1")
				}()

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "foo", c.Data.(*data).User)
			},
		},
		{
			name: "template syntax error",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.json.tmpl")},
					Raw:  []byte(`{"user": "{{ env "TEST_USER2" | foo }}"}`),
				}

				_ = os.Setenv("TEST_USER2", "foo")
				defer func() {
					_ = os.Unsetenv("TEST_USER2")
				}()

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "unable to render template foo.json.tmpl: template: :1: function \"foo\" not defined")
			},
		},
		{
			name: "template custom function",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.json.tmpl")},
					Raw:  []byte(`{"user": "{{ env "TEST_USER3" | extractUsername }}"}`),
					Data: &data{},
				}

				_ = os.Setenv("TEST_USER3", "foo\\bar")
				defer func() {
					_ = os.Unsetenv("TEST_USER3")
				}()

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "bar", c.Data.(*data).User)
			},
		},
		{
			name: "call parser",
			test: func(t *testing.T) {
				d := &data{}
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.json")},
					Raw:  []byte(`{"user": "foo"}`),
					Data: d,
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.True(t, d.calledParse, "parse function called")
			},
		},
		{
			name: "parser returns error",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.json")},
					Raw:  []byte(`{"user": "foo"}`),
					Data: &parseError{},
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "parsing file foo.json: TEST ERROR")
			},
		},
		{
			name: "opaque url",
			test: func(t *testing.T) {
				u := mustUrl("foo.json")
				u.Opaque = "foo.txt"

				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: u},
					Raw:  []byte(`foobar`),
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "foobar", c.Data)
			},
		},
		{
			name: "wrapped config",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.json")},
					Raw:  []byte(`{"user": "foo"}`),
					Data: &data{},
				}

				info := dynamic.ConfigInfo{
					Provider: "git",
					Url:      mustUrl("https://github.com/user/repo.git?file=/foo.json"),
				}

				dynamic.Wrap(info, c)

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "foo", c.Data.(*data).User)
			},
		},
		{
			name: "update existing data",
			test: func(t *testing.T) {
				type testType struct {
					Foo string
					Bar string
				}
				d := &testType{Bar: "bar"}

				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.json")},
					Raw:  []byte(`{"foo": "foo"}`),
					Data: d,
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "foo", c.Data.(*testType).Foo)
				require.Equal(t, "", c.Data.(*testType).Bar)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

type data struct {
	User        string
	calledParse bool
}

func (d *data) Parse(_ *dynamic.Config, _ dynamic.Reader) error {
	d.calledParse = true
	return nil
}

type parseError struct{}

func (d *parseError) Parse(_ *dynamic.Config, _ dynamic.Reader) error {
	return fmt.Errorf("TEST ERROR")
}

package dynamic_test

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				require.Equal(t, "Hello World", c.Data())
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
				require.Nil(t, c.Data())
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
				require.Nil(t, c.Data())
			},
		},
		{
			name: "json syntax error",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.json")},
					Raw:  []byte(`{"name", "foo"}`),
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "invalid character ',' after object key at line 1, column 8")
				require.Nil(t, c.Data())
			},
		},
		{
			name: "json structure error",
			test: func(t *testing.T) {
				d := &struct {
					Info struct{} `json:"info"`
				}{}

				c := dynamictest.NewConfig(d, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("foo.json")}), dynamictest.WithRaw([]byte(`{ "openapi": "3.0", "info": []}`)))

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "schema error at field 'info': expected object, got array at line 1, column 28")
				require.Nil(t, c.Data())
			},
		},
		{
			name: "known json",
			test: func(t *testing.T) {
				c := dynamictest.NewConfig(&data{}, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("foo.json")}), dynamictest.WithRaw([]byte(`{"user": "foo"}`)))

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "foo", c.Data().(*data).User)
			},
		},
		{
			name: "known json without extension",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo")},
					Raw:  []byte(`{"user": "foo"}`),
				}
				dynamic.Register("user", nil, &data{})

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "foo", c.Data().(*data).User)
				require.Equal(t, "application/json", c.Info.ContentType)
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
				require.Nil(t, c.Data())
			},
		},
		{
			name: "error yaml",
			test: func(t *testing.T) {
				c := dynamictest.NewConfig(&data{}, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("foo.yaml")}), dynamictest.WithRaw([]byte(`user: 'foo`)))

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "yaml: found unexpected end of stream")
			},
		},
		{
			name: "known yaml",
			test: func(t *testing.T) {
				c := dynamictest.NewConfig(&data{}, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("foo.yaml")}), dynamictest.WithRaw([]byte(`user: foo`)))

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "foo", c.Data().(*data).User)
			},
		},
		{
			name: "known yaml without extension",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo")},
					Raw:  []byte(`user: foo`),
				}
				dynamic.Register("user", nil, &data{})

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "foo", c.Data().(*data).User)
				require.Equal(t, "application/yaml", c.Info.ContentType)
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
				require.Equal(t, `print("Hello World")`, c.Data())
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
				require.Equal(t, `console.log('Hello World')`, c.Data())
			},
		},
		{
			name: "javascript without extension",
			test: func(t *testing.T) {
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo")},
					Raw:  []byte(`console.log('Hello World')`),
				}

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, `console.log('Hello World')`, c.Data())
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
				require.Equal(t, `console.log('Hello World')`, c.Data())
			},
		},
		{
			name: "template and json",
			test: func(t *testing.T) {
				c := dynamictest.NewConfig(&data{}, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("foo.json.tmpl")}), dynamictest.WithRaw([]byte(`{"user": "{{ env "TEST_USER1" }}"}`)))

				_ = os.Setenv("TEST_USER1", "foo")
				defer func() {
					_ = os.Unsetenv("TEST_USER1")
				}()

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "foo", c.Data().(*data).User)
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
			name: "call parser",
			test: func(t *testing.T) {
				d := &data{}
				c := dynamictest.NewConfig(d, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("foo.json")}), dynamictest.WithRaw([]byte(`{"user": "foo"}`)))

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.True(t, c.Data().(*data).calledParse, "parse function called")
			},
		},
		{
			name: "parser returns error",
			test: func(t *testing.T) {
				c := dynamictest.NewConfig(&parseError{}, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("foo.json")}), dynamictest.WithRaw([]byte(`{"user": "foo"}`)))

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
				require.Equal(t, "foobar", c.Data())
			},
		},
		{
			name: "wrapped config",
			test: func(t *testing.T) {
				c := dynamictest.NewConfig(&data{}, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("foo.json")}), dynamictest.WithRaw([]byte(`{"user": "foo"}`)))

				info := dynamic.ConfigInfo{
					Provider: "git",
					Url:      mustUrl("https://github.com/user/repo.git?file=/foo.json"),
				}

				dynamic.Wrap(info, c)

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "foo", c.Data().(*data).User)
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

				c := dynamictest.NewConfig(d, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("foo.json")}), dynamictest.WithRaw([]byte(`{"foo": "foo"}`)))

				err := dynamic.Parse(c, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "foo", c.Data().(*testType).Foo)
				require.Equal(t, "", c.Data().(*testType).Bar)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
			dynamic.Reset()
		})
	}
}

// Parse publishes a new object on every reparse. Whoever holds the object of an earlier parse
// (a store listener, a $ref resolution, a script) keeps a consistent old version.
func TestParse_Reparse(t *testing.T) {
	type untyped struct {
		Foo string
		Bar string
	}
	mustUrl := func(s string) *url.URL {
		u, err := url.Parse(s)
		require.NoError(t, err)
		return u
	}

	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "data of typed config held before reparse is unchanged",
			test: func(t *testing.T) {
				dynamic.Register("user", nil, &data{})
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: mustUrl("foo.yaml")},
					Raw:  []byte(`user: foo`),
				}
				require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))
				held := c.Data().(*data)

				c.Raw = []byte(`user: bar`)
				require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))

				require.Equal(t, "foo", held.User, "held object must not be reparsed in place")
				require.Equal(t, "bar", c.Data().(*data).User)
				require.NotSame(t, held, c.Data())
			},
		},
		{
			name: "data of untyped struct held before reparse is unchanged",
			test: func(t *testing.T) {
				c := dynamictest.NewConfig(&untyped{}, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("foo.json")}), dynamictest.WithRaw([]byte(`{"foo": "foo", "bar": "bar"}`)))
				require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))
				held := c.Data().(*untyped)

				c.Raw = []byte(`{"foo": "other"}`)
				require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))

				require.Equal(t, &untyped{Foo: "foo", Bar: "bar"}, held, "held object must not be reparsed in place")
				require.Equal(t, &untyped{Foo: "other"}, c.Data())
			},
		},
		{
			name: "failed reparse sets data to nil",
			test: func(t *testing.T) {
				c := dynamictest.NewConfig(&untyped{}, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("foo.json")}), dynamictest.WithRaw([]byte(`{"foo": "foo"}`)))
				require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))
				held := c.Data().(*untyped)

				c.Raw = []byte(`{"foo": `)
				require.Error(t, dynamic.Parse(c, &dynamictest.Reader{}))

				require.Nil(t, c.Data())
				require.Equal(t, "foo", held.Foo, "held object must stay as it was")
			},
		},
		{
			name: "failed hook keeps the previous data",
			test: func(t *testing.T) {
				c := dynamictest.NewConfig(&failingHook{}, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("foo.json")}), dynamictest.WithRaw([]byte(`{"foo": "foo"}`)))
				require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))
				held := c.Data().(*failingHook)

				c.Raw = []byte(`{"foo": "fail"}`)
				require.Error(t, dynamic.Parse(c, &dynamictest.Reader{}))

				require.Same(t, held, c.Data(), "unfinished object must not be published")
				require.Equal(t, "foo", held.Foo)
			},
		},
		{
			name: "reference added by the hook is registered on the config",
			test: func(t *testing.T) {
				ref := dynamictest.NewConfig(&untyped{}, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("file:/ref.json")}), dynamictest.WithRaw([]byte(`{}`)))
				reader := &dynamictest.Reader{Data: map[string]*dynamic.Config{"file:///ref.json": ref}}
				c := dynamictest.NewConfig(&refHook{}, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("file:/foo.json")}), dynamictest.WithRaw([]byte(`{}`)))

				require.NoError(t, dynamic.Parse(c, reader))

				require.Equal(t, []*dynamic.Config{ref}, c.Refs.List(false))
			},
		},
		{
			name: "target given by Read is only a type hint",
			test: func(t *testing.T) {
				// ConfigWatcher.Read sets the target of the caller as Data before the first parse
				target := &untyped{}
				c := dynamictest.NewConfig(target, dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustUrl("foo.json")}), dynamictest.WithRaw([]byte(`{"foo": "foo"}`)))

				require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))

				require.Equal(t, &untyped{Foo: "foo"}, c.Data(), "Data has the type of the hint")
				require.Equal(t, &untyped{}, target, "hint must not be filled in place")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
			dynamic.Reset()
		})
	}
}

// A reader that loops on Data (store listener, script, dashboard) while the ConfigWatcher
// reparses the file must neither race nor see a half-written object.
func TestParse_ReparseWhileReading(t *testing.T) {
	defer dynamic.Reset()
	dynamic.Register("user", nil, &data{})
	u, err := url.Parse("foo.yaml")
	require.NoError(t, err)
	c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}, Raw: []byte(`user: foo`)}
	require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				if d, ok := c.Data().(*data); ok {
					assert.Contains(t, []string{"foo", "bar"}, d.User, "reader must see a complete version")
				}
			}
		}
	}()

	for i := 0; i < 500; i++ {
		c.Raw = []byte(map[bool]string{true: `user: bar`, false: `user: foo`}[i%2 == 0])
		require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))
	}
	close(done)
	wg.Wait()
}

// failingHook fails in its Parse hook when Foo is "fail".
type failingHook struct {
	Foo string
}

func (f *failingHook) Parse(_ *dynamic.Config, _ dynamic.Reader) error {
	if f.Foo == "fail" {
		return fmt.Errorf("TEST ERROR")
	}
	return nil
}

// refHook reads file:/ref.json and registers it as a reference of the config it is parsed for.
type refHook struct{}

func (r *refHook) Parse(c *dynamic.Config, reader dynamic.Reader) error {
	ref, err := reader.Read(&url.URL{Scheme: "file", Path: "/ref.json"}, nil)
	if err != nil {
		return err
	}
	dynamic.AddRef(c, ref)
	return nil
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

func TestParse_ReparseScope(t *testing.T) {
	dynamic.Register("scopehook", nil, &scopeHook{})
	newConfig := func(raw string) *dynamic.Config {
		return &dynamic.Config{
			Info: dynamic.ConfigInfo{Url: &url.URL{Path: "/scope.yaml"}},
			Raw:  []byte(raw),
		}
	}

	t.Run("hook sees a fresh scope with the name of the config", func(t *testing.T) {
		c := newConfig(`scopehook: first`)
		require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))

		c.Raw = []byte(`scopehook: second`)
		require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))

		hook := c.Data().(*scopeHook)
		require.Equal(t, "/scope.yaml", hook.nameSeen)
		require.False(t, hook.anchorSeenBefore, "anchor of the previous parse must not be visible")
	})

	t.Run("dynamic anchor refers to the object of the latest parse", func(t *testing.T) {
		c := newConfig(`scopehook: first`)
		require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))
		c.Raw = []byte(`scopehook: second`)
		require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))

		v, err := c.Scope().GetDynamic("anchor")
		require.NoError(t, err)
		require.Same(t, c.Data(), v)
	})

	t.Run("data and scope of a reparse are published together", func(t *testing.T) {
		c := newConfig(`scopehook: first`)
		require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))
		firstScope := c.Scope()

		c.Raw = []byte(`scopehook: second`)
		require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))

		require.NotSame(t, firstScope, c.Scope())
		require.Equal(t, "second", c.Data().(*scopeHook).Scopehook)
		v, err := c.Scope().GetLexical("anchor")
		require.NoError(t, err)
		require.Same(t, c.Data(), v)
	})

	t.Run("failed hook keeps the data and scope of the previous parse", func(t *testing.T) {
		c := newConfig(`scopehook: first`)
		require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))
		previousData := c.Data()
		previousScope := c.Scope()

		c.Raw = []byte(`scopehook: fail`)
		require.Error(t, dynamic.Parse(c, &dynamictest.Reader{}))

		require.Same(t, previousData, c.Data())
		require.Same(t, previousScope, c.Scope())
		v, err := c.Scope().GetLexical("anchor")
		require.NoError(t, err)
		require.Equal(t, "first", v.(*scopeHook).Scopehook)
		_, err = c.Scope().GetLexical("partial")
		require.Error(t, err, "entries of the failed parse must not be published")
	})

	t.Run("failed data parse clears data and scope", func(t *testing.T) {
		c := newConfig(`scopehook: first`)
		require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))

		c.Raw = []byte(`scopehook: [`)
		require.Error(t, dynamic.Parse(c, &dynamictest.Reader{}))

		require.Nil(t, c.Data())
		_, err := c.Scope().GetLexical("anchor")
		require.Error(t, err, "anchors of the previous parse must not resolve")
	})
}

func TestResolveDynamic_ConfigWithoutData(t *testing.T) {
	_, err := dynamic.ResolveDynamic[*scopeHook]("#anchor", &dynamic.Config{}, &dynamictest.Reader{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "name 'anchor' not found")
}

func TestScope_LexicalAnchorOnConfigWithoutData(t *testing.T) {
	c := &dynamic.Config{}
	_, err := c.Scope().GetLexical("anchor")
	require.Error(t, err)
	require.Contains(t, err.Error(), "name 'anchor' not found")
}

// scopeHook registers its anchors in the scope of the config it is parsed for,
// a failing parse additionally registers "partial".
type scopeHook struct {
	Scopehook        string
	nameSeen         string
	anchorSeenBefore bool
}

func (h *scopeHook) Parse(c *dynamic.Config, _ dynamic.Reader) error {
	h.nameSeen = c.Scope().Name()
	c.Scope().OpenIfNeeded(c.Info.Path())
	_, err := c.Scope().GetLexical("anchor")
	h.anchorSeenBefore = err == nil

	if err = c.Scope().SetLexical("anchor", h); err != nil {
		return err
	}
	if err = c.Scope().SetDynamic("anchor", h); err != nil {
		return err
	}
	if h.Scopehook == "fail" {
		_ = c.Scope().SetLexical("partial", h)
		return fmt.Errorf("TEST ERROR")
	}
	return nil
}

// A watcher reparses a published config while another config resolves an anchor of it.
func TestParse_ReparseScopeWhileResolving(t *testing.T) {
	dynamic.Register("scopehook", nil, &scopeHook{})

	published := &dynamic.Config{
		Info: dynamic.ConfigInfo{Url: &url.URL{Scheme: "file", Path: "/b.yml"}},
		Raw:  []byte(`scopehook: first`),
	}
	require.NoError(t, dynamic.Parse(published, &dynamictest.Reader{}))

	reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
		return published, nil
	})
	referrer := &dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{Scheme: "file", Path: "/a.yml"}}}

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(done)
		for i := 0; i < 200; i++ {
			published.Raw = []byte(fmt.Sprintf("scopehook: v%v", i))
			assert.NoError(t, dynamic.Parse(published, &dynamictest.Reader{}))
		}
	}()

	for resolving := true; resolving; {
		select {
		case <-done:
			resolving = false
		default:
		}
		v, err := dynamic.ResolveDynamic[*scopeHook]("b.yml#anchor", referrer, reader)
		require.NoError(t, err)
		require.NotNil(t, v)
	}
	wg.Wait()
}

package dynamic_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"net/url"
	"testing"
)

func TestResolve(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "invalid ref",
			test: func(t *testing.T) {
				err := dynamic.Resolve(":80", "", &dynamic.Config{}, &dynamictest.Reader{})
				require.Error(t, err)
				require.EqualError(t, err, "parse \":80\": missing protocol scheme")
			},
		},
		{
			name: "resolve local reference root",
			test: func(t *testing.T) {
				type v struct {
					Foo string
				}
				s := v{Foo: "foo"}
				var result v

				err := dynamic.Resolve("#/", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Equal(t, s, result)
			},
		},
		{
			name: "resolve local reference",
			test: func(t *testing.T) {
				s := struct {
					Foo string
				}{Foo: "foo"}
				result := ""

				err := dynamic.Resolve("#/foo", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Equal(t, "foo", result)
			},
		},
		{
			name: "resolve local nested reference map",
			test: func(t *testing.T) {
				s := struct {
					Foo map[string]string
				}{Foo: map[string]string{"bar": "bar"}}
				result := ""

				err := dynamic.Resolve("#/foo/bar", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Equal(t, "bar", result)
			},
		},
		{
			name: "resolve local map entry not found",
			test: func(t *testing.T) {
				s := map[string]string{}
				result := ""

				err := dynamic.Resolve("#/foo", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.Error(t, err)
				require.EqualError(t, err, "resolve reference '#/foo' failed: invalid token reference \"foo\"")
			},
		},
		{
			name: "resolve local map when value is ref wrapper",
			test: func(t *testing.T) {
				type ref struct {
					Value string
				}
				s := map[string]ref{"foo": {Value: "foo"}}
				result := ""

				err := dynamic.Resolve("#/foo", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Equal(t, "foo", result)
			},
		},
		{
			name: "resolve local nested reference struct",
			test: func(t *testing.T) {
				type n struct {
					Bar string
				}
				s := struct {
					Foo n
				}{Foo: n{Bar: "bar"}}
				result := ""

				err := dynamic.Resolve("#/foo/bar", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Equal(t, "bar", result)
			},
		},
		{
			name: "resolve local nested reference struct with pointer",
			test: func(t *testing.T) {
				v := "bar"
				type n struct {
					Bar interface{}
				}
				s := struct {
					Foo n
				}{Foo: n{Bar: &v}}
				result := ""

				err := dynamic.Resolve("#/foo/bar", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Equal(t, "bar", result)
			},
		},
		{
			name: "resolve local struct field not found",
			test: func(t *testing.T) {
				s := struct{}{}
				result := ""

				err := dynamic.Resolve("#/foo", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.Error(t, err)
				require.EqualError(t, err, "resolve reference '#/foo' failed: invalid token reference \"foo\"")
			},
		},
		{
			name: "resolve local struct when field is ref wrapper",
			test: func(t *testing.T) {
				type ref struct {
					Value string
					Ref   interface{}
				}
				s := struct {
					Foo ref
				}{Foo: ref{Value: "foo"}}
				result := ""

				err := dynamic.Resolve("#/foo", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Equal(t, "foo", result)
			},
		},
		{
			name: "resolve local struct when field is ref wrapper but nil value",
			test: func(t *testing.T) {
				type ref struct {
					Value interface{}
					Ref   interface{}
				}
				s := struct {
					Foo ref
				}{Foo: ref{Value: nil}}
				result := ""

				err := dynamic.Resolve("#/foo", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.Error(t, err)
				require.EqualError(t, err, "resolve reference '#/foo' failed: path '/foo' not found")
			},
		},
		{
			name: "resolve local struct when field is ref wrapper and pointer",
			test: func(t *testing.T) {
				v := "foo"
				type ref struct {
					Value interface{}
					Ref   interface{}
				}
				s := struct {
					Foo ref
				}{Foo: ref{Value: &v}}
				result := ""

				err := dynamic.Resolve("#/foo", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Equal(t, "foo", result)
			},
		},
		{
			name: "resolve local nested struct with ref wrapper",
			test: func(t *testing.T) {
				type ref struct {
					Value map[string]string
				}
				s := struct {
					Foo ref
				}{Foo: ref{Value: map[string]string{"bar": "bar"}}}
				result := ""

				err := dynamic.Resolve("#/foo/bar", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Equal(t, "bar", result)
			},
		},
		{
			name: "resolve local reference float but int",
			test: func(t *testing.T) {
				s := struct {
					Foo int
				}{Foo: 12}
				var result float32

				err := dynamic.Resolve("#/foo", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.Error(t, err)
				require.EqualError(t, err, "resolve reference '#/foo' failed: expected type float32, got int")
			},
		},
		{
			name: "resolve local reference pointer to struct",
			test: func(t *testing.T) {
				type v struct {
					Value string
				}
				s := struct {
					Foo *v
				}{Foo: &v{Value: "foo"}}
				var result v

				err := dynamic.Resolve("#/foo", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Equal(t, "foo", result.Value)
			},
		},
		{
			name: "resolve local reference value is map",
			test: func(t *testing.T) {
				s := struct {
					Foo map[string]string
				}{Foo: map[string]string{"foo": "foo"}}
				var result map[string]string

				err := dynamic.Resolve("#/foo", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Contains(t, result, "foo")
			},
		},
		{
			name: "resolve local reference value implements PathResolver",
			test: func(t *testing.T) {
				s := struct {
					Foo *pathResolver
				}{Foo: &pathResolver{resolve: func(token string) (interface{}, error) {
					return "foo", nil
				}}}
				var result string

				err := dynamic.Resolve("#/foo", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Contains(t, result, "foo")
			},
		},
		{
			name: "resolve local reference value implements PathResolver but returns error",
			test: func(t *testing.T) {
				s := struct {
					Foo *pathResolver
				}{Foo: &pathResolver{resolve: func(token string) (interface{}, error) {
					return nil, fmt.Errorf("TEST ERROR")
				}}}
				var result string

				err := dynamic.Resolve("#/foo", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.Error(t, err)
				require.EqualError(t, err, "resolve reference '#/foo' failed: TEST ERROR")
			},
		},
		{
			name: "resolve local reference nested with PathResolver",
			test: func(t *testing.T) {
				s := struct {
					Foo *pathResolver
				}{Foo: &pathResolver{resolve: func(token string) (interface{}, error) {
					require.Equal(t, "bar", token)
					return "foo", nil
				}}}
				var result string

				err := dynamic.Resolve("#/foo/bar", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Contains(t, result, "foo")
			},
		},
		{
			name: "resolve local reference nested with PathResolver but error",
			test: func(t *testing.T) {
				s := struct {
					Foo *pathResolver
				}{Foo: &pathResolver{resolve: func(token string) (interface{}, error) {
					return nil, fmt.Errorf("TEST ERROR")
				}}}
				var result string

				err := dynamic.Resolve("#/foo/bar", &result, &dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.Error(t, err)
				require.EqualError(t, err, "resolve reference '#/foo/bar' failed: TEST ERROR")
			},
		},
		{
			name: "resolve global reference",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					require.Equal(t, "https://foo.bar", u.String())
					return &dynamic.Config{Data: "foo"}, nil
				})
				result := ""

				err := dynamic.Resolve("https://foo.bar", &result, &dynamic.Config{Info: dynamictest.NewConfigInfo()}, reader)

				require.NoError(t, err)
				require.Equal(t, "foo", result)
			},
		},
		{
			name: "resolve global reference but not found",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TESTING ERROR")
				})
				result := ""

				err := dynamic.Resolve("https://foo.bar", &result, &dynamic.Config{Info: dynamictest.NewConfigInfo()}, reader)

				require.Error(t, err)
				require.EqualError(t, err, "resolve reference 'https://foo.bar' failed: TESTING ERROR")
			},
		},
		{
			name: "resolve global reference with fragment",
			test: func(t *testing.T) {
				type value struct {
					Foo string
				}

				reader := dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					return &dynamic.Config{Data: value{Foo: "foo"}}, nil
				})
				result := ""

				cfg := &dynamic.Config{Info: dynamictest.NewConfigInfo(), Data: &value{}}
				err := dynamic.Resolve("https://foo.bar#/foo", &result, cfg, reader)

				require.NoError(t, err)
				require.Equal(t, "foo", result)
			},
		},
		{
			name: "resolve global reference with fragment but error",
			test: func(t *testing.T) {
				type value struct {
					Foo string
				}

				reader := dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					return &dynamic.Config{Data: value{Foo: "foo"}}, nil
				})
				result := ""

				cfg := &dynamic.Config{Info: dynamictest.NewConfigInfo(), Data: &value{}}
				err := dynamic.Resolve("https://foo.bar#/bar", &result, cfg, reader)

				require.Error(t, err)
				require.EqualError(t, err, "resolve reference 'https://foo.bar#/bar' failed: invalid token reference \"bar\"")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

type pathResolver struct {
	resolve func(token string) (interface{}, error)
}

func (p *pathResolver) Resolve(token string) (interface{}, error) {
	return p.resolve(token)
}

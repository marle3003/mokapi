package decoders

import (
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "args",
			f: func(t *testing.T) {
				s := &struct {
					Name  string
					Flag1 bool
					Flag2 bool
				}{
					Flag2: true,
				}
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--name=bar")
				os.Args = append(os.Args, "--flag1")
				os.Args = append(os.Args, "--no-flag2")

				err := Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Equal(t, "bar", s.Name)
				require.Equal(t, true, s.Flag1)
				require.Equal(t, false, s.Flag2)
			},
		},
		{
			name: "without =",
			f: func(t *testing.T) {
				s := &struct {
					Name string
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--name")
				os.Args = append(os.Args, "bar")

				err := Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Equal(t, "bar", s.Name)
			},
		},
		{
			name: "array explode",
			f: func(t *testing.T) {
				s := &struct {
					Names []string `explode:"name"`
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--name", "foo", "--name", "bar")

				err := Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, s.Names)
			},
		},
		{
			name: "array with item contains a space",
			f: func(t *testing.T) {
				s := &struct {
					Names []string
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--names")
				os.Args = append(os.Args, "bar foo \"foo bar\"")

				err := Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Equal(t, []string{"bar", "foo", "foo bar"}, s.Names)
			},
		},
		{
			name: "array override with index operator",
			f: func(t *testing.T) {
				s := &struct {
					Names []string `explode:"name"`
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--name", "foo", "--name", "bar", "--names[0]", "x")

				err := Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Equal(t, []string{"x", "bar"}, s.Names)
			},
		},
		{
			name: "env var",
			f: func(t *testing.T) {
				s := &struct {
					Name     string
					SkipName string `flag:"skip-name"`
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				err := os.Setenv("MOKAPI_name", "bar")
				defer os.Unsetenv("MOKAPI_name")
				require.NoError(t, err)
				err = os.Setenv("MOKAPI_SKIP_NAME", "bar")
				defer os.Unsetenv("MOKAPI_SKIP_NAME")
				require.NoError(t, err)

				err = Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Equal(t, "bar", s.Name)
				require.Equal(t, "bar", s.SkipName)
			},
		},
		{
			name: "env var overrides cli args",
			f: func(t *testing.T) {
				s := &struct {
					Name string
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--name=bar")
				err := os.Setenv("MOKAPI_name", "barr")
				defer os.Unsetenv("MOKAPI_name")
				require.NoError(t, err)

				err = Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Equal(t, "barr", s.Name)
			},
		},
		{
			name: "env var overrides cli array",
			f: func(t *testing.T) {
				s := &struct {
					Name []string
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--name", "foo", "--name", "bar")
				err := os.Setenv("MOKAPI_name", "barr")
				defer os.Unsetenv("MOKAPI_name")
				require.NoError(t, err)

				err = Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Equal(t, []string{"barr"}, s.Name)
			},
		},
		{
			name: "env var array single",
			f: func(t *testing.T) {
				s := &struct {
					Urls []string
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				err := os.Setenv("MOKAPI_urls", "https://foo.bar")
				defer os.Unsetenv("MOKAPI_urls")
				require.NoError(t, err)

				err = Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Contains(t, s.Urls, "https://foo.bar")
				require.Equal(t, 1, len(s.Urls))
				require.Equal(t, 1, cap(s.Urls))
			},
		},
		{
			name: "env var array",
			f: func(t *testing.T) {
				s := &struct {
					Urls []string
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				err := os.Setenv("MOKAPI_urls[0]", "https://foo.bar")
				defer os.Unsetenv("MOKAPI_urls[0]")
				require.NoError(t, err)
				err = os.Setenv("MOKAPI_urls[1]", "https://mokapi.io")
				require.NoError(t, err)
				defer os.Unsetenv("MOKAPI_urls[1]")
				err = os.Setenv("MOKAPI_urls_2", "https://foo.com")
				require.NoError(t, err)
				defer os.Unsetenv("MOKAPI_urls_2")
				err = os.Setenv("MOKAPI_urls.10", "https://bar.com")
				require.NoError(t, err)
				defer os.Unsetenv("MOKAPI_urls.10")

				err = Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Contains(t, s.Urls, "https://foo.bar")
				require.Contains(t, s.Urls, "https://mokapi.io")
				require.Contains(t, s.Urls, "https://foo.com")
				require.Contains(t, s.Urls, "https://bar.com")
				require.Equal(t, 11, len(s.Urls))
				require.Equal(t, 22, cap(s.Urls))
			},
		},
		{
			name: "env var array update index with [0]",
			f: func(t *testing.T) {
				s := &struct {
					Items []struct {
						Name  string
						Value int64
					}
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				err := os.Setenv("MOKAPI_items[0].name", "mokapi")
				defer os.Unsetenv("MOKAPI_items[0].name")
				require.NoError(t, err)
				err = os.Setenv("MOKAPI_items[0].value", "123")
				require.NoError(t, err)
				defer os.Unsetenv("MOKAPI_items[0].value")

				err = Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Len(t, s.Items, 1)
				require.Equal(t, s.Items[0].Name, "mokapi")
				require.Equal(t, s.Items[0].Value, int64(123))
			},
		},
		{
			name: "env var array update index with _0_",
			f: func(t *testing.T) {
				s := &struct {
					Items []struct {
						Name  string
						Value int64
					}
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				err := os.Setenv("MOKAPI_items_0_name", "mokapi")
				defer os.Unsetenv("MOKAPI_items_0_name")
				require.NoError(t, err)
				err = os.Setenv("MOKAPI_items_0_value", "123")
				require.NoError(t, err)
				defer os.Unsetenv("MOKAPI_items_0_value")

				err = Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Len(t, s.Items, 1)
				require.Equal(t, s.Items[0].Name, "mokapi")
				require.Equal(t, s.Items[0].Value, int64(123))
			},
		},
		{
			name: "env var array update index with .0.",
			f: func(t *testing.T) {
				s := &struct {
					Items []struct {
						Name  string
						Value int64
					}
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				err := os.Setenv("MOKAPI_items_0_name", "mokapi")
				defer os.Unsetenv("MOKAPI_items_0_name")
				require.NoError(t, err)
				err = os.Setenv("MOKAPI_ITEMS_0_VALUE", "123")
				require.NoError(t, err)
				defer os.Unsetenv("MOKAPI_ITEMS_0_VALUE.0.value")
				err = os.Setenv("MOKAPI_items.0.value", "123")
				require.NoError(t, err)
				defer os.Unsetenv("MOKAPI_items.0.value")

				err = Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Len(t, s.Items, 1)
				require.Equal(t, s.Items[0].Name, "mokapi")
				require.Equal(t, s.Items[0].Value, int64(123))
			},
		},
		{
			name: "unknown argument",
			f: func(t *testing.T) {
				s := &struct {
					Name string
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--foo=bar")

				err := Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.EqualError(t, err, "configuration error foo: not found")
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			os.Args = nil
			tc.f(t)
		})
	}
}

func TestLoad_Invalid(t *testing.T) {
	testcases := []struct {
		args string
	}{
		{args: "name=bar"},
		{args: "-"},
		{args: "--"},
		{args: "-=bar"},
		{args: "---name=bar"},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.args, func(t *testing.T) {
			os.Args = nil
			os.Args = append(os.Args, "mokapi.exe")
			args := strings.Split(tc.args, " ")
			os.Args = append(os.Args, args...)
			s := &struct {
				Name string
			}{}
			err := Load([]ConfigDecoder{&FlagDecoder{}}, s)
			require.Error(t, err)
		})
	}
}

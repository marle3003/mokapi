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
					Name string
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--name=bar")

				err := Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Equal(t, "bar", s.Name)
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
			name: "env var",
			f: func(t *testing.T) {
				s := &struct {
					Name string
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				err := os.Setenv("MOKAPI_name", "bar")
				defer os.Unsetenv("MOKAPI_name")
				require.NoError(t, err)

				err = Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Equal(t, "bar", s.Name)
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

				err = Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Contains(t, s.Urls, "https://foo.bar")
				require.Contains(t, s.Urls, "https://mokapi.io")
			},
		},
		{
			name: "env var array update index",
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
			name: "unknown argument",
			f: func(t *testing.T) {
				s := &struct {
					Name string
				}{}
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--foo=bar")

				err := Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.EqualError(t, err, "configuration error foo: configuration not found")
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
		{args: "--name"},
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

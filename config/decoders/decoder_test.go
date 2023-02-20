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
				defer os.Unsetenv("MOKAPI_foo")
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
				defer os.Unsetenv("MOKAPI_foo")
				require.NoError(t, err)

				err = Load([]ConfigDecoder{&FlagDecoder{}}, s)
				require.NoError(t, err)
				require.Equal(t, "barr", s.Name)
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

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.args, func(t *testing.T) {
			t.Parallel()
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

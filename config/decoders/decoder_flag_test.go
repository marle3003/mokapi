package decoders

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/provider/file/filetest"
	"testing"
)

func TestFlagDecoder_Decode(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "string",
			f: func(t *testing.T) {
				s := &struct {
					Name string
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string][]string{"name": {"foobar"}}, s)
				require.NoError(t, err)
				require.Equal(t, "foobar", s.Name)
			},
		},
		{
			name: "bool",
			f: func(t *testing.T) {
				s := &struct {
					Flag1 bool
					Flag2 bool
					Flag3 bool
					Flag4 bool
				}{
					Flag3: true,
					Flag4: true,
				}
				d := &FlagDecoder{}
				err := d.Decode(map[string][]string{"flag1": {"true"}, "flag2": {"1"}, "no-flag3": {""}, "no-flag4": {"false"}}, s)
				require.NoError(t, err)
				require.True(t, s.Flag1)
				require.True(t, s.Flag2)
				require.False(t, s.Flag3)
				require.True(t, s.Flag4)
			},
		},
		{
			name: "not a bool",
			f: func(t *testing.T) {
				s := &struct {
					Flag1 bool
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string][]string{"flag1": {"foo"}}, s)
				require.EqualError(t, err, "configuration error 'flag1' value '[foo]': value foo cannot be parsed as bool: strconv.ParseBool: parsing \"foo\": invalid syntax")
				require.False(t, s.Flag1)
			},
		},
		{
			name: "nested with dot (old)",
			f: func(t *testing.T) {
				s := &struct {
					Key   string
					Value struct {
						Flag bool
					}
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string][]string{"key": {"foo"}, "value.flag": {"true"}}, s)
				require.NoError(t, err)
				require.Equal(t, "foo", s.Key)
				require.True(t, s.Value.Flag)
			},
		},
		{
			name: "nested with - (new)",
			f: func(t *testing.T) {
				s := &struct {
					Key   string
					Value struct {
						Flag bool
					}
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string][]string{"key": {"foo"}, "value-flag": {"true"}}, s)
				require.NoError(t, err)
				require.Equal(t, "foo", s.Key)
				require.True(t, s.Value.Flag)
			},
		},
		{
			name: "capitalized",
			f: func(t *testing.T) {
				s := &struct {
					Key string
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string][]string{"Key": {"foo"}}, s)
				require.NoError(t, err)
				require.Equal(t, "foo", s.Key)
			},
		},
		{
			name: "map",
			f: func(t *testing.T) {
				s := &struct {
					Key map[string]string
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string][]string{"Key.foo": {"bar"}}, s)
				require.NoError(t, err)
				require.Equal(t, map[string]string{"foo": "bar"}, s.Key)
			},
		},
		{
			name: "array",
			f: func(t *testing.T) {
				s := &struct {
					Key []string
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string][]string{"Key[0]": {"bar"}, "Key[1]": {"foo"}}, s)
				require.NoError(t, err)
				require.Equal(t, []string{"bar", "foo"}, s.Key)
			},
		},
		{
			name: "array shorthand",
			f: func(t *testing.T) {
				s := &struct {
					Key []string
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string][]string{"Key": {"bar foo"}}, s)
				require.NoError(t, err)
				require.Equal(t, []string{"bar", "foo"}, s.Key)
			},
		},
		{
			name: "array shorthand with item contains a space",
			f: func(t *testing.T) {
				s := &struct {
					Key []string
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string][]string{"Key": {"bar foo \"foo bar\""}}, s)
				require.NoError(t, err)
				require.Equal(t, []string{"bar", "foo", "foo bar"}, s.Key)
			},
		},
		{
			name: "map with array",
			f: func(t *testing.T) {
				s := &struct {
					Key map[string][]string
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string][]string{"Key.foo[0]": {"bar"}}, s)
				require.NoError(t, err)
				require.Equal(t, map[string][]string{"foo": {"bar"}}, s.Key)
			},
		},
		{
			name: "map pointer struct",
			f: func(t *testing.T) {
				type test struct {
					Name string
					Foo  string
				}
				s := &struct {
					Key map[string]*test
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string][]string{"Key.foo.Name": {"Bob"}, "Key.foo.Foo": {"bar"}}, s)
				require.NoError(t, err)
				require.Equal(t, "Bob", s.Key["foo"].Name)
				require.Equal(t, "bar", s.Key["foo"].Foo)
			},
		},
		{
			name: "map struct",
			f: func(t *testing.T) {
				type test struct {
					Name string
					Foo  string
				}
				s := &struct {
					Key map[string]test
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string][]string{"Key.foo.Name": {"Bob"}, "Key.foo.Foo": {"bar"}}, s)
				require.NoError(t, err)
				require.Equal(t, "Bob", s.Key["foo"].Name)
				require.Equal(t, "bar", s.Key["foo"].Foo)
			},
		},
		{
			name: "parameters from file in current directory",
			f: func(t *testing.T) {
				type test struct {
					Name string
					Foo  string
				}
				s := &struct {
					Key map[string]test
				}{}

				fs := &filetest.MockFS{Entries: []*filetest.Entry{
					{
						Name:  "test.json",
						IsDir: false,
						Data:  []byte(`{"name": "Bob", "foo": "bar"}`),
					}}}

				d := &FlagDecoder{fs: fs}
				err := d.Decode(map[string][]string{"Key.foo": {"file://test.json"}}, s)
				require.NoError(t, err)
				require.Equal(t, "Bob", s.Key["foo"].Name)
				require.Equal(t, "bar", s.Key["foo"].Foo)
			},
		},
		{
			name: "parameters from file absolute path",
			f: func(t *testing.T) {
				type test struct {
					Name string
					Foo  string
				}
				s := &struct {
					Key map[string]test
				}{}

				fs := &filetest.MockFS{Entries: []*filetest.Entry{
					{
						Name:  "/tmp/test.json",
						IsDir: false,
						Data:  []byte(`{"name": "Bob", "foo": "bar"}`),
					}}}

				d := &FlagDecoder{fs: fs}
				err := d.Decode(map[string][]string{"Key.foo": {"file:///tmp/test.json"}}, s)
				require.NoError(t, err)
				require.Equal(t, "Bob", s.Key["foo"].Name)
				require.Equal(t, "bar", s.Key["foo"].Foo)
			},
		},
		{
			name: "parameters from file absolute path",
			f: func(t *testing.T) {
				s := &struct {
					SkipName string `flag:"skip-name"`
				}{}

				d := &FlagDecoder{}
				err := d.Decode(map[string][]string{"skip-name": {"foo"}}, s)
				require.NoError(t, err)
				require.Equal(t, "foo", s.SkipName)
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

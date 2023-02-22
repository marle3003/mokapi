package decoders

import (
	"github.com/stretchr/testify/require"
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
				err := d.Decode(map[string]string{"name": "foobar"}, s)
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
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string]string{"flag1": "true", "flag2": "1"}, s)
				require.NoError(t, err)
				require.True(t, s.Flag1)
				require.True(t, s.Flag2)
			},
		},
		{
			name: "not a bool",
			f: func(t *testing.T) {
				s := &struct {
					Flag1 bool
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string]string{"flag1": "foo"}, s)
				require.EqualError(t, err, "configuration error flag1: value foo cannot be parsed as bool: strconv.ParseBool: parsing \"foo\": invalid syntax")
				require.False(t, s.Flag1)
			},
		},
		{
			name: "nested",
			f: func(t *testing.T) {
				s := &struct {
					Key   string
					Value struct {
						Flag bool
					}
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string]string{"key": "foo", "value.flag": "true"}, s)
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
				err := d.Decode(map[string]string{"Key": "foo"}, s)
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
				err := d.Decode(map[string]string{"Key.foo": "bar"}, s)
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
				err := d.Decode(map[string]string{"Key[0]": "bar"}, s)
				require.NoError(t, err)
				require.Equal(t, []string{"bar"}, s.Key)
			},
		},
		{
			name: "map with array",
			f: func(t *testing.T) {
				s := &struct {
					Key map[string][]string
				}{}
				d := &FlagDecoder{}
				err := d.Decode(map[string]string{"Key.foo[0]": "bar"}, s)
				require.NoError(t, err)
				require.Equal(t, map[string][]string{"foo": {"bar"}}, s.Key)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.f(t)
		})
	}
}

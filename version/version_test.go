package version

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestVersion(t *testing.T) {
	testcases := []struct {
		name string
		s    string
		test func(t *testing.T, v Version)
	}{
		{
			name: "empty",
			s:    "",
			test: func(t *testing.T, v Version) {
				require.Equal(t, "0.0.0", v.String())
				require.True(t, v.IsEmpty())
			},
		},
		{
			name: "major",
			s:    "1",
			test: func(t *testing.T, v Version) {
				require.Equal(t, "1.0.0", v.String())
				require.Equal(t, 1, v.Major)
				require.False(t, v.IsEmpty())
			},
		},
		{
			name: "major error",
			s:    "a",
			test: func(t *testing.T, v Version) {
				require.Equal(t, "0.0.0", v.String())
			},
		},
		{
			name: "minor",
			s:    "1.1",
			test: func(t *testing.T, v Version) {
				require.Equal(t, "1.1.0", v.String())
				require.Equal(t, 1, v.Minor)
			},
		},
		{
			name: "minor error",
			s:    "1.b",
			test: func(t *testing.T, v Version) {
				require.Equal(t, "1.0.0", v.String())
			},
		},
		{
			name: "patch",
			s:    "1.1.1",
			test: func(t *testing.T, v Version) {
				require.Equal(t, "1.1.1", v.String())
				require.Equal(t, 1, v.Patch)
			},
		},
		{
			name: "patch error",
			s:    "1.1.c",
			test: func(t *testing.T, v Version) {
				require.Equal(t, "1.1.0", v.String())
			},
		},
		{
			name: "IsEqual",
			s:    "1.1.0",
			test: func(t *testing.T, v Version) {
				require.True(t, v.IsEqual(New("1.1.0")))
				require.False(t, v.IsEqual(New("1.1.1")))
			},
		},
		{
			name: "IsLower",
			s:    "1.1.0",
			test: func(t *testing.T, v Version) {
				require.True(t, v.IsLower(New("1.1.1")))
				require.False(t, v.IsLower(New("1.1.0")))
			},
		},
		{
			name: "IsSupported",
			s:    "1.1.0",
			test: func(t *testing.T, v Version) {
				require.True(t, v.IsSupported(New("1.1.0")))
				require.False(t, v.IsSupported(New("1.1.1")))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			v := New(tc.s)
			tc.test(t, v)
		})
	}
}

func TestVersion_UnmarshalJSON(t *testing.T) {
	var v Version
	err := json.Unmarshal([]byte(`"1.2.3"`), &v)
	require.NoError(t, err)
	require.Equal(t, "1.2.3", v.String())

	err = json.Unmarshal([]byte(`{}`), &v)
	require.Error(t, err)
}

func TestVersion_UnmarshalYAML(t *testing.T) {
	var v Version
	err := yaml.Unmarshal([]byte(`1.2.3`), &v)
	require.NoError(t, err)
	require.Equal(t, "1.2.3", v.String())

	err = yaml.Unmarshal([]byte(`{}`), &v)
	require.Error(t, err)
}

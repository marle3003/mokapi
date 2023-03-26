package decoders

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestFileDecoder_Decode(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "no filename set",
			f: func(t *testing.T) {
				s := &struct{ Name string }{}
				d := &FileDecoder{}
				err := d.Decode(map[string]string{}, s)
				require.NoError(t, err)
			},
		},
		{
			name: "/etc/mokapi/mokapi.yaml",
			f: func(t *testing.T) {
				s := &struct{ Name string }{}
				f := func(path string) ([]byte, error) {
					if path == "/etc/mokapi/mokapi.yaml" {
						return []byte("name: foobar"), nil
					}
					return nil, fmt.Errorf("not found")
				}
				d := &FileDecoder{readFile: f}
				err := d.Decode(map[string]string{}, s)
				require.NoError(t, err)
				require.Equal(t, "foobar", s.Name)
			},
		},
		{
			name: "file does not exist",
			f: func(t *testing.T) {
				s := &struct{ Name string }{}
				d := &FileDecoder{filename: "test.yml"}
				err := d.Decode(map[string]string{}, s)
				require.Contains(t, err.Error(), "open test.yml:")
			},
		},
		{
			name: "empty file",
			f: func(t *testing.T) {
				s := &struct{ Name string }{}
				f := func(path string) ([]byte, error) { return []byte(""), nil }
				d := &FileDecoder{filename: "mokapi.yaml", readFile: f}
				err := d.Decode(map[string]string{}, s)
				require.NoError(t, err)
			},
		},
		{
			name: "yaml schema error",
			f: func(t *testing.T) {
				s := &struct{ Name string }{}
				f := func(path string) ([]byte, error) { return []byte("name: {}"), nil }
				d := &FileDecoder{filename: "mokapi.yml", readFile: f}
				err := d.Decode(map[string]string{}, s)
				require.EqualError(t, err, "parsing YAML file mokapi.yml: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!map into string")
			},
		},
		{
			name: "file with data",
			f: func(t *testing.T) {
				s := &struct{ Name string }{}
				f := func(path string) ([]byte, error) { return []byte("name: foobar"), nil }
				d := &FileDecoder{readFile: f}
				err := d.Decode(map[string]string{"configfile": "mokapi.yaml"}, s)
				require.NoError(t, err)
				require.Equal(t, "foobar", s.Name)
			},
		},
		{
			name: "temp file with data",
			f: func(t *testing.T) {
				s := &struct{ Name string }{}
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "name: foobar")}
				err := d.Decode(map[string]string{"configfile": d.filename}, s)
				require.NoError(t, err)
				require.Equal(t, "foobar", s.Name)
			},
		},
		{
			name: "pascal case",
			f: func(t *testing.T) {
				s := &struct {
					InstallDir string `yaml:"installDir"`
				}{}
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "installDir: foobar")}
				err := d.Decode(map[string]string{"configfile": d.filename}, s)
				require.NoError(t, err)
				require.Equal(t, "foobar", s.InstallDir)
			},
		},
		{
			name: "map",
			f: func(t *testing.T) {
				s := &struct {
					Key map[string]string
				}{}
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "key: {foo: bar}")}
				err := d.Decode(map[string]string{"configFile": d.filename}, s)
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
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "key: [bar]")}
				err := d.Decode(map[string]string{"configFile": d.filename}, s)
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
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "key: {foo: [bar]}")}
				err := d.Decode(map[string]string{"configFile": d.filename}, s)
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
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "key: {foo: {name: Bob, foo: bar}}")}
				err := d.Decode(map[string]string{"configFile": d.filename}, s)
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
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "key: {foo: {name: Bob, foo: bar}}")}
				err := d.Decode(map[string]string{"configFile": d.filename}, s)
				require.NoError(t, err)
				require.Equal(t, "Bob", s.Key["foo"].Name)
				require.Equal(t, "bar", s.Key["foo"].Foo)
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

func createTempFile(t *testing.T, filename, data string) string {
	path := filepath.Join(t.TempDir(), filename)
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	_, err = file.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	return path
}

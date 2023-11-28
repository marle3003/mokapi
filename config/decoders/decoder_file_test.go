package decoders

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestFileDecoder_Decode(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "no filename set",
			test: func(t *testing.T) {
				s := &struct{ Name string }{}
				d := NewDefaultFileDecoder()
				err := d.Decode(map[string]string{}, s)
				require.NoError(t, err)
			},
		},
		{
			name: "file in folder mokapi in etc",
			test: func(t *testing.T) {
				s := &struct{ Name string }{}
				f := func(path string) ([]byte, error) {
					// if test is executed on windows we get second path
					if path == "/etc/mokapi/mokapi.yaml" || path == "\\etc\\mokapi\\mokapi.yaml" {
						return []byte("name: foobar"), nil
					}
					return nil, fs.ErrNotExist
				}
				d := &FileDecoder{readFile: f}
				err := d.Decode(map[string]string{}, s)
				require.NoError(t, err)
				require.Equal(t, "foobar", s.Name)
			},
		},
		{
			name: "file does not exist",
			test: func(t *testing.T) {
				s := &struct{ Name string }{}
				f := func(path string) ([]byte, error) { return []byte(""), fmt.Errorf("file not found") }
				d := &FileDecoder{filename: "test.yml", readFile: f}
				err := d.Decode(map[string]string{}, s)
				require.Error(t, err)
			},
		},
		{
			name: "empty file",
			test: func(t *testing.T) {
				s := &struct{ Name string }{}
				f := func(path string) ([]byte, error) { return []byte(""), nil }
				d := &FileDecoder{filename: "mokapi.yaml", readFile: f}
				err := d.Decode(map[string]string{}, s)
				require.NoError(t, err)
			},
		},
		{
			name: "yaml schema error",
			test: func(t *testing.T) {
				s := &struct{ Name string }{}
				f := func(path string) ([]byte, error) { return []byte("name: {}"), nil }
				d := &FileDecoder{filename: "mokapi.yml", readFile: f}
				err := d.Decode(map[string]string{}, s)
				require.EqualError(t, err, "parse file 'mokapi.yml' failed: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!map into string")
			},
		},
		{
			name: "file with data",
			test: func(t *testing.T) {
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
			test: func(t *testing.T) {
				s := &struct{ Name string }{}
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "name: foobar"), readFile: os.ReadFile}
				err := d.Decode(map[string]string{"configfile": d.filename}, s)
				require.NoError(t, err)
				require.Equal(t, "foobar", s.Name)
			},
		},
		{
			name: "pascal case",
			test: func(t *testing.T) {
				s := &struct {
					InstallDir string `yaml:"installDir"`
				}{}
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "installDir: foobar"), readFile: os.ReadFile}
				err := d.Decode(map[string]string{"configfile": d.filename}, s)
				require.NoError(t, err)
				require.Equal(t, "foobar", s.InstallDir)
			},
		},
		{
			name: "map",
			test: func(t *testing.T) {
				s := &struct {
					Key map[string]string
				}{}
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "key: {foo: bar}"), readFile: os.ReadFile}
				err := d.Decode(map[string]string{"configFile": d.filename}, s)
				require.NoError(t, err)
				require.Equal(t, map[string]string{"foo": "bar"}, s.Key)
			},
		},
		{
			name: "array",
			test: func(t *testing.T) {
				s := &struct {
					Key []string
				}{}
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "key: [bar]"), readFile: os.ReadFile}
				err := d.Decode(map[string]string{"configFile": d.filename}, s)
				require.NoError(t, err)
				require.Equal(t, []string{"bar"}, s.Key)
			},
		},
		{
			name: "map with array",
			test: func(t *testing.T) {
				s := &struct {
					Key map[string][]string
				}{}
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "key: {foo: [bar]}"), readFile: os.ReadFile}
				err := d.Decode(map[string]string{"configFile": d.filename}, s)
				require.NoError(t, err)
				require.Equal(t, map[string][]string{"foo": {"bar"}}, s.Key)
			},
		},
		{
			name: "map pointer struct",
			test: func(t *testing.T) {
				type test struct {
					Name string
					Foo  string
				}
				s := &struct {
					Key map[string]*test
				}{}
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "key: {foo: {name: Bob, foo: bar}}"), readFile: os.ReadFile}
				err := d.Decode(map[string]string{"configFile": d.filename}, s)
				require.NoError(t, err)
				require.Equal(t, "Bob", s.Key["foo"].Name)
				require.Equal(t, "bar", s.Key["foo"].Foo)
			},
		},
		{
			name: "map struct",
			test: func(t *testing.T) {
				type test struct {
					Name string
					Foo  string
				}
				s := &struct {
					Key map[string]test
				}{}
				d := &FileDecoder{filename: createTempFile(t, "test.yml", "key: {foo: {name: Bob, foo: bar}}"), readFile: os.ReadFile}
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
			tc.test(t)
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

package cli_test

import (
	"fmt"
	"io/fs"
	"mokapi/pkg/cli"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileDecoder_Decode(t *testing.T) {
	newCmd := func(args []string, cfg any) *cli.Command {
		c := &cli.Command{EnvPrefix: "Mokapi_"}
		c.SetArgs(args)
		c.Config = cfg
		c.Run = func(cmd *cli.Command, args []string) error {
			return nil
		}
		return c
	}

	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
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
				cli.SetReadFileFS(f)
				c := newCmd([]string{}, &s)
				c.Flags().String("name", "", "")
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, "foobar", s.Name)
			},
		},
		{
			name: "file does not exist",
			test: func(t *testing.T) {
				s := &struct{ Name string }{}
				f := func(path string) ([]byte, error) { return []byte(""), fmt.Errorf("file not found") }
				cli.SetReadFileFS(f)
				c := newCmd([]string{}, &s)
				err := c.Execute()
				require.Error(t, err)
			},
		},
		{
			name: "empty file",
			test: func(t *testing.T) {
				s := &struct{ Name string }{}
				f := func(path string) ([]byte, error) { return []byte(""), nil }
				cli.SetReadFileFS(f)
				c := newCmd([]string{}, &s)
				c.Flags().String("name", "", "")
				err := c.Execute()
				require.NoError(t, err)
			},
		},
		{
			name: "yaml schema error",
			test: func(t *testing.T) {
				s := &struct{ Name int }{}
				f := func(path string) ([]byte, error) { return []byte("name: {}"), nil }
				cli.SetReadFileFS(f)
				c := newCmd([]string{}, &s)
				c.Flags().String("name", "", "")
				err := c.Execute()
				require.EqualError(t, err, "parse file 'mokapi.yaml' failed: cannot unmarshal object into int")
			},
		},
		{
			name: "temp file with data",
			test: func(t *testing.T) {
				s := &struct{ Name string }{}
				path := createTempFile(t, "test.yml", "name: foobar")

				c := newCmd([]string{"--config-file", path}, s)
				c.Flags().String("name", "", "")
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, "foobar", s.Name)
			},
		},
		{
			name: "pascal case",
			test: func(t *testing.T) {
				s := &struct {
					InstallDir string `yaml:"installDir" flag:"install-dir"`
				}{}
				path := createTempFile(t, "test.yml", "installDir: foobar")
				c := newCmd([]string{"--config-file", path}, s)
				c.Flags().String("install-dir", "", "")

				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, "foobar", s.InstallDir)
			},
		},
		{
			name: "map",
			test: func(t *testing.T) {
				s := &struct {
					Values map[string]string
				}{}
				path := createTempFile(t, "test.yml", "values: {foo: bar}")
				c := newCmd([]string{"--config-file", path}, s)
				c.Flags().DynamicString("values-<key>", "", "")

				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, map[string]string{"foo": "bar"}, s.Values)
			},
		},
		{
			name: "array",
			test: func(t *testing.T) {
				s := &struct {
					Key []string
				}{}
				path := createTempFile(t, "test.yml", "key: [bar]")
				c := newCmd([]string{"--config-file", path}, s)
				c.Flags().DynamicString("key[<index>]", "", "")

				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, []string{"bar"}, s.Key)
			},
		},
		{
			name: "map with array",
			test: func(t *testing.T) {
				s := &struct {
					Values map[string][]string
				}{}
				path := createTempFile(t, "test.yml", "values: {foo: [bar]}")
				c := newCmd([]string{"--config-file", path}, s)
				c.Flags().DynamicString("values-<key>[<index>]", "", "")

				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, map[string][]string{"foo": {"bar"}}, s.Values)
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
					Values map[string]*test
				}{}
				path := createTempFile(t, "test.yml", "values: {foo: {name: Bob, foo: bar}}")
				c := newCmd([]string{"--config-file", path}, s)
				c.Flags().DynamicString("values-<key>-name", "", "")
				c.Flags().DynamicString("values-<key>-foo", "", "")

				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, "Bob", s.Values["foo"].Name)
				require.Equal(t, "bar", s.Values["foo"].Foo)
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
					Values map[string]test
				}{}
				path := createTempFile(t, "test.yml", "values: {foo: {name: Bob, foo: bar}}")
				c := newCmd([]string{"--config-file", path}, s)
				c.Flags().DynamicString("values-<key>-name", "", "")
				c.Flags().DynamicString("values-<key>-foo", "", "")

				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, "Bob", s.Values["foo"].Name)
				require.Equal(t, "bar", s.Values["foo"].Foo)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				cli.SetReadFileFS(os.ReadFile)
			}()

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
	defer func() {
		_ = file.Close()
	}()

	_, err = file.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	return path
}

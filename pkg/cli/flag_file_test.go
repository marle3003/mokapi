package cli_test

import (
	"mokapi/pkg/cli"
	"mokapi/pkg/cli/clitest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileDecoder_Decode(t *testing.T) {
	newCmd := func(args []string, cfg any) *cli.Command {
		c := &cli.Command{Name: "foo"}
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
			name: "read existing file",
			test: func(t *testing.T) {
				s := &struct{ Name string }{}
				cli.SetFileReader(&clitest.TestFileReader{Files: map[string][]byte{
					"/etc/foo/foo.yaml": []byte("name: foobar"),
				}})
				c := newCmd([]string{}, &s)
				c.SetConfigPath("/etc/foo")
				c.Flags().String("name", "", cli.FlagDoc{})
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, "foobar", s.Name)
			},
		},
		{
			name: "file does not exist should not return error when not file flag is set",
			test: func(t *testing.T) {
				s := &struct{ Name string }{}
				cli.SetFileReader(&clitest.TestFileReader{Files: map[string][]byte{}})
				c := newCmd([]string{}, &s)
				c.SetConfigPath("/etc/foo")
				err := c.Execute()
				require.NoError(t, err)
			},
		},
		{
			name: "empty file",
			test: func(t *testing.T) {
				s := &struct{ Name string }{}
				cli.SetFileReader(&clitest.TestFileReader{Files: map[string][]byte{"/etc/foo": []byte("")}})
				c := newCmd([]string{}, &s)
				c.SetConfigPath("/etc/foo")
				c.Flags().String("name", "", cli.FlagDoc{})
				err := c.Execute()
				require.NoError(t, err)
			},
		},
		{
			name: "yaml schema error",
			test: func(t *testing.T) {
				s := &struct{ Count string }{}
				cli.SetFileReader(&clitest.TestFileReader{Files: map[string][]byte{"/etc/foo.yaml": []byte("count: foo")}})
				c := newCmd([]string{}, &s)
				c.SetConfigPath("/etc")
				c.Flags().Int("count", 0, cli.FlagDoc{})
				err := c.Execute()
				require.EqualError(t, err, "failed to set flag count: parsing foo: invalid syntax")
			},
		},
		{
			name: "use file flag",
			test: func(t *testing.T) {
				s := &struct{ Name string }{}
				path := createTempFile(t, "test.yml", "name: foobar")

				c := newCmd([]string{"--config-file", path}, s)
				c.Flags().String("name", "", cli.FlagDoc{})
				c.Flags().File("config-file", cli.FlagDoc{})
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, "foobar", s.Name)
			},
		},
		{
			name: "file does not exist should return error when file flag is set",
			test: func(t *testing.T) {
				s := &struct{ Name string }{}
				cli.SetFileReader(&clitest.TestFileReader{Files: map[string][]byte{}})
				c := newCmd([]string{"--config-file", "/etc/foo.yaml"}, s)
				c.Flags().File("config-file", cli.FlagDoc{})
				err := c.Execute()
				require.EqualError(t, err, "read config file '/etc/foo.yaml' failed: file does not exist")
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
				c.Flags().String("install-dir", "", cli.FlagDoc{})
				c.Flags().File("config-file", cli.FlagDoc{})
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
				c.Flags().DynamicString("values-<key>", cli.FlagDoc{})
				c.Flags().File("config-file", cli.FlagDoc{})
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
				c.Flags().DynamicString("key[<index>]", cli.FlagDoc{})
				c.Flags().File("config-file", cli.FlagDoc{})
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
				c.Flags().DynamicString("values-<key>[<index>]", cli.FlagDoc{})
				c.Flags().File("config-file", cli.FlagDoc{})
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
				c.Flags().DynamicString("values-<key>-name", cli.FlagDoc{})
				c.Flags().DynamicString("values-<key>-foo", cli.FlagDoc{})
				c.Flags().File("config-file", cli.FlagDoc{})
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
				c.Flags().DynamicString("values-<key>-name", cli.FlagDoc{})
				c.Flags().DynamicString("values-<key>-foo", cli.FlagDoc{})
				c.Flags().File("config-file", cli.FlagDoc{})
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
				cli.SetFileReader(&cli.FileReader{})
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

package mokapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"mokapi/config/static"
	"mokapi/pkg/cli"
	"mokapi/pkg/cmd/mokapi"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

/*func TestMokapi_Cmd(t *testing.T) {
	stdOut := os.Stdout
	stdErr := os.Stderr

	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = writer
	os.Stderr = writer
	defer func() {
		os.Stdout = stdOut
		os.Stderr = stdErr
	}()

	os.Args = nil
	os.Args = append(os.Args, "mokapi.exe")
	os.Args = append(os.Args, []string{"version"}...)

	cmd := mokapi.NewCmdMokapi(context.Background())
	err = cmd.Execute()
	require.NoError(t, err)

	_ = writer.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, reader)
	_ = reader.Close()

	require.Equal(t, "", buf.String())
}*/

func TestMain_Flags(t *testing.T) {
	testcases := []struct {
		name string
		args []string
		test func(t *testing.T, cfg *static.Config)
	}{
		{
			name: "default",
			args: []string{},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "info", cfg.Log.Level)
				require.Equal(t, "text", cfg.Log.Format)
				require.Equal(t, 8080, cfg.Api.Port)
				require.Equal(t, true, cfg.Api.Dashboard)
				require.Equal(t, []string{"_"}, cfg.Providers.File.SkipPrefix)
				require.Equal(t, int64(100), cfg.Event.Store["default"].Size)
				require.Equal(t, "0.85", cfg.DataGen.OptionalProperties)
			},
		},
		{
			name: "--providers-file-filenames",
			args: []string{"--providers-file-filenames", "foo.yaml"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "--providers-file-filenames comma separated",
			args: []string{"--providers-file-filenames", "foo.yaml,bar.yaml"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo.yaml", "bar.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "--providers-file-filenames twice",
			args: []string{"--providers-file-filenames", "foo.yaml", "--providers-file-filenames", "bar.yaml"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo.yaml", "bar.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "--providers-file-filename",
			args: []string{"--providers-file-filename", "foo.yaml"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "--event-store",
			args: []string{"--event-store", "foo={\"size\":250}"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, static.Store{Size: 250}, cfg.Event.Store["foo"])
			},
		},
		{
			name: "--event-store-foo-size",
			args: []string{"--event-store-foo-size", "250"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, int64(250), cfg.Event.Store["foo"].Size)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := mokapi.NewCmdMokapi(context.Background())
			cmd.SetArgs(tc.args)

			cfg := static.NewConfig()
			cmd.Run = func(cmd *cli.Command, args []string) error {
				cfg = cmd.Config.(*static.Config)
				return nil
			}
			err := cmd.Execute()
			require.NoError(t, err)

			tc.test(t, cfg)
		})
	}
}

// Tests from old design
func TestStaticConfig(t *testing.T) {
	testcases := []struct {
		name string
		args []string
		test func(t *testing.T, cfg *static.Config)
	}{
		{
			name: "assign with =",
			args: []string{"--log-level=debug"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "debug", cfg.Log.Level)
			},
		},
		{
			name: "assign without =",
			args: []string{"--log-level", "debug"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "debug", cfg.Log.Level)
			},
		},
		{
			name: "--help",
			args: []string{"--help"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, true, cfg.Help)
			},
		},
		{
			name: "-h",
			args: []string{"-h"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, true, cfg.Help)
			},
		},
		{
			name: "--version",
			args: []string{"--version"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, true, cfg.Version)
			},
		},
		{
			name: "-v",
			args: []string{"-v"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, true, cfg.Version)
			},
		},
		{
			name: "json",
			args: []string{`--providers-file={"filename":"foo.yaml","directory":"foo", "skipPrefix":["_"]}`},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo.yaml"}, cfg.Providers.File.Filenames)
				require.Equal(t, []string{"foo"}, cfg.Providers.File.Directories)
				require.Equal(t, []string{"_"}, cfg.Providers.File.SkipPrefix)
			},
		},
		{
			name: "shorthand object",
			args: []string{"--providers-file", "filename=foo.yaml"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "args",
			args: []string{
				"--providers-git-repositories[0]-url=https://github.com/PATH-TO/REPOSITORY?ref=branch-name",
				"--providers-git-repositories[0]-pull-interval=5m",
				"--providers-git-repositories[1]-pull-interval=5h",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "https://github.com/PATH-TO/REPOSITORY?ref=branch-name", cfg.Providers.Git.Repositories[0].Url)
				require.Equal(t, "5m", cfg.Providers.Git.Repositories[0].PullInterval)
				require.Equal(t, "", cfg.Providers.Git.Repositories[1].Url)
				require.Equal(t, "5h", cfg.Providers.Git.Repositories[1].PullInterval)
			},
		},
		{
			name: "shorthand array",
			args: []string{
				"--providers-git-repositories",
				"url=foo,pullInterval=5m url=bar,pullInterval=5h",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Len(t, cfg.Providers.Git.Repositories, 2)
				require.Equal(t, "foo", cfg.Providers.Git.Repositories[0].Url)
				require.Equal(t, "5m", cfg.Providers.Git.Repositories[0].PullInterval)
				require.Equal(t, "bar", cfg.Providers.Git.Repositories[1].Url)
				require.Equal(t, "5h", cfg.Providers.Git.Repositories[1].PullInterval)
			},
		},
		{
			name: "explode with json",
			args: []string{
				`--providers-git-repository={"url":"https://github.com/PATH-TO/REPOSITORY?ref=branch-name","pullInterval":"5m"}`,
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "https://github.com/PATH-TO/REPOSITORY?ref=branch-name", cfg.Providers.Git.Repositories[0].Url)
				require.Equal(t, "5m", cfg.Providers.Git.Repositories[0].PullInterval)
			},
		},
		{
			name: "file provider include",
			args: []string{
				"--providers-file-include",
				`mokapi/**/*.json mokapi/**/*.yaml "foo bar/**/*.yaml"`,
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"mokapi/**/*.json", "mokapi/**/*.yaml", "foo bar/**/*.yaml"}, cfg.Providers.File.Include)
			},
		},
		{
			name: "file provider include with space",
			args: []string{
				"--Providers.file.include",
				`"C:\Documents and Settings\" C:\Work"`,
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"C:\\Documents and Settings\\", "C:\\Work"}, cfg.Providers.File.Include)
			},
		},
		{
			name: "file provider include twice",
			args: []string{
				"--providers-file-include", "foo",
				"--providers-file-include", "bar",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo", "bar"}, cfg.Providers.File.Include)
			},
		},
		{
			name: "file provider include overwrite",
			args: []string{
				"--providers-file-include", "foo",
				"--providers-file-include[0]", "bar",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"bar"}, cfg.Providers.File.Include)
			},
		},
		{
			name: "git provider set url",
			args: []string{
				"--providers-git-url", "foo",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo"}, cfg.Providers.Git.Urls)
			},
		},
		{
			name: "git provider set urls",
			args: []string{
				"--providers-git-urls", "foo",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo"}, cfg.Providers.Git.Urls)
			},
		},
		{
			name: "http provider set url",
			args: []string{
				"--providers-http-url", "foo",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo"}, cfg.Providers.Http.Urls)
			},
		},
		{
			name: "http provider set urls using explode",
			args: []string{
				"--providers-http-url", "foo",
				"--providers-http-url", "bar",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo", "bar"}, cfg.Providers.Http.Urls)
			},
		},
		{
			name: "http provider set urls",
			args: []string{
				"--providers-http-urls", "foo bar",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo", "bar"}, cfg.Providers.Http.Urls)
			},
		},
		{
			name: "http provider",
			args: []string{
				"--providers-http", `urls=foo bar,pollInterval=5s,pollTimeout=30s,proxy=bar,tlsSkipVerify=true`,
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo", "bar"}, cfg.Providers.Http.Urls)
				require.Equal(t, "5s", cfg.Providers.Http.PollInterval)
				require.Equal(t, "30s", cfg.Providers.Http.PollTimeout)
				require.Equal(t, true, cfg.Providers.Http.TlsSkipVerify)
				require.Equal(t, "bar", cfg.Providers.Http.Proxy)
			},
		},
		{
			name: "npm provider global folders",
			args: []string{
				"--providers-npm-global-folders", "/etc/foo",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "/etc/foo", cfg.Providers.Npm.GlobalFolders[0])
			},
		},
		{
			name: "config",
			args: []string{
				"--config", `{"openapi": "3.0"}`,
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Len(t, cfg.Configs, 1)
				require.Equal(t, "{\"openapi\": \"3.0\"}", cfg.Configs[0])
			},
		},
		{
			name: "positional parameter file",
			args: []string{
				"foo.json",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo.json"}, cfg.Args)
				// next check requires to run cfg.Parse
				require.Equal(t, "foo.json", cfg.Providers.File.Filenames[0])
			},
		},
		{
			name: "positional parameter http",
			args: []string{
				"http://foo.io/foo.json",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "http://foo.io/foo.json", cfg.Providers.Http.Urls[0])
			},
		},
		{
			name: "positional parameter https",
			args: []string{
				"https://foo.io/foo.json",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "https://foo.io/foo.json", cfg.Providers.Http.Urls[0])
			},
		},
		{
			name: "positional parameter git with https",
			args: []string{
				"git+https://foo.io/foo.json",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "https://foo.io/foo.json", cfg.Providers.Git.Urls[0])
			},
		},
		{
			name: "positional parameter npm",
			args: []string{
				"npm://bar/foo.txt?scope=@foo",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "npm://bar/foo.txt?scope=@foo", cfg.Providers.Npm.Packages[0].Name)
			},
		},
		{
			name: "positional parameter Windows path",
			args: []string{
				"C:\\bar",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "C:\\bar", cfg.Providers.File.Filenames[0])
			},
		},
		{
			name: "data-gen optional properties",
			args: []string{
				"--data-gen-optional-properties", "often",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, 0.85, cfg.DataGen.OptionalPropertiesProbability())
			},
		},
		{
			name: "data-gen optional properties always",
			args: []string{
				"--data-gen-optional-properties", "always",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, 1.0, cfg.DataGen.OptionalPropertiesProbability())
			},
		},
		{
			name: "data-gen optional properties sometimes",
			args: []string{
				"--data-gen-optional-properties", "sometimes",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, 0.5, cfg.DataGen.OptionalPropertiesProbability())
			},
		},
		{
			name: "data-gen optional properties 0.3",
			args: []string{
				"--data-gen-optional-properties", "0.3",
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, 0.3, cfg.DataGen.OptionalPropertiesProbability())
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cmd := mokapi.NewCmdMokapi(context.Background())
			cmd.SetArgs(tc.args)

			cfg := static.NewConfig()
			cmd.Run = func(cmd *cli.Command, args []string) error {
				cfg = cmd.Config.(*static.Config)
				cfg.Args = args
				return cfg.Parse()
			}
			err := cmd.Execute()
			require.NoError(t, err)

			tc.test(t, cfg)
		})
	}
}

func TestMokapi_Env(t *testing.T) {
	testcases := []struct {
		name string
		env  map[string]string
		test func(t *testing.T, cfg *static.Config, err error)
	}{
		{
			name: "env var",
			env: map[string]string{
				"MOKAPI_Providers_GIT_Repositories[0]_Url":           "https://github.com/PATH-TO/REPOSITORY",
				"MOKAPI_Providers_GIT_Repositories[0]_Pull_Interval": "3m",
			},
			test: func(t *testing.T, cfg *static.Config, err error) {
				require.NoError(t, err)
				require.Len(t, cfg.Providers.Git.Repositories, 1)
				require.Equal(t, "https://github.com/PATH-TO/REPOSITORY", cfg.Providers.Git.Repositories[0].Url)
				require.Equal(t, "3m", cfg.Providers.Git.Repositories[0].PullInterval)
			},
		},
		{
			name: "data-gen env var",
			env: map[string]string{
				"MOKAPI_DATA_GEN_OPTIONAL_PROPERTIES": "sometimes",
			},
			test: func(t *testing.T, cfg *static.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, 0.5, cfg.DataGen.OptionalPropertiesProbability())
			},
		},
		{
			name: "not supported env var",
			env: map[string]string{
				"MOKAPI_NOT_SUPPORTED": "foo",
			},
			test: func(t *testing.T, cfg *static.Config, err error) {
				require.EqualError(t, err, "unknown environment variable 'MOKAPI_NOT_SUPPORTED' (value 'foo')")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				for k := range tc.env {
					_ = os.Unsetenv(k)
				}
			}()
			for k, v := range tc.env {
				err := os.Setenv(k, v)
				require.NoError(t, err)
			}

			cmd := mokapi.NewCmdMokapi(context.Background())
			cmd.SetArgs([]string{})

			cfg := static.NewConfig()
			cmd.Run = func(cmd *cli.Command, args []string) error {
				cfg = cmd.Config.(*static.Config)
				return cfg.Parse()
			}
			err := cmd.Execute()
			tc.test(t, cfg, err)
		})
	}
}

func TestMokapi_File(t *testing.T) {
	newCmd := func(args []string) (*cli.Command, *static.Config) {
		c := mokapi.NewCmdMokapi(context.Background())
		c.SetArgs(args)
		c.Run = func(cmd *cli.Command, args []string) error {
			return nil
		}
		return c, c.Config.(*static.Config)
	}

	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "config parameter with file path as value",
			test: func(t *testing.T) {
				path := createTempFile(t, "test.json", `{"openapi": "3.0"}`)
				c, cfg := newCmd([]string{"--config", fmt.Sprintf("file:%s", path)})

				err := c.Execute()
				require.NoError(t, err)
				require.Len(t, cfg.Configs, 1)
				require.Equal(t, "{\"openapi\": \"3.0\"}", cfg.Configs[0])
			},
		},
		{
			name: "configfile json",
			test: func(t *testing.T) {
				path := createTempFile(t, "test.json", `{"configs": [ { "openapi": "3.0", "info": { "name": "foo" } } ]}`)
				c, cfg := newCmd([]string{"--config-file", path})
				err := c.Execute()
				require.NoError(t, err)

				require.Len(t, cfg.Configs, 1)

				actual := map[string]interface{}{}
				err = json.Unmarshal([]byte(cfg.Configs[0]), &actual)
				require.NoError(t, err)
				expected := map[string]interface{}{
					"openapi": "3.0",
					"info": map[string]interface{}{
						"name": "foo",
					},
				}
				require.Equal(t, expected, actual)
			},
		},
		{
			name: "config-file yaml",
			test: func(t *testing.T) {
				path := createTempFile(t, "foo.yaml", `
configs:
  - openapi: "3.0"
    info: 
      name: foo
`)
				c, cfg := newCmd([]string{"--config-file", path})
				err := c.Execute()
				require.NoError(t, err)

				actual := map[string]interface{}{}
				err = json.Unmarshal([]byte(cfg.Configs[0]), &actual)
				require.NoError(t, err)
				expected := map[string]interface{}{
					"openapi": "3.0",
					"info": map[string]interface{}{
						"name": "foo",
					},
				}

				require.Len(t, cfg.Configs, 1)
				require.Equal(t, expected, actual)
			},
		},
		{
			name: "config-file",
			test: func(t *testing.T) {
				path := createTempFile(t, "foo.json", `{"log": { "level": "error" } }`)
				c, cfg := newCmd([]string{"--config-file", path})
				err := c.Execute()
				require.NoError(t, err)

				require.Equal(t, "error", cfg.Log.Level)
			},
		},
		{
			name: "cli-input",
			test: func(t *testing.T) {
				path := createTempFile(t, "foo.json", `{"log": { "level": "error" } }`)
				c, cfg := newCmd([]string{"--cli-input", path})
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, "error", cfg.Log.Level)
			},
		},
		{
			name: "cli-input file provider directories",
			test: func(t *testing.T) {
				path := createTempFile(t, "foo.yaml", `
providers:
  file:
    directory: foo
`)
				c, cfg := newCmd([]string{"--cli-input", path})
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, []string{"foo"}, cfg.Providers.File.Directories)
			},
		},
		{
			name: "cli-input file provider directories",
			test: func(t *testing.T) {
				path := createTempFile(t, "foo.yaml", `
providers:
  file:
    directories: ["/foo", "/bar"]
`)
				c, cfg := newCmd([]string{"--cli-input", path})
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, []string{"/foo", "/bar"}, cfg.Providers.File.Directories)
			},
		},
		{
			name: "cli-input file provider directory",
			test: func(t *testing.T) {
				path := createTempFile(t, "foo.json", `{"providers":{"file":{"directory":"foo"}}}`)
				c, cfg := newCmd([]string{"--cli-input", path})
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, []string{"foo"}, cfg.Providers.File.Directories)
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

func TestPositionalArg_Error(t *testing.T) {
	cmd := mokapi.NewCmdMokapi(context.Background())
	cmd.SetArgs([]string{"foo://bar"})

	cfg := static.NewConfig()
	cmd.Run = func(cmd *cli.Command, args []string) error {
		cfg = cmd.Config.(*static.Config)
		cfg.Args = args
		return cfg.Parse()
	}
	err := cmd.Execute()
	require.EqualError(t, err, "positional argument is not supported: foo://bar")
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

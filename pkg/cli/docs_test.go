package cli_test

import (
	"errors"
	"fmt"
	"mokapi/pkg/cli"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type failingWriter struct{}

func (failingWriter) WriteString(string) (int, error) { return 0, errors.New("disk full") }

func TestDocs(t *testing.T) {
	testcases := []struct {
		name string
		args []string
		test func(t *testing.T)
	}{
		{
			name: "one flag",
			test: func(t *testing.T) {
				c := cli.Command{Name: "Foo", Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().String("foo", "", cli.FlagDoc{})
				sb := &strings.Builder{}
				err := c.WriteMarkdown(sb)
				require.NoError(t, err)
				require.Equal(t, `---
title: Foo CLI Flags
description: A complete list of all Foo flags, with descriptions, defaults, and examples of how to set the option in config file, environment variables, or CLI.
---

# Foo

<div class="flags">

## Flags

### General

| Flag | Description |
|------|-------------|
| --[foo](#foo) |  |
| --[help](#help) | Show help information and exit |

</div>

## <a id="foo" name="foo"></a>foo

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --foo | FOO | - | string | - |

## <a id="help" name="help"></a>help

Show help information and exit

| Flag | Shorthand | Env | Config file | Type | Default |
|------|:---------:|------|-------------|:----:|:-------:|
| --help | -h | HELP | - | bool | false |

`, sb.String())
			},
		},
		{
			name: "one flag with alias",
			test: func(t *testing.T) {
				c := cli.Command{Name: "Foo", Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().String("foo", "", cli.FlagDoc{})
				c.Flags().Alias("foo", "alias")
				sb := &strings.Builder{}
				err := c.WriteMarkdown(sb)
				require.NoError(t, err)
				require.Equal(t, `---
title: Foo CLI Flags
description: A complete list of all Foo flags, with descriptions, defaults, and examples of how to set the option in config file, environment variables, or CLI.
---

# Foo

<div class="flags">

## Flags

### General

| Flag | Description |
|------|-------------|
| --[foo](#foo) |  |
| --[help](#help) | Show help information and exit |

</div>

## <a id="foo" name="foo"></a>foo

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --foo | FOO | - | string | - |

### Aliases

- alias

## <a id="help" name="help"></a>help

Show help information and exit

| Flag | Shorthand | Env | Config file | Type | Default |
|------|:---------:|------|-------------|:----:|:-------:|
| --help | -h | HELP | - | bool | false |

`, sb.String())
			},
		},
		{
			name: "one flag with documentation",
			test: func(t *testing.T) {
				c := cli.Command{Name: "Foo", Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().String("foo", "", cli.FlagDoc{
					Short: "A short description of the foo flag",
					Long:  `Some long description here`,
					Examples: []cli.Example{
						{
							Codes: []cli.Code{
								{Title: "CLI", Source: "--foo bar"},
								{Title: "Env", Source: "FOO=bar"},
								{Title: "File", Source: "foo: bar"},
							},
						},
					},
				})
				sb := &strings.Builder{}
				err := c.WriteMarkdown(sb)
				require.NoError(t, err)
				require.Equal(t, fmt.Sprintf(`---
title: Foo CLI Flags
description: A complete list of all Foo flags, with descriptions, defaults, and examples of how to set the option in config file, environment variables, or CLI.
---

# Foo

<div class="flags">

## Flags

### General

| Flag | Description |
|------|-------------|
| --[foo](#foo) | A short description of the foo flag |
| --[help](#help) | Show help information and exit |

</div>

## <a id="foo" name="foo"></a>foo

Some long description here

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --foo | FOO | - | string | - |

%s

## <a id="help" name="help"></a>help

Show help information and exit

| Flag | Shorthand | Env | Config file | Type | Default |
|------|:---------:|------|-------------|:----:|:-------:|
| --help | -h | HELP | - | bool | false |

`, "```bash tab=CLI\n--foo bar\n```\n```bash tab=Env\nFOO=bar\n```\n```bash tab=File\nfoo: bar\n```"), sb.String())
			},
		},
		{
			name: "dynamic flag with index",
			test: func(t *testing.T) {
				c := cli.Command{Name: "Foo", Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().DynamicString("foo[<index>]-bar", cli.FlagDoc{Short: "Bar of foo"})
				sb := &strings.Builder{}
				err := c.WriteMarkdown(sb)
				require.NoError(t, err)
				require.Contains(t, sb.String(), "| --[foo[<index>]-bar](#foo-index-bar) | Bar of foo |\n")
				require.Contains(t, sb.String(), "## <a id=\"foo-index-bar\" name=\"foo-index-bar\"></a>foo[&lt;index&gt;]-bar\n")
			},
		},
		{
			name: "groups in order of first use, ungrouped flags last",
			test: func(t *testing.T) {
				c := cli.Command{Name: "Foo", Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().Group("Log")
				c.Flags().String("log-level", "info", cli.FlagDoc{Short: "Set log level"})
				c.Flags().Group("API")
				c.Flags().Int("api-port", 8080, cli.FlagDoc{Short: "Port"})
				c.Flags().Group("Log")
				c.Flags().String("log-format", "text", cli.FlagDoc{Short: "Set log format"})
				sb := &strings.Builder{}
				err := c.WriteMarkdown(sb)
				require.NoError(t, err)
				require.Contains(t, sb.String(), `## Flags

### Log

| Flag | Description |
|------|-------------|
| --[log-level](#log-level) | Set log level |
| --[log-format](#log-format) | Set log format |

### API

| Flag | Description |
|------|-------------|
| --[api-port](#api-port) | Port |

### General

| Flag | Description |
|------|-------------|
| --[help](#help) | Show help information and exit |

</div>
`)
			},
		},
		{
			name: "indexed variants are listed in the section of their parent",
			test: func(t *testing.T) {
				c := cli.Command{Name: "Foo", Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().StringSlice("repos", nil, false, cli.FlagDoc{Short: "Repositories"})
				c.Flags().DynamicString("repos[<index>]-url", cli.FlagDoc{Short: "Set the URL"})
				c.Flags().DynamicInt("repos-<name>-size", cli.FlagDoc{Short: "Set the size"})
				sb := &strings.Builder{}
				err := c.WriteMarkdown(sb)
				require.NoError(t, err)
				docs := sb.String()

				overview, details, _ := strings.Cut(docs, "</div>")
				require.Contains(t, overview, "| --[repos](#repos) | Repositories |\n")
				require.NotContains(t, overview, "<index>")
				require.NotContains(t, overview, "<name>")

				require.Contains(t, details, `### Sub-options

| Flag | Config file | Type | Description |
|------|-------------|:----:|-------------|
| <a id="repos-index-url" name="repos-index-url"></a>`+"`--repos[<index>]-url`"+` | - | string | Set the URL |
| <a id="repos-name-size" name="repos-name-size"></a>`+"`--repos-<name>-size`"+` | - | int | Set the size |
`)
				require.Less(t, strings.Index(details, `<a id="repos" name="repos"></a>repos`), strings.Index(details, "### Sub-options"))
			},
		},
		{
			name: "config file key",
			test: func(t *testing.T) {
				type config struct {
					Log struct{ Level string }
				}
				c := cli.Command{Name: "Foo", Config: &config{}, EnvPrefix: "FOO_", Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().String("log-level", "info", cli.FlagDoc{})
				sb := &strings.Builder{}
				err := c.WriteMarkdown(sb)
				require.NoError(t, err)
				require.Contains(t, sb.String(), `| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --log-level | FOO_LOG_LEVEL | log.level | string | info |
`)
				require.Contains(t, sb.String(), "| --help | -h | FOO_HELP | - | bool | false |\n")
			},
		},
		{
			name: "config file key with placeholder is escaped",
			test: func(t *testing.T) {
				type config struct {
					Store map[string]struct{ Size int }
				}
				c := cli.Command{Name: "Foo", Config: &config{}, Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().String("store", "", cli.FlagDoc{})
				c.Flags().DynamicInt("store-<name>-size", cli.FlagDoc{})
				sb := &strings.Builder{}
				err := c.WriteMarkdown(sb)
				require.NoError(t, err)
				require.Contains(t, sb.String(), "| store.&lt;name&gt;.size | int |")
			},
		},
		{
			name: "config file only option",
			test: func(t *testing.T) {
				c := cli.Command{Name: "Foo", Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().Group("Log")
				c.Flags().String("log-level", "info", cli.FlagDoc{Short: "Set log level"})
				c.ConfigFileOption("javascript.modules", cli.FlagDoc{
					Short: "Module aliases",
					Long:  "Defines module aliases.",
					Examples: []cli.Example{
						{
							Codes: []cli.Code{
								{Title: "File (YAML)", Language: "yaml", Source: "javascript:\n  modules:\n    aliases:\n      a: b"},
								{Title: "File (JSON)", Language: "json", Source: `{"javascript": {"modules": {"aliases": {"a": "b"}}}}`},
							},
						},
					},
				})
				sb := &strings.Builder{}
				err := c.WriteMarkdown(sb)
				require.NoError(t, err)
				docs := sb.String()

				overview, details, _ := strings.Cut(docs, "</div>")
				require.Contains(t, overview, `### Config file only

| Option | Description |
|--------|-------------|
| [javascript.modules](#javascript-modules) | Module aliases |
`)
				require.Contains(t, details, "## <a id=\"javascript-modules\" name=\"javascript-modules\"></a>javascript.modules\n\nDefines module aliases.\n\n"+
					"```yaml tab=File (YAML)\njavascript:\n  modules:\n    aliases:\n      a: b\n```\n"+
					"```json tab=File (JSON)\n{\"javascript\": {\"modules\": {\"aliases\": {\"a\": \"b\"}}}}\n```\n")
			},
		},
		{
			name: "child command",
			test: func(t *testing.T) {
				child := &cli.Command{
					Name:    "bar",
					Use:     "foo bar [flags] FILE",
					Short:   "Do bar",
					Long:    "Does bar.",
					Example: "  foo bar a.json",
					Run:     func(cmd *cli.Command, args []string) error { return nil },
				}
				child.Flags().IntShort("count", "n", 1, cli.FlagDoc{Short: "Number of bars"})
				c := cli.Command{Name: "Foo", Short: "Root", EnvPrefix: "FOO_", Commands: []*cli.Command{child}, Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				sb := &strings.Builder{}
				err := c.WriteMarkdown(sb)
				require.NoError(t, err)
				docs := sb.String()

				require.Contains(t, docs, "## Commands\n\n| Command | Description |\n|---------|-------------|\n| [bar](#bar) | Do bar |\n\n")

				_, section, found := strings.Cut(docs, "## <a id=\"bar\" name=\"bar\"></a>bar\n")
				require.True(t, found, "child section")
				require.True(t, strings.HasPrefix(section, "\nDo bar\n\n### Description\n\nDoes bar.\n\n### Usage\n\n"), section)
				require.Contains(t, section, "### Example\n\n```bash\nfoo bar a.json\n```\n\n")
				require.Contains(t, section, `<div class="flags">

### Flags

#### General

| Flag | Description |
|------|-------------|
| --[count](#bar-count) | Number of bars |
| --[help](#bar-help) | Show help information and exit |

</div>
`)
				require.Contains(t, section, "### <a id=\"bar-count\" name=\"bar-count\"></a>count\n\nNumber of bars\n\n| Flag | Shorthand | Env | Config file | Type | Default |\n|------|:---------:|------|-------------|:----:|:-------:|\n| --count | -n | FOO_COUNT | - | int | 1 |\n")
				// child comes after the flags of the root command
				require.Less(t, strings.Index(docs, `<a id="help" name="help"></a>`), strings.Index(docs, `<a id="bar"`))
			},
		},
		{
			name: "pipe in a description does not end the table cell",
			test: func(t *testing.T) {
				c := cli.Command{Name: "Foo", Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().String("level", "info", cli.FlagDoc{Short: "Set level (debug|info)"})
				sb := &strings.Builder{}
				require.NoError(t, c.WriteMarkdown(sb))
				require.Contains(t, sb.String(), "| --[level](#level) | Set level (debug\\|info) |\n")
			},
		},
		{
			name: "long description and examples of a variant are kept below the sub-options",
			test: func(t *testing.T) {
				c := cli.Command{Name: "Foo", Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().StringSlice("repos", nil, false, cli.FlagDoc{Short: "Repositories"})
				c.Flags().DynamicString("repos[<index>]-url", cli.FlagDoc{
					Short: "Set the URL",
					Long:  "Specifies the repository URL.",
					Examples: []cli.Example{
						{Codes: []cli.Code{{Title: "CLI", Source: "--repos[0]-url x"}}},
					},
				})
				c.Flags().DynamicString("repos[<index>]-name", cli.FlagDoc{Short: "Set the name"})
				sb := &strings.Builder{}
				require.NoError(t, c.WriteMarkdown(sb))
				_, details, _ := strings.Cut(sb.String(), "</div>")

				require.Contains(t, details, "| Set the URL |\n| <a id=\"repos-index-name\" name=\"repos-index-name\"></a>`--repos[<index>]-name` | - | string | Set the name |\n\n"+
					"#### `--repos[<index>]-url`\n\nSpecifies the repository URL.\n\n```bash tab=CLI\n--repos[0]-url x\n```\n\n")
				require.NotContains(t, details, "#### `--repos[<index>]-name`")
			},
		},
		{
			name: "example with title, description and code without title",
			test: func(t *testing.T) {
				c := cli.Command{Name: "Foo", Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().String("foo", "", cli.FlagDoc{Examples: []cli.Example{
					{Title: "Set foo", Description: "Foo is set to bar.", Codes: []cli.Code{{Source: "--foo bar"}}},
				}})
				sb := &strings.Builder{}
				require.NoError(t, c.WriteMarkdown(sb))
				require.Contains(t, sb.String(), "### Set foo\n\nFoo is set to bar.\n\n```bash\n--foo bar\n```\n\n")
			},
		},
		{
			name: "command without name",
			test: func(t *testing.T) {
				c := cli.Command{}
				require.EqualError(t, c.WriteMarkdown(&strings.Builder{}), "command must specify a name")
			},
		},
		{
			name: "writer fails",
			test: func(t *testing.T) {
				c := cli.Command{Name: "Foo"}
				require.EqualError(t, c.WriteMarkdown(failingWriter{}), "disk full")
			},
		},
		{
			name: "GenMarkdown writes the file named after the command",
			test: func(t *testing.T) {
				dir := t.TempDir()
				c := cli.Command{Name: "my cmd"}
				require.NoError(t, c.GenMarkdown(dir))
				b, err := os.ReadFile(filepath.Join(dir, "my_cmd.md"))
				require.NoError(t, err)
				require.Contains(t, string(b), "# my cmd\n")
				require.Error(t, c.GenMarkdown(filepath.Join(dir, "missing")))
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

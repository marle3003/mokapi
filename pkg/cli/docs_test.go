package cli_test

import (
	"fmt"
	"mokapi/pkg/cli"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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

| Name | Usage |
|------|-------|
| --[foo](#foo) |  |
| --[help](#help) | Show help information and exit |

</div>

## <a name=foo></a>foo

| Flag | Env  | Type | Default |
|------|------|:----:|:-------:|
| --foo | FOO | string | - |

## <a name=help></a>help

Show help information and exit

| Flag | Shorthand | Env  | Type | Default |
|------|:---------:|------|:----:|:-------:|
| --help | -h | HELP | bool | false |

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

| Name | Usage |
|------|-------|
| --[foo](#foo) |  |
| --[help](#help) | Show help information and exit |

</div>

## <a name=foo></a>foo

| Flag | Env  | Type | Default |
|------|------|:----:|:-------:|
| --foo | FOO | string | - |

### Aliases

- alias

## <a name=help></a>help

Show help information and exit

| Flag | Shorthand | Env  | Type | Default |
|------|:---------:|------|:----:|:-------:|
| --help | -h | HELP | bool | false |

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

| Name | Usage |
|------|-------|
| --[foo](#foo) | A short description of the foo flag |
| --[help](#help) | Show help information and exit |

</div>

## <a name=foo></a>foo

Some long description here

| Flag | Env  | Type | Default |
|------|------|:----:|:-------:|
| --foo | FOO | string | - |

%s

## <a name=help></a>help

Show help information and exit

| Flag | Shorthand | Env  | Type | Default |
|------|:---------:|------|:----:|:-------:|
| --help | -h | HELP | bool | false |

`, "```bash tab=CLI\n--foo bar\n```\n```bash tab=Env\nFOO=bar\n```\n```bash tab=File\nfoo: bar\n```"), sb.String())
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

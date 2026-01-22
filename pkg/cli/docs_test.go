package cli_test

import (
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
			name: "string value",
			test: func(t *testing.T) {
				c := cli.Command{Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().String("foo", "", cli.FlagDoc{})
				sb := &strings.Builder{}
				err := c.WriteMarkdown(sb)
				require.NoError(t, err)
				require.Equal(t, `---
title: Mokapi CLI Flags
description: A complete list of all Mokapi flags, with descriptions, defaults, and examples of how to set the option in config file, environment variables, or CLI.
---

# 



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

| Flag | Env  | Type | Default |
|------|------|:----:|:-------:|
| --help | HELP | bool | false |


`, sb.String())
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

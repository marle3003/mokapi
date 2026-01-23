package mokapi

import (
	"mokapi/pkg/cli"
)

func NewGenCliDocCmd() *cli.Command {
	cmd := &cli.Command{
		Name: "generator for CLI documentation",
		Use:  "gen-cli-doc [flags]",
		Run: func(cmd *cli.Command, args []string) error {
			c := NewCmdMokapi()
			return c.GenMarkdown(cmd.Flags().GetString("output-dir"))
		},
	}

	cmd.Flags().StringShort("output-dir", "-o", "./docs/configuration/static", cli.FlagDoc{})

	return cmd
}

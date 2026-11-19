package flags

import "mokapi/pkg/cli"

func RegisterGeneralFlags(cmd *cli.Command) {
	cmd.Flags().Group(cli.DefaultGroup)
	cmd.Flags().StringSlice("config", []string{}, true, config)
	cmd.Flags().StringSlice("configs", []string{}, false, configs)

	cmd.Flags().File("config-file", cliInput)
	cmd.Flags().Alias("config-file", "cli-input")

	cmd.Flags().String("generate-cli-skeleton", "", generateCliSkeleton)
}

var cliInput = cli.FlagDoc{
	Short: "Read configuration from a file",
	Long: `Reads configuration from a file provided via the command line.
The file content is processed as configuration input and merged with other configuration sources such as inline configuration and environment variables.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--config-file=mokapi.yaml"},
			},
		},
	},
}

var config = cli.FlagDoc{
	Short: "Provide inline configuration data",
	Long: `Provides inline configuration data directly via the command line.
This option can be used multiple times to supply additional configuration fragments, which are merged with other configuration sources.
To pass several fragments in one argument, use --configs instead.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--config {\"openapi\": \"3.0\", \"info\": { \"title\": \"foo\" } }\n--config file:/path/to/config.yaml"},
			},
		},
	},
}

var configs = cli.FlagDoc{
	Short: "Provide inline configuration data",
	Long: `Provides inline configuration data directly via the command line.
Several fragments are passed in a single argument, separated by spaces.
To add a single fragment per flag, use --config instead.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--configs file:/path/to/config.yaml  file:/path/to/config.yaml"},
			},
		},
	},
}

var generateCliSkeleton = cli.FlagDoc{
	Short: "Generate a configuration skeleton and exit",
	Long: `Prints a configuration skeleton and exits without starting Mokapi.
If set without a value, the full skeleton is generated. If a value is provided, only the specified section is generated, e.g. providers.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--generate-cli-skeleton\n--generate-cli-skeleton providers"},
			},
		},
	},
}

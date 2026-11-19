package flags

import "mokapi/pkg/cli"

func RegisterLogFlags(cmd *cli.Command) {
	cmd.Flags().Group("Log")
	cmd.Flags().String("log-level", "info", logLevel)
	cmd.Flags().String("log-format", "text", logFormat)
}

var logLevel = cli.FlagDoc{
	Short: "Set log level (debug|info|warn|error)",
	Long:  "The default level of log messages is info. You can set the log level to one of the following, listed in order of least to most information. The level is cumulative: for the debug level, the log file also includes messages at the info, warn, and error levels.\n- Debug\n- Info\n- Warn\n- Error\n",
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--log-level warn"},
				{Title: "Env", Source: "MOKAPI_LOG_LEVEL=warn"},
				{Title: "File", Source: "log:\n  level: warn", Language: "yaml"},
			},
		},
	},
}

var logFormat = cli.FlagDoc{
	Short: "Set log output format (text|json)",
	Long:  `Defines how Mokapi writes log messages. Use text for human-readable output or json to get one JSON object per line, which is easier to process with log collectors.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--log-format json"},
				{Title: "Env", Source: "MOKAPI_LOG_FORMAT=json"},
				{Title: "File", Source: "log:\n  format: json", Language: "yaml"},
			},
		},
	},
}

package flags

import "mokapi/pkg/cli"

func RegisterHealthFlags(cmd *cli.Command) {
	cmd.Flags().Bool("health-enabled", true, healthEnabled)
	cmd.Flags().Int("health-port", 8080, healthPort)
	cmd.Flags().String("health-path", "/health", healthPath)
	cmd.Flags().Bool("health-log", false, healthLog)
}

var healthEnabled = cli.FlagDoc{
	Short: "Enables or disables the health check endpoint entirely.",
	Long: `Enables or disables the health check endpoint entirely.
When set to false, Mokapi will not expose any health endpoint.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--health-enabled false"},
				{Title: "Env", Source: "MOKAPI_HEALTH_ENABLED=false"},
				{Title: "File", Source: "health:\n  enabled: false"},
			},
		},
	},
}

var healthPort = cli.FlagDoc{
	Short: "The port on which the health endpoint is exposed.",
	Long: `The port on which the health endpoint is exposed.
- If the value matches the dashboard/API port, the health endpoint is served by the same HTTP server.
- If a different port is specified, Mokapi starts a separate HTTP listener dedicated to health checks.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--health-port 8081"},
				{Title: "Env", Source: "MOKAPI_HEALTH_PORT=8081"},
				{Title: "File", Source: "health:\n  port: 8081"},
			},
		},
	},
}

var healthPath = cli.FlagDoc{
	Short: "The HTTP path for the health endpoint (default: /health).",
	Long: `The HTTP path for the health endpoint.
- If empty, the default path /health is used.
- The value must be an absolute path (starting with /).`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--health-path /health/live"},
				{Title: "Env", Source: "MOKAPI_HEALTH_PATH=/health/live"},
				{Title: "File", Source: "health:\n  path: /health/live"},
			},
		},
	},
}

var healthLog = cli.FlagDoc{
	Short: "Controls whether HTTP requests to the health endpoint are logged.",
	Long: `Controls whether HTTP requests to the health endpoint are logged.
By default, health check requests are not logged to avoid excessive log noise
from load balancers, uptime monitors, and orchestration systems.

Enable this option when debugging health check behavior.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--health-log"},
				{Title: "Env", Source: "MOKAPI_HEALTH_LOG=true"},
				{Title: "File", Source: "health:\n  log: true"},
			},
		},
	},
}

package flags

import "mokapi/pkg/cli"

func RegisterMcpFlags(cmd *cli.Command) {
	cmd.Flags().Bool("mcp-server-enabled", false, mcpServerEnabled)
	cmd.Flags().Int("mcp-server-port", 8080, mcpServerPort)
	cmd.Flags().String("mcp-server-path", "/mcp", mcpServerPath)
}

var mcpServerEnabled = cli.FlagDoc{
	Short: "Enable the MCP server",
	Long: `Enables the MCP (Model Context Protocol) server.
When enabled, Mokapi exposes an MCP-compatible endpoint that allows external tools and AI clients to interact with mocked APIs, events, and configuration through a standardized protocol.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--mcp-server-enabled"},
				{Title: "Env", Source: "MOKAPI_MCP_SERVER=true"},
				{Title: "File", Source: "mcp:\n  server:\n    enabled: true"},
			},
		},
	},
}

var mcpServerPort = cli.FlagDoc{
	Short: "Port for the MCP server",
	Long: `Specifies the TCP port on which the MCP server listens.
The MCP server provides access to Mokapi resources via the Model Context Protocol and can be consumed by compatible clients and tools.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--mcp-server-port 5000"},
				{Title: "Env", Source: "MOKAPI_MCP_PORT=5000"},
				{Title: "File", Source: "mcp:\n  server:\n    port: 5000"},
			},
		},
	},
}

var mcpServerPath = cli.FlagDoc{
	Short: "Path for the MCP server endpoint",
	Long: `Defines the HTTP path under which the MCP server is exposed.
This is useful when Mokapi is running behind a reverse proxy or when the MCP endpoint should be served under a specific path.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--mcp-server-path /foo/mcp"},
				{Title: "Env", Source: "MOKAPI_MCP_PATH=/foo/mcp"},
				{Title: "File", Source: "mcp:\n  server:\n    path: /foo/mcp"},
			},
		},
	},
}

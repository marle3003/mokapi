package flags

import "mokapi/pkg/cli"

func RegisterApiFlags(cmd *cli.Command) {
	cmd.Flags().Group("API & Dashboard")
	cmd.Flags().Int("api-port", 8080, apiPort)
	cmd.Flags().String("api-path", "", apiPath)
	cmd.Flags().String("api-base", "", apiBase)
	cmd.Flags().Bool("api-dashboard", true, apiDashboard)
	cmd.Flags().Bool("api-search-enabled", true, apiSearch)
	cmd.Flags().String("api-search-index-path", "", apiSearchIndexPath)
	cmd.Flags().Bool("api-search-in-memory", false, apiSearchInMemory)
}

var apiPort = cli.FlagDoc{
	Short: "Port for the API server",
	Long: `Specifies the TCP port on which the Mokapi API server listens.
The API server is the central entry point where developers can access all mocked services, events, and configuration data exposed by Mokapi. It also serves the web dashboard when enabled.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--api-port 5000"},
				{Title: "Env", Source: "MOKAPI_API_PORT=5000"},
				{Title: "File", Source: "api:\n  port: 5000", Language: "yaml"},
			},
		},
	},
}

var apiPath = cli.FlagDoc{
	Short: "Path prefix for the API and dashboard",
	Long: `Defines a path prefix under which the API and web dashboard are served.
This is useful when Mokapi is hosted behind a reverse proxy and needs to be accessible under a specific URL path.
See [Running the Dashboard Behind a Reverse Proxy](/docs/configuration/use-cases/reverse-proxy.md) for how it differs from [api-base](#api-base).`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--api-path /mokapi/dashboard"},
				{Title: "Env", Source: "MOKAPI_API_PATH=/mokapi/dashboard"},
				{Title: "File", Source: "api:\n  path: /mokapi/dashboard", Language: "yaml"},
			},
		},
	},
}

var apiBase = cli.FlagDoc{
	Short: "Base path used when the API is behind a reverse proxy",
	Long: `Specifies the external path under which the browser reaches the web dashboard when Mokapi runs behind a reverse proxy.
Mokapi sets this value as the base URL of the dashboard page, so the dashboard loads its assets and calls the API under this path.
It doesn't change the paths Mokapi itself serves (that's api-path). If api-base isn't set, the dashboard uses api-path instead.
See [Running the Dashboard Behind a Reverse Proxy](/docs/configuration/use-cases/reverse-proxy.md) for when you need it.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--api-base /mokapi/dashboard"},
				{Title: "Env", Source: "MOKAPI_API_BASE=/mokapi/dashboard"},
				{Title: "File", Source: "api:\n  base: /mokapi/dashboard", Language: "yaml"},
			},
		},
	},
}

var apiDashboard = cli.FlagDoc{
	Short: "Enable the web dashboard",
	Long: `Enables or disables the Mokapi web dashboard.
When disabled, the API server continues to run, but the dashboard UI is not exposed.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--api-dashboard true\n--api-dashboard\n--no-api-dashboard"},
				{Title: "Env", Source: "MOKAPI_API_DASHBOARD=false"},
				{Title: "File", Source: "api:\n  dashboard: false", Language: "yaml"},
			},
		},
	},
}

var apiSearch = cli.FlagDoc{
	Short: "Enable search functionality in the dashboard",
	Long: `Enables search functionality in the web dashboard.
When enabled, the dashboard shows the search panel (Search icon in the activity bar, or the / key). Users can search through mocked APIs, resources, configuration files and recorded events such as HTTP requests.
When disabled, Mokapi doesn't index your APIs and the dashboard hides the Search icon and the panel.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--api-search-enabled true\n--api-search-enabled\n--no-api-search-enabled"},
				{Title: "Env", Source: "MOKAPI_API_SEARCH_ENABLED=true"},
				{Title: "File", Source: "api:\n  search:\n    enabled: true", Language: "yaml"},
			},
		},
	},
}

var apiSearchIndexPath = cli.FlagDoc{
	Short: "Set storage location for the dashboard search index",
	Long: `Defines where Mokapi stores the search index used by the web dashboard.

By default, Mokapi stores the search index in a temporary directory on disk.
This reduces memory usage and provides predictable memory behavior.
Each Mokapi instance uses its own temporary directory, which is removed on shutdown.

If a file system path is specified, the index will be stored at that location during Mokapi's runtime.
The index is rebuilt at startup and does not block application startup.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{
					Title:  "CLI",
					Source: "--api-search-index-path /var/lib/mokapi/search",
				},
				{
					Title:  "Env",
					Source: "MOKAPI_API_SEARCH_INDEX_PATH=/var/lib/mokapi/search",
				},
				{
					Title: "File",
					Source: `api:
  search:
    indexPath: /var/lib/mokapi/search`,
					Language: "yaml",
				},
			},
		},
	},
}

var apiSearchInMemory = cli.FlagDoc{
	Short: "Store the dashboard search index entirely in memory",
	Long: `Forces Mokapi to keep the search index entirely in memory instead of using disk storage.

This provides fast indexing and search performance but increases Go heap memory usage.

This option is recommended only for small projects or development environments.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{
					Title:  "CLI",
					Source: "--api-search-in-memory",
				},
				{
					Title:  "Env",
					Source: "MOKAPI_API_SEARCH_IN_MEMORY=true",
				},
				{
					Title: "File",
					Source: `api:
  search:
    inMemory: true`,
					Language: "yaml",
				},
			},
		},
	},
}

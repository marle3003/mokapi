package flags

import "mokapi/pkg/cli"

func RegisterHttpProvider(cmd *cli.Command) {
	cmd.Flags().String("providers-http", "", providerHttp)
	cmd.Flags().StringSlice("providers-http-url", []string{}, true, providerHttpUrl)
	cmd.Flags().StringSlice("providers-http-urls", []string{}, false, providerHttpUrls)
	cmd.Flags().String("providers-http-poll-interval", "3m", providerHttpPollInterval)
	cmd.Flags().String("providers-http-poll-timeout", "5s", providerHttpPollTimeout)
	cmd.Flags().String("providers-http-proxy", "", providerHttpProxy)
	cmd.Flags().Bool("providers-http-tls-skip-verify", false, providerHttpTlsSkipVerify)
	cmd.Flags().String("providers-http-ca", "", providerHttpCa)
}

var providerHttp = cli.FlagDoc{
	Short: "Configure an HTTP-based provider using shorthand syntax",
	Long: `Enables the HTTP provider using a shorthand configuration.
When enabled, Mokapi fetches dynamic configuration from one or more HTTP endpoints. The provider periodically polls the configured URLs and applies updates automatically when the remote configuration changes.
Additional flags allow you to control polling behavior, timeouts, proxy settings, and TLS verification.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-http url=https://foo.bar/file.yaml,proxy=https://proxy.example.com"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_HTTP=url=https://foo.bar/file.yaml,proxy=https://proxy.example.com"},
				{Title: "File", Source: "providers:\n  http:\n    urls: https://foo.bar/file.yaml\n    proxy: https://proxy.example.com", Language: "yaml"},
			},
		},
	},
}

var providerHttpUrl = cli.FlagDoc{
	Short: "Fetch configuration from an HTTP endpoint",
	Long: `Specifies a single HTTP endpoint from which configuration is fetched.
This option can be used multiple times to define additional endpoints. Each endpoint is polled at the configured interval, and changes are applied dynamically.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-http-url https://foo.bar/file.yaml"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_HTTP_URL=https://foo.bar/file.yaml"},
				{Title: "File", Source: "providers:\n  http:\n    urls: https://foo.bar/file.yaml", Language: "yaml"},
			},
		},
	},
}

var providerHttpUrls = cli.FlagDoc{
	Short: "Fetch configurations from HTTP endpoints",
	Long: `Specifies multiple HTTP endpoints from which configuration is fetched.
This option is equivalent to using providers-http-url multiple times, but allows defining all endpoints in a single argument or configuration block.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-http-urls https://foo.bar/file1.yaml https://foo.bar/file2.yaml"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_HTTP_URLS=https://foo.bar/file1.yaml https://foo.bar/file2.yaml"},
				{Title: "File", Source: "providers:\n  http:\n    urls: [https://foo.bar/file1.yaml https://foo.bar/file2.yaml]", Language: "yaml"},
			},
		},
	},
}

var providerHttpPollInterval = cli.FlagDoc{
	Short: "Polling interval for HTTP endpoints",
	Long: `Defines how often the configured HTTP endpoints are polled for changes.
The value must be a valid duration string, such as "30s", "1m", or "5m". Shorter intervals result in faster updates but may increase network usage.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-http-poll-interval 10s"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_HTTP_POLL_INTERVAL=10s"},
				{Title: "File", Source: "providers:\n  http:\n    pollInterval: 10s", Language: "yaml"},
			},
		},
	},
}

var providerHttpPollTimeout = cli.FlagDoc{
	Short: "Timeout for HTTP polling requests",
	Long: `Sets the maximum duration allowed for a single HTTP polling request.
If the request does not complete within this time, it is aborted and retried at the next polling interval.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-http-poll-timeout 10s"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_HTTP_POLL_TIMEOUT=10s"},
				{Title: "File", Source: "providers:\n  http:\n    pollTimeout: 10s", Language: "yaml"},
			},
		},
	},
}

var providerHttpProxy = cli.FlagDoc{
	Short: "HTTP proxy URL",
	Long: `Configures an HTTP proxy used for all requests made by the HTTP provider.
This is useful in environments where outbound traffic must go through a proxy server.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-http-proxy http://localhost:3128"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_HTTP_PROXY=http://localhost:3128"},
				{Title: "File", Source: "providers:\n  http:\n    proxy: http://localhost:3128", Language: "yaml"},
			},
		},
	},
}

var providerHttpTlsSkipVerify = cli.FlagDoc{
	Short: "Skip TLS certificate verification",
	Long: `Disables TLS certificate verification for HTTPS endpoints.
This option can be useful in development environments or in enterprise setups where HTTPS traffic is intercepted by a proxy using a custom or internal certificate.
However, skipping certificate verification is insecure and should be avoided when possible. A safer alternative is to configure a custom certificate authority using providers-http-ca or to install the proxy certificate into the system certification pool.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-http-tls-skip-verify true\n--providers-http-tls-skip-verify"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_HTTP_TLS_SKIP_VERIFY=true"},
				{Title: "File", Source: "providers:\n  http:\n    tlsSkipVerify: true", Language: "yaml"},
			},
		},
	},
}

var providerHttpCa = cli.FlagDoc{
	Short: "Custom certificate authority file (default: system certification pool)",
	Long: `Specifies a custom certificate authority (CA) file used to verify TLS connections.
When set, the provided CA file is used in addition to, or instead of, the system certificate pool. This is useful when working with internal or self-signed certificates.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-http-ca /path/to/mycert.pem"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_HTTP_CA=/path/to/mycert.pem"},
				{Title: "File", Source: "providers:\n  http:\n    ca: /path/to/mycert.pem", Language: "yaml"},
			},
		},
	},
}

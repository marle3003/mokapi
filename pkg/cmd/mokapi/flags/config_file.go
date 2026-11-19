package flags

import "mokapi/pkg/cli"

// RegisterConfigFileOptions documents the options that have no flag and can only be set in the config file.
func RegisterConfigFileOptions(cmd *cli.Command) {
	cmd.ConfigFileOption("javascript.modules", javascriptModules)
	cmd.ConfigFileOption("certificates", certificates)
}

var javascriptModules = cli.FlagDoc{
	Short: "Module aliases for JavaScript imports",
	Long: `Config file only. Defines module aliases. Before Mokapi resolves an import, it replaces the import path with the target of the matching alias.
A key may contain one ` + "`*`" + ` wildcard; the matched text replaces the ` + "`*`" + ` in the target.
See [Module Aliases](/docs/javascript-api/modules.md#module-aliases) for details.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "File (YAML)", Language: "yaml", Source: `javascript:
  modules:
    aliases:
      msw: mokapi-msw-adapter
      "msw/*": mokapi-msw-adapter/*`},
				{Title: "File (JSON)", Language: "json", Source: `{
  "javascript": {
    "modules": {
      "aliases": {
        "msw": "mokapi-msw-adapter",
        "msw/*": "mokapi-msw-adapter/*"
      }
    }
  }
}`},
			},
		},
	},
}

var certificates = cli.FlagDoc{
	Short: "Static TLS certificates",
	Long: `Config file only. Defines static certificates that Mokapi serves for TLS connections instead of generating one with the root CA.
Each entry has a ` + "`cert`" + ` and a ` + "`key`" + `, given either as a file path or as the PEM content. A path starting with ` + "`./`" + ` is resolved from the directory of the config file.
A certificate is used for its common name and its DNS names. Other host names still get a generated certificate.
The ` + "`key`" + ` may be omitted if the ` + "`cert`" + ` file also contains the private key.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "File (YAML)", Language: "yaml", Source: `certificates:
  static:
    - cert: /path/to/cert.pem
      key: /path/to/key.pem`},
				{Title: "File (JSON)", Language: "json", Source: `{
  "certificates": {
    "static": [
      { "cert": "/path/to/cert.pem", "key": "/path/to/key.pem" }
    ]
  }
}`},
			},
		},
	},
}

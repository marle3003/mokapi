package flags

import "mokapi/pkg/cli"

func RegisterNpmProvider(cmd *cli.Command) {
	cmd.Flags().String("providers-npm", "", providerNpm)
	cmd.Flags().StringSlice("providers-npm-global-folder", []string{}, true, providerNpmGlobalFolder)
	cmd.Flags().StringSlice("providers-npm-global-folders", []string{}, false, providerNpmGlobalFolders)
	// package
	cmd.Flags().StringSlice("providers-npm-package", []string{}, true, providerNpmPackage)
	cmd.Flags().StringSlice("providers-npm-packages", []string{}, false, providerNpmPackages)
	cmd.Flags().DynamicString("providers-npm-packages[<index>]", providerNpmPackagesIndex)
	cmd.Flags().DynamicString("providers-npm-packages[<index>]-name", providerNpmPackagesIndexName)
	cmd.Flags().DynamicStringSlice("providers-npm-packages[<index>]-file", true, providerNpmPackagesIndexFile)
	cmd.Flags().DynamicStringSlice("providers-npm-packages[<index>]-files", false, providerNpmPackagesIndexFiles)
	cmd.Flags().DynamicStringSlice("providers-npm-packages[<index>]-include", false, providerNpmPackagesIndexInclude)
}

var providerNpm = cli.FlagDoc{
	Short: "Configure an npm-based provider using shorthand syntax",
	Long: `Enables the npm provider using a shorthand configuration.
When enabled, Mokapi loads configuration files from npm packages or from globally installed npm folders. This allows distributing configuration alongside npm packages and reusing it across projects.
Additional flags allow you to control which packages or files are loaded.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-npm package={\"name\": \"foo-api\"},globalFolders=/npm\n--providers-npm package={\\\"name\\\": \\\"foo-api\\\"},globalFolders=/npm # Windows\n"},
				{Title: "Env", Source: "MOKAPI__PROVIDERS_NPM=url=https://foo.bar/file.yaml,proxy=https://proxy.example.com"},
				{Title: "File", Source: "providers:\n  npm:\n    packages:\n     - name: foo-api\n    globalFolders: [/npm]", Language: "yaml"},
			},
		},
	},
}

var providerNpmGlobalFolder = cli.FlagDoc{
	Short: "Load configuration from a global npm folder",
	Long: `Adds additional folders where Mokapi looks for npm packages.
By default, Mokapi resolves npm packages by searching the current working directory and its parent directories, following the same resolution strategy as npm. This option allows you to extend that search path with one or more additional global npm folders.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-npm-global-folder /npm"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_NPM_GLOBAL_FOLDER=/npm"},
				{Title: "File", Source: "providers:\n  npm:\n    globalFolders: [/npm]", Language: "yaml"},
			},
		},
	},
}

var providerNpmGlobalFolders = cli.FlagDoc{
	Short: "Load configuration from a global npm folder",
	Long: `Specifies multiple global npm folders from which configuration files are loaded.
This option is equivalent to using providers-npm-global-folder multiple times.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-npm-global-folders /npm /npm2"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_NPM_GLOBAL_FOLDER=/npm /npm2"},
				{Title: "File", Source: "providers:\n  npm:\n    globalFolders: [/npm, /npm2]", Language: "yaml"},
			},
		},
	},
}

var providerNpmPackage = cli.FlagDoc{
	Short: "Configure an npm package using shorthand syntax",
	Long: `Configures a single npm package as a source for configuration files.
The package must be available locally or installed globally. Configuration files are loaded from the package contents.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-npm-package name=foo-api,file=api.yaml"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_NPM_PACKAGE=name=foo-api,file=api.yaml"},
				{Title: "File", Source: "providers:\n  npm:\n    packages:\n     - name: foo-api\n       files: [api.yaml]", Language: "yaml"},
			},
		},
	},
}

var providerNpmPackages = cli.FlagDoc{
	Short: "Configure npm packages using shorthand syntax",
	Long: `Configures multiple npm packages as sources for configuration files.
This option is equivalent to using providers-npm-package multiple times.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-npm-packages name=foo-api name=bar-api"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_NPM_PACKAGE=name=foo-api,file=api.yaml"},
				{Title: "File", Source: "providers:\n  npm:\n    packages:\n     - name: foo-api\n       files: [api.yaml]", Language: "yaml"},
			},
		},
	},
}

var providerNpmPackagesIndex = cli.FlagDoc{
	Short: "Configure the package at the specified index using shorthand syntax",
	Long: `Configures an npm package at the specified index in the packages list.
This option is useful when packages are defined via environment variables or configuration files, but need to be adjusted or overridden using CLI arguments. It allows precise, index-based modification of existing package definitions.
`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-npm-packages[0] file=api.yaml"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_NPM_PACKAGES[0]=file=api.yaml"},
			},
		},
	},
}

var providerNpmPackagesIndexName = cli.FlagDoc{
	Short: "Set the name of the npm package",
	Long: `Sets or overrides the name of the npm package at the specified index.
This is commonly used to adjust a package definition that was initially provided via environment variables or configuration files.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-npm-packages[0]-name bar-api"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_NPM_PACKAGES[0]-name=bar-api"},
			},
		},
	},
}

var providerNpmPackagesIndexFile = cli.FlagDoc{
	Short: "Allow only specific files from the package",
	Long: `Restricts configuration loading to specific files within the npm package at the specified index.
This option can be used multiple times and is useful for refining or overriding file selection defined via environment variables or configuration files.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-npm-packages[0]-file dist/model/openapi/complete-api.json"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_NPM_PACKAGES[0]_FILE=dist/model/openapi/complete-api.json"},
			},
		},
	},
}

var providerNpmPackagesIndexFiles = cli.FlagDoc{
	Short: "Allow only specific files from the package",
	Long: `Specifies multiple files within the npm package that are allowed to be loaded.
This option is equivalent to using providers-npm-packages[<index>]-file multiple times and can be used to override existing package configuration.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-npm-packages[0]-files api.yaml api2.yaml"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_NPM_PACKAGES[0]_FILES=api.yaml api2.yaml"},
			},
		},
	},
}

var providerNpmPackagesIndexInclude = cli.FlagDoc{
	Short: "Include only matching files or patterns from the package",
	Long: `Defines include patterns that files must match to be loaded from the npm package at the specified index.
This option is useful for narrowing or overriding file selection rules defined via environment variables or configuration files. When empty, all supported files are included by default.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-npm-packages[0]-include **/api/"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_NPM_PACKAGES[0]_FILES=**/api/"},
			},
		},
	},
}

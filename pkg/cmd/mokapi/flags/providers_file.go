package flags

import "mokapi/pkg/cli"

func RegisterFileProvider(cmd *cli.Command) {
	cmd.Flags().String("providers-file", "", providerFile)
	cmd.Flags().StringSlice("providers-file-filename", nil, true, providerFileFilename)
	cmd.Flags().StringSlice("providers-file-filenames", nil, false, providerFileFilenames)
	cmd.Flags().StringSlice("providers-file-directory", []string{}, true, providerFileDirectory)
	cmd.Flags().StringSlice("providers-file-directories", []string{}, false, providerFileDirectories)
	cmd.Flags().StringSlice("providers-file-skip-prefix", []string{"_"}, false, providerFileSkipPrefix)
	cmd.Flags().StringSlice("providers-file-include", []string{}, false, providerFileInclude)
	cmd.Flags().DynamicString("providers-file-include[<index>]", providerFileIncludeIndex)
	cmd.Flags().StringSlice("providers-file-exclude", []string{}, false, providerFileExclude)
	cmd.Flags().DynamicString("providers-file-exclude[<index>]", providerFileExcludeIndex)

	// directories
	cmd.Flags().DynamicString("providers-file-directories[<index>]", providerFileDirectoriesIndex)
	cmd.Flags().DynamicString("providers-file-directories[<index>]-path", providerFileDirectoriesPath)
	cmd.Flags().DynamicString("providers-file-directories[<index>]-include", providerFileDirectoriesInclude)
	cmd.Flags().DynamicString("providers-file-directories[<index>]-exclude", providerFileDirectoriesExclude)
}

var providerFile = cli.FlagDoc{
	Short: "Configure a File-based provider using shorthand syntax",
	Long: `Enables the file provider using a shorthand configuration. This option is useful for quick setups without defining a full configuration block.
When set, Mokapi loads dynamic configuration files from the specified path. You can further control how files are loaded using additional providers-file-* flags such as include rules, directories, and skip prefixes.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-file directory=/foo,fileSkipPrefix=skip"},
				{Title: "Env", Source: "MOKAPI_FILE=directory=/foo,fileSkipPrefix=skip"},
				{Title: "File", Source: "providers:\n  file:\n    directory: /foo\n    fileSkipPrefix: [skip]", Language: "yaml"},
			},
		},
	},
}

var providerFileFilename = cli.FlagDoc{
	Short: "Load dynamic configuration from a file",
	Long: `Specifies a single configuration file to load using the file provider.
This option can be used multiple times to load additional files.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-file-filename foobar.yaml"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_FILE_FILENAME=foobar.yaml"},
				{Title: "File", Source: "providers:\n  file:\n    filenames: [foobar.yaml]", Language: "yaml"},
			},
		},
	},
}

var providerFileFilenames = cli.FlagDoc{
	Short: "Load dynamic configuration from a file",
	Long: `Specifies multiple configuration files to load using the file provider.
This option is equivalent to using providers-file-filename multiple times, but allows defining all files in a single argument`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-file-filenames foo.yaml bar.yaml"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_FILE_FILENAMES=foo.yaml bar.yaml"},
				{Title: "File", Source: "providers:\n  file:\n    filenames: [foo.yaml, bar.yaml]", Language: "yaml"},
			},
		},
	},
}

var providerFileDirectory = cli.FlagDoc{
	Short: "Load the dynamic configuration from directories",
	Long: `Specifies a directory from which configuration files are loaded.
All supported configuration files in the directory are processed. The directory is watched for changes, allowing dynamic reloading when files are added, modified, or removed.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-file-directory ./configs"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_FILE_DIRECTORY=./configs"},
				{Title: "File", Source: "providers:\n  file:\n    directories: [./configs]", Language: "yaml"},
			},
		},
	},
}

var providerFileDirectories = cli.FlagDoc{
	Short: "Load the dynamic configuration from directories",
	Long: `Specifies multiple directories from which configuration files are loaded.
All supported configuration files in each directory are processed. Directories are watched for changes, allowing dynamic reloading across all configured paths.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-file-directories ./configs ./data"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_FILE_DIRECTORIES=./configs ./data"},
				{Title: "File", Source: "providers:\n  file:\n    directories: [./configs, ./data]", Language: "yaml"},
			},
		},
	},
}

var providerFileSkipPrefix = cli.FlagDoc{
	Short: "One or more prefixes that indicate whether a file or directory should be skipped.",
	Long: `Defines prefixes that cause files or directories to be ignored by the file provider.
Any file or directory whose name starts with one of the configured prefixes is skipped. This is useful for ignoring temporary files, backups, or disabled configurations.
By default, files and directories starting with "_" are skipped.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-file-skip-prefix skip_"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_FILE_SKIP_PREFIX=skip_"},
				{Title: "File", Source: "providers:\n  file:\n    skipPrefix: [skip_]", Language: "yaml"},
			},
		},
	},
}

var providerFileInclude = cli.FlagDoc{
	Short: "One or more patterns that a file must match, except when empty",
	Long: `Defines include patterns that configuration files must match to be loaded.
If at least one include pattern is specified, only files matching one of the patterns are processed. When empty, all supported files are included by default.
Patterns typically follow glob-style matching.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-file-include *.json *.yaml"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_FILE_INCLUDE=*.json *.yaml"},
				{Title: "File", Source: "providers:\n  file:\n    include: ['*.json', '*.yaml']", Language: "yaml"},
			},
		},
	},
}

var providerFileIncludeIndex = cli.FlagDoc{
	Short: "Set include rule at the specified index",
	Long: `Sets or overrides an include rule at a specific index in the include list.
This option is mainly intended for advanced use cases or programmatic configuration where precise control over individual include rules is required.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-file-include[1] *.yaml"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_FILE_INCLUDE[1]=*.yaml"},
			},
		},
	},
}

var providerFileExclude = cli.FlagDoc{
	Short: "Exclude files or directories matching patterns",
	Long: `Defines patterns for files or directories that should be excluded when loading configuration.
Any file or directory matching one of the exclude patterns is ignored, even if it would otherwise be included. When empty, no files are explicitly excluded.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-file-exclude tmp/*"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_FILE_EXCLUDE=tmp/*"},
				{Title: "File", Source: "providers:\n  file:\n    exclude: ['tmp/*']", Language: "yaml"},
			},
		},
	},
}

var providerFileExcludeIndex = cli.FlagDoc{
	Short: "Set exclude rule at the specified index",
	Long: `Sets or overrides an exclude rule at the specified index.
This option is useful when exclude rules are defined via environment variables or configuration files but need to be adjusted or overridden using CLI arguments.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-file-exclude[1] tmp/*"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_FILE_EXCLUDE[1]=tmp/*"},
			},
		},
	},
}

var providerFileDirectoriesIndex = cli.FlagDoc{
	Short: "Configure the directory at the specified index",
	Long: `Configures a directory entry at the specified index in the directories list.
This option allows directory-specific configuration and is mainly intended for advanced or programmatic use cases where precise control over individual directories is required.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-file-directories[0]-path ./mocks"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_FILE_DIRECTORIES[0]_PATH=./mocks"},
			},
		},
	},
}

var providerFileDirectoriesPath = cli.FlagDoc{
	Short: "Set the directory path",
	Long: `Specifies the filesystem path of the directory to load configuration files from.
All supported configuration files in this directory are processed, subject to include and exclude rules.
`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-file-directories[1]-path ./mocks"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_FILE_DIRECTORIES[1]_PATH=./mocks"},
			},
		},
	},
}

var providerFileDirectoriesInclude = cli.FlagDoc{
	Short: "Include only matching files or patterns",
	Long: `Defines include patterns that files in the directory must match to be loaded.
If at least one include pattern is specified, only files matching one of the patterns are processed. When empty, all supported files are included by default.
`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-file-directories[1]-include *index.js"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_FILE_DIRECTORIES[1]_INCLUDE=*index.js"},
			},
		},
	},
}

var providerFileDirectoriesExclude = cli.FlagDoc{
	Short: "Exclude matching files or patterns",
	Long: `Defines exclude patterns for files or directories within the specified directory.
Any file or directory matching one of the exclude patterns is ignored, even if it matches an include rule.
`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-file-exclude[1] *.tmp"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_FILE_EXCLUDE[1]=*.tmp"},
			},
		},
	},
}

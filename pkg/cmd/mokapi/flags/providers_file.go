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

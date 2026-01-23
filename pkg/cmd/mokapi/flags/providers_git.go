package flags

import "mokapi/pkg/cli"

func RegisterGitProvider(cmd *cli.Command) {
	cmd.Flags().String("providers-git", "", providerGit)
	cmd.Flags().StringSlice("providers-git-url", []string{}, true, providerGitUrl)
	cmd.Flags().StringSlice("providers-git-urls", []string{}, false, providerGitUrls)
	cmd.Flags().String("providers-git-pull-interval", "3m", providerGitPullInterval)
	cmd.Flags().String("providers-git-temp-dir", "", providerGitTempDir)
	cmd.Flags().StringSlice("providers-git-repository", []string{}, true, providerGitRepository)
	cmd.Flags().StringSlice("providers-git-repositories", []string{}, false, providerGitRepositories)

	// repository
	cmd.Flags().DynamicString("providers-git-repositories[<index>]", providerGitRepositoriesIndex)
	cmd.Flags().DynamicString("providers-git-repositories[<index>]-url", providerGitRepositoriesUrl)
	cmd.Flags().DynamicStringSlice("providers-git-repositories[<index>]-file", true, providerGitRepositoriesFile)
	cmd.Flags().DynamicStringSlice("providers-git-repositories[<index>]-files", false, providerGitRepositoriesFiles)
	cmd.Flags().DynamicStringSlice("providers-git-repositories[<index>]-include", false, providerGitRepositoriesInclude)
	cmd.Flags().DynamicString("providers-git-repositories[<index>]-auth-github", providerGitRepositoriesAuthGitHub)
	cmd.Flags().DynamicString("providers-git-repositories[<index>]-pull-interval", cli.FlagDoc{Short: "Override pull interval for this repository"})
}

var providerGit = cli.FlagDoc{
	Short: "Configure a Git-based provider using shorthand syntax",
	Long: `Enables the Git provider using a shorthand configuration.
When enabled, Mokapi clones one or more Git repositories and loads configuration files from them. Repositories are periodically pulled to detect changes and apply updates dynamically.
Additional flags allow you to control polling behavior, repository selection, authentication, and file filtering.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-git pullInterval=10s,tempDir=/tempdir"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_GIT=pullInterval=10s,tempDir=/tempdir"},
				{Title: "File", Source: "providers:\n  git:\n    pullInterval: 10s\n    tempDir: /tempdir", Language: "yaml"},
			},
		},
	},
}

var providerGitUrl = cli.FlagDoc{
	Short: "Clone configuration from a Git repository",
	Long: `Specifies a single Git repository from which configuration is cloned.
This option can be used multiple times to define additional repositories. Each repository is cloned and periodically pulled to fetch updates.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-git-url https://github.com/foo/foo.git"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_GIT_URL=https://github.com/foo/foo.git"},
				{Title: "File", Source: "providers:\n  git:\n    urls: https://github.com/foo/foo.git", Language: "yaml"},
			},
		},
	},
}

var providerGitUrls = cli.FlagDoc{
	Short: "Clone configuration from Git repositories",
	Long: `Specifies multiple Git repositories from which configuration is cloned.
This option is equivalent to using providers-git-url multiple times, but allows defining all repositories in a single argument or configuration block.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-git-urls https://github.com/foo/foo.git https://github.com/bar/bar.git"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_GIT_URLS=https://github.com/foo/foo.git https://github.com/bar/bar.git"},
				{Title: "File", Source: "providers:\n  git:\n    urls: [https://github.com/foo/foo.git https://github.com/bar/bar.git]", Language: "yaml"},
			},
		},
	},
}

var providerGitPullInterval = cli.FlagDoc{
	Short: "Interval for pulling updates from Git repositories",
	Long: `Defines how often Git repositories are pulled to check for updates.
The value must be a valid duration string, such as "30s", "1m", or "5m". Shorter intervals result in faster updates but may increase network and Git server load.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-git-pull-interval 10s"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_GIT_PULL_INTERVAL=10s"},
				{Title: "File", Source: "providers:\n  git:\n    pullInterval: 10s"},
			},
		},
	},
}

var providerGitTempDir = cli.FlagDoc{
	Short: "Temporary directory used for Git checkouts",
	Long: `Specifies the directory used for cloning and checking out Git repositories.
If not set, Mokapi uses a default temporary directory. Setting this option can be useful for controlling disk usage or for environments with restricted filesystem access.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-git-temp-dir /tempdir"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_GIT_TEMP_DIR=/tempdir"},
				{Title: "File", Source: "providers:\n  git:\n    tempDir: /tempdir", Language: "yaml"},
			},
		},
	},
}

var providerGitRepository = cli.FlagDoc{
	Short: "Configure a Git repository using shorthand syntax",
	Long: `Configures a single Git repository using a shorthand syntax.
This option allows defining repository-specific settings such as allowed files, authentication, and pull intervals.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-git-repository url=https://github.com/foo/foo.git,include=*.json"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_GIT_REPOSITORY=url=https://github.com/foo/foo.git,tempDir=/tempdir"},
				{Title: "File", Source: "providers:\n  git:\n    repositories:\n     - url: https://github.com/foo/foo.git\n       include: '*.json'", Language: "yaml"},
			},
		},
	},
}

var providerGitRepositories = cli.FlagDoc{
	Short: "Configure Git repositories using shorthand syntax",
	Long: `Configures multiple Git repositories using shorthand syntax.
This option is equivalent to using providers-git-repository multiple times, but allows defining all repositories in a single argument or configuration block.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-git-repositories url=https://github.com/foo/foo.git,include=*.json url=https://github.com/bar/bar.git,include=*.yaml"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_GIT_REPOSITORIES=url=https://github.com/foo/foo.git,include=*.json url=https://github.com/bar/bar.git,include=*.yaml"},
				{Title: "File", Source: "providers:\n  git:\n    repositories:\n     - url: https://github.com/foo/foo.git\n       include: ['*.json']\n     - url: https://github.com/bar/bar.git\n       include: ['*.yaml']", Language: "yaml"},
			},
		},
	},
}

var providerGitRepositoriesIndex = cli.FlagDoc{
	Short: "Configure the repository at the specified index using shorthand syntax",
	Long: `Configures a Git repository at the specified index in the repositories list.
This option is mainly intended for advanced or programmatic configurations where precise control over individual repositories is required.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-git-repositories[0] url=https://github.com/foo/foo.git,include=*.json"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]=url=https://github.com/foo/foo.git,include=*.json"},
				{Title: "File", Source: "providers:\n  git:\n    repositories:\n     - url: https://github.com/foo/foo.git\n       include: ['*.json']", Language: "yaml"},
			},
		},
	},
}

var providerGitRepositoriesUrl = cli.FlagDoc{
	Short: "Set the repository URL",
	Long: `Specifies the Git repository URL.
The repository is cloned and used as a source for configuration files.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-git-repositories[0]-url https://github.com/foo/foo.git"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_URL=https://github.com/foo/foo.git"},
				{Title: "File", Source: "providers:\n  git:\n    repositories:\n     - url: https://github.com/foo/foo.git", Language: "yaml"},
			},
		},
	},
}

var providerGitRepositoriesFile = cli.FlagDoc{
	Short: "Allow only specific files from the repository",
	Long: `Restricts configuration loading to specific files within the repository.
Only the specified files are considered. This option can be used multiple times to allow additional files.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-git-repositories[0]-file mokapi/api.json"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_FILE=mokapi/api.json"},
				{Title: "File", Source: "providers:\n  git:\n    repositories:\n     - files: [mokapi/api.json]", Language: "yaml"},
			},
		},
	},
}

var providerGitRepositoriesFiles = cli.FlagDoc{
	Short: "Allow only specific files from the repository",
	Long: `Specifies multiple files within the repository that are allowed to be loaded.
This option is equivalent to using providers-git-repositories[<index>]-file multiple times.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-git-repositories[0]-files mokapi/api.json mokapi/handler.js"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_FILES=mokapi/api.json mokapi/handler.js"},
				{Title: "File", Source: "providers:\n  git:\n    repositories:\n     - files: [mokapi/api.json mokapi/handler.js]", Language: "yaml"},
			},
		},
	},
}

var providerGitRepositoriesInclude = cli.FlagDoc{
	Short: "Include only matching files or patterns",
	Long: `Defines include patterns that files must match to be loaded from the repository.
If at least one include pattern is specified, only files matching one of the patterns are processed. When empty, all supported files are included by default.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-git-repositories[0]-include mokapi/**/*.json"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_INCLUDE=mokapi/**/*.json"},
				{Title: "File", Source: "providers:\n  git:\n    repositories:\n     - includes: ['mokapi/**/*.json']", Language: "yaml"},
			},
		},
	},
}

var providerGitRepositoriesAuthGitHub = cli.FlagDoc{
	Short: "Authenticate using GitHub credentials",
	Long: `Enables authentication using GitHub credentials for the repository.
This option allows accessing private repositories hosted on GitHub by using credentials provided via environment variables or the GitHub CLI configuration.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--providers-git-repositories[0]-auth-github appId=12345,installationId=123456789,privateKey=2024-2-25.private-key.pem"},
				{Title: "Env", Source: "MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_AUTH_GITHUB=appId=12345,installationId=123456789,privateKey=2024-2-25.private-key.pem"},
				{Title: "File", Source: "providers:\n  git:\n    repositories:\n     - auth:\n       github:\n         appId: 12345\n         installationId: 12345\n         privateKey: 2024-2-25.private-key.pem", Language: "yaml"},
			},
		},
	},
}

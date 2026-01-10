package mokapi

import (
	"context"
	"fmt"
	stdlog "log"
	"mokapi/api"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/mail"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/feature"
	"mokapi/pkg/cli"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/directory"
	mail2 "mokapi/providers/mail"
	"mokapi/providers/openapi"
	"mokapi/providers/swagger"
	"mokapi/runtime"
	"mokapi/safe"
	"mokapi/schema/json/generator"
	"mokapi/server"
	"mokapi/server/cert"
	"mokapi/version"
	"strings"

	log "github.com/sirupsen/logrus"
)

const logo = "888b     d888          888             d8888          d8b \n8888b   d8888          888            d88888          Y8P \n88888b.d88888          888           d88P888              \n888Y88888P888  .d88b.  888  888     d88P 888 88888b.  888 \n888 Y888P 888 d88\"\"88b 888 .88P    d88P  888 888 \"88b 888 \n888  Y8P  888 888  888 888888K    d88P   888 888  888 888 \n888   \"   888 Y88..88P 888 \"88b  d8888888888 888 d88P 888 \n888       888  \"Y88P\"  888  888 d88P     888 88888P\"  888 \n        v%s by Marcel Lehmann%s 888          \n        https://mokapi.io                    888          \n                                             888   \n"

func NewCmdMokapi() *cli.Command {
	cfg := static.NewConfig()

	cmd := &cli.Command{
		Name:    "mokapi",
		Use:     "mokapi [flags] [CONFIG-URL|DIRECTORY|FILE]...",
		Short:   "Start Mokapi and serve mocked APIs",
		Long:    `Mokapi is an easy, modern and flexible API mocking tool using Go and Javascript.`,
		Config:  cfg,
		Version: version.BuildVersion,
		Run: func(cmd *cli.Command, args []string) error {
			cfg := cmd.Config.(*static.Config)
			if err := applyPositionalArgs(cfg, args); err != nil {
				return err
			}
			return runRoot(cmd, cfg)
		},
		Commands: []*cli.Command{
			NewCmdSampleData(),
		},
		EnvPrefix: "MOKAPI_",
	}

	cmd.SetConfigPath(".", "/etc/mokapi")

	// file provider
	cmd.Flags().String("providers-file", "", "File-based provider using shorthand syntax: `filename=FILE,directory=DIR`")
	cmd.Flags().StringSlice("providers-file-filename", nil, "Load dynamic configuration from a file", true)
	cmd.Flags().StringSlice("providers-file-filenames", nil, "Load the dynamic configuration from files", false)
	cmd.Flags().StringSlice("providers-file-directory", []string{}, "Load the dynamic configuration from directories", true)
	cmd.Flags().StringSlice("providers-file-directories", []string{}, "Load the dynamic configuration from directories", false)
	cmd.Flags().StringSlice("providers-file-skip-prefix", []string{"_"}, "One or more prefixes that indicate whether a file or directory should be skipped.", false)
	cmd.Flags().StringSlice("providers-file-include", []string{}, "One or more patterns that a file must match, except when empty", false)
	cmd.Flags().DynamicString("providers-file-include[<index>]", "Set include rule at the specified index")

	// git provider
	cmd.Flags().String("providers-git", "", "Configure a Git-based provider using shorthand syntax")
	cmd.Flags().StringSlice("providers-git-url", []string{}, "Clone configuration from a Git repository", true)
	cmd.Flags().StringSlice("providers-git-urls", []string{}, "Clone configuration from a Git repository", false)
	cmd.Flags().String("providers-git-pull-interval", "3m", "Interval for pulling updates from Git repositories")
	cmd.Flags().String("providers-git-temp-dir", "", "Temporary directory used for Git checkouts")
	cmd.Flags().StringSlice("providers-git-repository", []string{}, "Configure a Git repository using shorthand syntax", true)
	cmd.Flags().StringSlice("providers-git-repositories", []string{}, "Configure a Git repository using shorthand syntax", false)
	// git repository
	cmd.Flags().DynamicString("providers-git-repositories[<index>]", "Configure the repository at the specified index using shorthand syntax")
	cmd.Flags().DynamicString("providers-git-repositories[<index>]-url", "Set the repository URL")
	cmd.Flags().DynamicStringSlice("providers-git-repositories[<index>]-file", "Allow only specific files from the repository", true)
	cmd.Flags().DynamicStringSlice("providers-git-repositories[<index>]-files", "Allow only specific files from the repository", false)
	cmd.Flags().DynamicStringSlice("providers-git-repositories[<index>]-include", "Include only matching files or patterns", false)
	cmd.Flags().DynamicString("providers-git-repositories[<index>]-auth-github", "Authenticate using GitHub credentials")
	cmd.Flags().DynamicString("providers-git-repositories[<index>]-pull-interval", "Override pull interval for this repository")

	// http provider
	cmd.Flags().String("providers-http", "", "Configure an HTTP-based provider using shorthand syntax")
	cmd.Flags().StringSlice("providers-http-url", []string{}, "Fetch configuration from an HTTP endpoint", true)
	cmd.Flags().StringSlice("providers-http-urls", []string{}, "Fetch configurations from an HTTP endpoints", false)
	cmd.Flags().String("providers-http-poll-interval", "3m", "Polling interval for HTTP endpoints")
	cmd.Flags().String("providers-http-poll-timeout", "5s", "Timeout for HTTP polling requests")
	cmd.Flags().String("providers-http-proxy", "", "HTTP proxy URL")
	cmd.Flags().Bool("providers-http-tls-skip-verify", false, "Skip TLS certificate verification")
	cmd.Flags().String("providers-http-ca", "", "Custom certificate authority file (default: system certification pool).")

	// npm provider
	cmd.Flags().String("providers-npm", "", "Configure an npm-based provider using shorthand syntax")
	cmd.Flags().StringSlice("providers-npm-global-folder", []string{}, "Load configuration from a global npm folder", true)
	cmd.Flags().StringSlice("providers-npm-global-folders", []string{}, "Load configuration from a global npm folder", false)
	// npm package
	cmd.Flags().StringSlice("providers-npm-package", []string{}, "Configure an npm package using shorthand syntax", true)
	cmd.Flags().StringSlice("providers-npm-packages", []string{}, "Configure an npm package using shorthand syntax", false)
	cmd.Flags().DynamicString("providers-npm-packages[<index>]", "Configure the package at the specified index using shorthand syntax")
	cmd.Flags().DynamicString("providers-npm-packages[<index>]-name", "Set the name of the npm package")
	cmd.Flags().DynamicStringSlice("providers-npm-packages[<index>]-file", "Allow only specific files from the package", true)
	cmd.Flags().DynamicStringSlice("providers-npm-packages[<index>]-files", "Allow only specific files from the package", false)
	cmd.Flags().DynamicStringSlice("providers-npm-packages[<index>]-include", "Include only matching files or patterns from the package", false)

	// API
	cmd.Flags().Int("api-port", 8080, "Port for the API server")
	cmd.Flags().String("api-path", "", "Path prefix for the API and dashboard")
	cmd.Flags().String("api-base", "", "Base path used when the API is behind a reverse proxy")
	cmd.Flags().Bool("api-dashboard", true, "Enable the web dashboard")
	cmd.Flags().Bool("api-search-enabled", false, "Enable search functionality in the dashboard")

	cmd.Flags().String("root-ca-cert", "", "Root CA certificate used for signing generated certificates")
	cmd.Flags().String("root-ca-key", "", "Private key of the root CA")

	cmd.Flags().Int("event-store-default-size", 100, "Default maximum number of stored events per API")
	cmd.Flags().String("event-store", "", "Configure event store using shorthand syntax")
	cmd.Flags().DynamicInt("event-store-<name>-size", "Override event store size for a specific API")

	cmd.Flags().String("data-gen-optional-properties", "0.85", "")

	cmd.Flags().StringSlice("config", []string{}, "Provide inline configuration data", true)
	cmd.Flags().StringSlice("configs", []string{}, "Provide inline configuration data", false)

	// config file
	cmd.Flags().File("config-file", "Read configuration from a file")
	cmd.Flags().Alias("config-file", "cli-input")

	// logging
	cmd.Flags().String("log-level", "info", "Set log level (debug|info|warn|error)").WithExample(
		cli.Example{
			Codes: []cli.Code{
				{Title: "Cli", Source: "--log-level=warn"},
				{Title: "Env", Source: "MOKAPI_LOG_LEVEL=warn"},
				{Title: "File", Source: "log:\n  level: warn"},
			},
		},
	).WithDescription("The default level of log messages is info. You can set the log level to one of the following, listed in order of least to most information. The level is cumulative: for the debug level, the log file also includes messages at the info, warn, and error levels.\n- Debug\n- Info\n- Warn\n- Error\n")
	cmd.Flags().String("log-format", "text", "Set log output format (text|json)")

	cmd.Flags().String("generate-cli-skeleton", "", "Generate a configuration skeleton and exit. If set without a value, generates the full skeleton. "+
		"If a value is provided, generates only the specified section (e.g. `providers`).")

	return cmd
}

func runRoot(cmd *cli.Command, cfg *static.Config) error {
	versionString := version.BuildVersion

	if f := cmd.Flags().Lookup("generate-cli-skeleton"); f != nil && f.Value.IsSet() {
		writeSkeleton(f.Value.String())
		return nil
	}

	feature.Enable(cfg.Features)
	generator.SetConfig(cfg.DataGen)

	fmt.Printf(logo, version.BuildVersion, strings.Repeat(" ", 17-len(versionString)))

	configureLogging(cfg)

	s, err := createServer(cfg)
	if err != nil {
		log.WithField("error", err).Error("error creating server")
	}

	go func() {
		err = s.Start()
		if err != nil {
			log.Error(err)
		}
	}()

	if ctx := cmd.Context(); ctx != nil {
		<-ctx.Done()
	}
	s.Close()
	return nil
}

func createServer(cfg *static.Config) (*server.Server, error) {
	pool := safe.NewPool(context.Background())
	app := runtime.New(cfg)

	watcher := server.NewConfigWatcher(cfg)
	scriptEngine := engine.New(watcher, app, cfg, true)
	certStore, err := cert.NewStore(cfg)
	if err != nil {
		return nil, err
	}
	http := server.NewHttpManager(scriptEngine, certStore, app)
	kafka := server.NewKafkaManager(scriptEngine, app)
	mqtt := server.NewMqttManager(scriptEngine, app)
	mailManager := server.NewMailManager(app, scriptEngine, certStore)
	ldap := server.NewLdapDirectoryManager(scriptEngine, certStore, app)

	watcher.AddListener(func(e dynamic.ConfigEvent) {
		kafka.UpdateConfig(e)
		mqtt.UpdateConfig(e)
		http.Update(e)
		mailManager.UpdateConfig(e)
		ldap.UpdateConfig(e)
		if err := scriptEngine.AddScript(e); err != nil {
			log.Error(err)
		}
		app.UpdateConfig(e)
	})

	if u, err := api.BuildUrl(cfg.Api); err == nil {
		err = http.AddInternalService("api", u, api.New(app, cfg.Api))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return server.NewServer(pool, app, watcher, kafka, http, mailManager, ldap, scriptEngine), nil
}

func configureLogging(cfg *static.Config) {
	stdlog.SetFlags(stdlog.Lshortfile | stdlog.LstdFlags)

	level, err := log.ParseLevel(cfg.Log.Level)
	if err != nil {
		log.WithField("logLevel", cfg.Log.Level).Errorf("error parsing log level: %v", err.Error())
	}
	log.SetLevel(level)

	if strings.ToLower(cfg.Log.Format) == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		formatter := &log.TextFormatter{DisableColors: false, FullTimestamp: true, DisableSorting: true}
		log.SetFormatter(formatter)
	}
}

func init() {
	registerDynamicTypes()
}

func registerDynamicTypes() {
	dynamic.Register("openapi", func(v version.Version) bool {
		return true
	}, &openapi.Config{})
	dynamic.Register("asyncapi", func(v version.Version) bool {
		return v.Major == 2
	}, &asyncApi.Config{})
	dynamic.Register("asyncapi", func(v version.Version) bool {
		return v.Major == 3
	}, &asyncapi3.Config{})
	dynamic.Register("swagger", func(v version.Version) bool {
		return true
	}, &swagger.Config{})
	dynamic.Register("ldap", func(v version.Version) bool {
		return true
	}, &directory.Config{})
	dynamic.Register("smtp", func(v version.Version) bool {
		return true
	}, &mail.Config{})
	dynamic.Register("mail", func(v version.Version) bool {
		return true
	}, &mail2.Config{})
}

func applyPositionalArgs(cfg *static.Config, args []string) error {
	cfg.Args = args
	err := cfg.Parse()
	if err != nil {
		return fmt.Errorf("parse config failed: %w", err)
	}
	return nil
}

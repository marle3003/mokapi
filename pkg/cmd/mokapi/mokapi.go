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
	"mokapi/pkg/cmd/mokapi/flags"
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

	flags.RegisterFileProvider(cmd)
	flags.RegisterGitProvider(cmd)
	flags.RegisterHttpProvider(cmd)
	flags.RegisterNpmProvider(cmd)

	// API
	cmd.Flags().Int("api-port", 8080, cli.FlagDoc{Short: "Port for the API server"})
	cmd.Flags().String("api-path", "", cli.FlagDoc{Short: "Path prefix for the API and dashboard"})
	cmd.Flags().String("api-base", "", cli.FlagDoc{Short: "Base path used when the API is behind a reverse proxy"})
	cmd.Flags().Bool("api-dashboard", true, cli.FlagDoc{Short: "Enable the web dashboard"})
	cmd.Flags().Bool("api-search-enabled", false, cli.FlagDoc{Short: "Enable search functionality in the dashboard"})

	cmd.Flags().String("root-ca-cert", "", cli.FlagDoc{Short: "Root CA certificate used for signing generated certificates"})
	cmd.Flags().String("root-ca-key", "", cli.FlagDoc{Short: "Private key of the root CA"})

	cmd.Flags().Int("event-store-default-size", 100, cli.FlagDoc{Short: "Default maximum number of stored events per API"})
	cmd.Flags().String("event-store", "", cli.FlagDoc{Short: "Configure event store using shorthand syntax"})
	cmd.Flags().DynamicInt("event-store-<name>-size", cli.FlagDoc{Short: "Override event store size for a specific API"})

	cmd.Flags().String("data-gen-optional-properties", "0.85", cli.FlagDoc{Short: ""})

	cmd.Flags().StringSlice("config", []string{}, true, cli.FlagDoc{Short: "Provide inline configuration data"})
	cmd.Flags().StringSlice("configs", []string{}, false, cli.FlagDoc{Short: "Provide inline configuration data"})

	// config file
	cmd.Flags().File("cli-input", cli.FlagDoc{Short: "Read configuration from a file"})
	cmd.Flags().Alias("cli-input", "config-file")

	// logging
	cmd.Flags().String("log-level", "info", cli.FlagDoc{Short: "Set log level (debug|info|warn|error)"}).WithExample(
		cli.Example{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--log-level warn"},
				{Title: "Env", Source: "MOKAPI_LOG_LEVEL=warn"},
				{Title: "File", Source: "log:\n  level: warn", Language: "yaml"},
			},
		},
	).WithDescription("The default level of log messages is info. You can set the log level to one of the following, listed in order of least to most information. The level is cumulative: for the debug level, the log file also includes messages at the info, warn, and error levels.\n- Debug\n- Info\n- Warn\n- Error\n")
	cmd.Flags().String("log-format", "text", cli.FlagDoc{Short: "Set log output format (text|json)"})

	cmd.Flags().String("generate-cli-skeleton", "", cli.FlagDoc{Short: "Generate a configuration skeleton and exit. If set without a value, generates the full skeleton. " +
		"If a value is provided, generates only the specified section (e.g. `providers`)."})

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

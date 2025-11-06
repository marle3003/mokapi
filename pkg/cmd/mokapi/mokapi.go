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

func NewCmdMokapi(ctx context.Context) *cli.Command {
	cmd := &cli.Command{
		Name:   "mokapi",
		Short:  "Start Mokapi and serve mocked APIs",
		Long:   `Mokapi is an easy, modern and flexible API mocking tool using Go and Javascript.`,
		Config: &static.Config{},
		Run: func(cmd *cli.Command, args []string) error {
			return runRoot(cmd.Config.(*static.Config), ctx)
		},
		Commands: []*cli.Command{
			NewCmdSampleData(),
		},
		EnvPrefix: "MOKAPI_",
	}

	cmd.Flags().BoolShort("version", "v", false, "Show version information and exit")
	cmd.Flags().Bool("generate-cli-skeleton", false, "Generates the skeleton configuration file")

	// config file
	cmd.Flags().String("config", "", "Path to configuration file (aliases: --config-file, --cli-input)")
	cmd.Flags().String("config-file", "", "Alias for --config")
	cmd.Flags().String("cli-input", "", "Alias for --config")

	// logging
	cmd.Flags().String("log-level", "info", "Mokapi log level (default is info)")
	cmd.Flags().String("log-format", "text", "Mokapi log format: json|text (default is text)")

	// file provider
	cmd.Flags().String("providers-file", "", "")
	cmd.Flags().StringSlice("providers-file-filename", []string{}, "Load the dynamic configuration from files", true)
	cmd.Flags().StringSlice("providers-file-filenames", []string{}, "Load the dynamic configuration from files", false)
	cmd.Flags().StringSlice("providers-file-directory", []string{}, "Load the dynamic configuration from directories", true)
	cmd.Flags().StringSlice("providers-file-directories", []string{}, "Load the dynamic configuration from directories", false)
	cmd.Flags().StringSlice("providers-file-skip-prefix", []string{"_"}, "", false)
	cmd.Flags().StringSlice("providers-file-include", []string{}, "", false)
	cmd.Flags().DynamicStringSlice("providers-file-include[<index>]", []string{}, "", false)

	// git provider
	cmd.Flags().String("providers-git", "", "")
	cmd.Flags().StringSlice("providers-git-url", []string{}, "", true)
	cmd.Flags().StringSlice("providers-git-urls", []string{}, "", false)
	cmd.Flags().String("providers-git-pull-interval", "3m", "")
	cmd.Flags().String("providers-git-temp-dir", "", "")
	cmd.Flags().StringSlice("providers-git-repository", []string{}, "flag for shorthand syntax", true)
	cmd.Flags().StringSlice("providers-git-repositories", []string{}, "flag for shorthand syntax", false)
	// git repository
	cmd.Flags().DynamicString("providers-git-repositories[<index>]", "", "set indexed repository using shorthand syntax")
	cmd.Flags().DynamicString("providers-git-repositories[<index>]-url", "", "set URL of the repository")
	cmd.Flags().DynamicStringSlice("providers-git-repositories[<index>]-file", []string{}, "Specifies an allow list of files to include in mokapi", true)
	cmd.Flags().DynamicStringSlice("providers-git-repositories[<index>]-files", []string{}, "Specifies an allow list of files to include in mokapi", false)
	cmd.Flags().DynamicStringSlice("providers-git-repositories[<index>]-include", []string{}, "Specifies an array of filenames or pattern to include in mokapi", false)
	cmd.Flags().DynamicString("providers-git-repositories[<index>]-auth-github", "", "Specifies an array of filenames or pattern to include in mokapi")
	cmd.Flags().DynamicString("providers-git-repositories[<index>]-pull-interval", "", "Specifies an array of filenames or pattern to include in mokapi")

	// http provider
	cmd.Flags().String("providers-http", "", "")
	cmd.Flags().StringSlice("providers-http-url", []string{}, "", true)
	cmd.Flags().StringSlice("providers-http-urls", []string{}, "", false)
	cmd.Flags().String("providers-http-poll-interval", "3m", "")
	cmd.Flags().String("providers-http-poll-timeout", "5s", "")
	cmd.Flags().String("providers-http-proxy", "", "")
	cmd.Flags().String("providers-http-tls-skip-verify", "", "")
	cmd.Flags().String("providers-http-ca", "", "Certificate authority")

	// npm provider
	cmd.Flags().String("providers-npm", "", "")
	cmd.Flags().StringSlice("providers-npm-global-folder", []string{}, "", true)
	cmd.Flags().StringSlice("providers-npm-global-folders", []string{}, "", false)
	// npm package
	cmd.Flags().StringSlice("providers-npm-package", []string{}, "flag for shorthand syntax", true)
	cmd.Flags().StringSlice("providers-npm-packages", []string{}, "flag for shorthand syntax", false)
	cmd.Flags().DynamicString("providers-npm-packages[<index>]", "", "set indexed repository using shorthand syntax")
	cmd.Flags().DynamicString("providers-npm-packages[<index>]-name", "", "set URL of the repository")
	cmd.Flags().DynamicStringSlice("providers-npm-packages[<index>]-file", []string{}, "Specifies an allow list of files to include in mokapi", true)
	cmd.Flags().DynamicStringSlice("providers-npm-packages[<index>]-files", []string{}, "Specifies an allow list of files to include in mokapi", false)
	cmd.Flags().DynamicStringSlice("providers-npm-packages[<index>]-include", []string{}, "Specifies an array of filenames or pattern to include in mokapi", false)

	// API
	cmd.Flags().Int("api-port", 8080, "API port (Default 8080). The API is available on the path /api")
	cmd.Flags().String("api-path", "", "The path prefix where dashboard is served (default empty)")
	cmd.Flags().String("api-base", "", "The base path of the dashboard useful in case of url rewriting (default empty)")
	cmd.Flags().Bool("api-dashboard", true, "Activate dashboard (default true). The dashboard is available at the same port as the API but on the path / by default.")
	cmd.Flags().Bool("api-search-enabled", false, "enables search feature")

	cmd.Flags().String("root-ca-cert", "", "CA Certificate for signing certificate generated at runtime")
	cmd.Flags().String("root-ca-key", "", "Private Key of CA for signing certificate generated at runtime")

	cmd.Flags().Int("event-store-default-size", 100, "Sets the default maximum number of events stored for each event type (e.g., HTTP, Kafka), unless overridden individually. (default 100)")
	cmd.Flags().String("event-store", "", "")
	cmd.Flags().DynamicInt("event-store-<name>-size", 100, "Overrides the default event store size for a specific API by name.")

	cmd.Flags().String("data-gen-optional-properties", "0.85", "")

	cmd.Flags().StringSlice("config", []string{}, "plain configuration data as argument", true)
	cmd.Flags().StringSlice("configs", []string{}, "plain configuration data as argument", false)

	cmd.Flags().BoolShort("help", "h", false, "Show help information")

	return cmd
}

func runRoot(cfg *static.Config, ctx context.Context) error {
	versionString := version.BuildVersion

	err := cfg.Parse()
	if err != nil {
		log.Errorf("parse config failed: %v", err)
		return nil
	}

	/*switch {
	case viper.GetBool("version"):
		fmt.Println(versionString)
		return nil
	case viper.GetBool("generate.cli.skeleton"):
		writeSkeleton(cfg)
		return nil
	}*/

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

	<-ctx.Done()
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

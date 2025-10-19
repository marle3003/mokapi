package main

import (
	"context"
	"fmt"
	stdlog "log"
	"mokapi/api"
	"mokapi/config/decoders"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/mail"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/feature"
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
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const logo = "888b     d888          888             d8888          d8b \n8888b   d8888          888            d88888          Y8P \n88888b.d88888          888           d88P888              \n888Y88888P888  .d88b.  888  888     d88P 888 88888b.  888 \n888 Y888P 888 d88\"\"88b 888 .88P    d88P  888 888 \"88b 888 \n888  Y8P  888 888  888 888888K    d88P   888 888  888 888 \n888   \"   888 Y88..88P 888 \"88b  d8888888888 888 d88P 888 \n888       888  \"Y88P\"  888  888 d88P     888 88888P\"  888 \n        v%s by Marcel Lehmann%s 888          \n        https://mokapi.io                    888          \n                                             888   \n"

func main() {
	versionString := version.BuildVersion

	cfg := static.NewConfig()
	configDecoders := []decoders.ConfigDecoder{decoders.NewDefaultFileDecoder(), decoders.NewFlagDecoder()}
	err := decoders.Load(configDecoders, cfg)
	if err != nil {
		log.Errorf("load config failed: %v", err)
		return
	}
	err = cfg.Parse()
	if err != nil {
		log.Errorf("parse config failed: %v", err)
		return
	}

	switch {
	case cfg.Help:
		printHelp()
		return
	case cfg.GenerateSkeleton != nil:
		writeSkeleton(cfg)
		return
	case cfg.Version:
		fmt.Println(versionString)
		return
	}

	feature.Enable(cfg.Features)
	generator.SetConfig(cfg.DataGen)

	fmt.Printf(logo, version.BuildVersion, strings.Repeat(" ", 17-len(versionString)))

	configureLogging(cfg)
	registerDynamicTypes()

	s, err := createServer(cfg)
	if err != nil {
		log.WithField("error", err).Error("error creating server")
	}

	exitChannel := make(chan os.Signal, 1)
	signal.Notify(exitChannel, os.Interrupt)
	signal.Notify(exitChannel, syscall.SIGTERM)
	signal.Notify(exitChannel, syscall.SIGKILL)
	go func() {
		<-exitChannel
		fmt.Println("Shutting down")
		s.Close()
		os.Exit(0)
	}()

	err = s.Start()
	if err != nil {
		log.Error(err)
	}
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

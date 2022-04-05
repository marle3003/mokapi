package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	stdlog "log"
	"mokapi/config/decoders"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/script"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/runtime"
	"mokapi/safe"
	"mokapi/server"
	"mokapi/server/cert"
	"mokapi/version"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const logo = "888b     d888          888             d8888          d8b \n8888b   d8888          888            d88888          Y8P \n88888b.d88888          888           d88P888              \n888Y88888P888  .d88b.  888  888     d88P 888 88888b.  888 \n888 Y888P 888 d88\"\"88b 888 .88P    d88P  888 888 \"88b 888 \n888  Y8P  888 888  888 888888K    d88P   888 888  888 888 \n888   \"   888 Y88..88P 888 \"88b  d8888888888 888 d88P 888 \n888       888  \"Y88P\"  888  888 d88P     888 88888P\"  888 \n        v%s by Marcel Lehmann%s 888          \n        https://github.com/marle3003/mokapi  888          \n                                             888   \n"

func main() {
	versionString := version.BuildVersion
	fmt.Printf(logo, version.BuildVersion, strings.Repeat(" ", 17-len(versionString)))

	cfg := static.NewConfig()
	configDecoders := []decoders.ConfigDecoder{&decoders.FileDecoder{}, &decoders.FlagDecoder{}}
	err := decoders.Load(configDecoders, cfg)
	if err != nil {
		fmt.Println("Error", err)
	}

	configureLogging(cfg)

	s, err := createServer(cfg)
	if err != nil {
		log.WithField("error", err).Error("error creating server")
	}

	ctx, cancel := context.WithCancel(context.Background())

	exitChannel := make(chan os.Signal, 1)
	signal.Notify(exitChannel, os.Interrupt)
	signal.Notify(exitChannel, syscall.SIGTERM)
	signal.Notify(exitChannel, syscall.SIGKILL)
	go func() {
		<-exitChannel
		fmt.Println("Shutting down")
		cancel()
		s.Close()
		os.Exit(0)
	}()

	err = s.Start(ctx)
	if err != nil {
		log.Error(err)
	}
}

func createServer(cfg *static.Config) (*server.Server, error) {
	pool := safe.NewPool(context.Background())
	app := runtime.New()
	watcher := dynamic.NewConfigWatcher(cfg)
	scriptEngine := engine.New(watcher)
	certStore, err := cert.NewStore(cfg)
	if err != nil {
		return nil, err
	}
	kafka := make(server.KafkaClusters)
	web := make(server.HttpServers)
	mail := make(server.SmtpServers)
	directories := make(server.LdapDirectories)
	managerHttp := server.NewHttpManager(web, scriptEngine, certStore, app)
	managerKafka := server.NewKafkaManager(kafka, scriptEngine, app)
	managerLdap := server.NewLdapDirectoryManager(directories, scriptEngine, certStore, app)

	watcher.AddListener(func(file *common.File) {
		managerKafka.UpdateConfig(file)
		managerHttp.Update(file)
		mail.UpdateConfig(file, certStore, scriptEngine)
		managerLdap.UpdateConfig(file)
	})
	watcher.AddListener(func(f *common.File) {
		if s, ok := f.Data.(*script.Script); ok {
			err := scriptEngine.AddScript(f.Url, s.Code)
			if err != nil {
				log.Error(err)
			}
		}
	})

	return server.NewServer(pool, watcher, kafka, web, mail, directories, scriptEngine), nil
}

func configureLogging(cfg *static.Config) {
	stdlog.SetFlags(stdlog.Lshortfile | stdlog.LstdFlags)

	if cfg.Log != nil {
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
}

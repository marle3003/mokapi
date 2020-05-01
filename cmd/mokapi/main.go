package main

import (
	"fmt"
	stdlog "log"
	"mokapi/config/decoders"
	"mokapi/config/static"
	"mokapi/server"
	"strings"

	log "github.com/sirupsen/logrus"
)

func main() {

	cfg := static.NewConfig()
	configDecoders := []decoders.ConfigDecoder{&decoders.FileDecoder{}, &decoders.FlagDecoder{}}
	error := decoders.Load(configDecoders, cfg)
	if error != nil {
		fmt.Println("Error", error)
	}

	configureLogging(cfg)

	s, error := createServer(cfg)
	if error != nil {
		log.WithField("error", error).Error("error creating server")
	}

	s.Start()

	s.Wait()
}

func createServer(cfg *static.Config) (*server.Server, error) {
	api := server.NewApiServer()
	watcher := server.NewConfigWatcher(&cfg.Providers.File)

	return server.NewServer(api, watcher), nil
}

func configureLogging(cfg *static.Config) {
	stdlog.SetFlags(stdlog.Lshortfile | stdlog.LstdFlags)

	if cfg.Log != nil {
		level, error := log.ParseLevel(cfg.Log.Level)
		if error != nil {
			log.WithField("logLevel", cfg.Log.Level).Error("Error parsing log level")
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

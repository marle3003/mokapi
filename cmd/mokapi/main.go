package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	stdlog "log"
	"mokapi/config/decoders"
	"mokapi/config/static"
	"mokapi/server"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
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

	s.Start()

	exitChannel := make(chan os.Signal, 1)
	signal.Notify(exitChannel, os.Interrupt)
	signal.Notify(exitChannel, syscall.SIGTERM)
	go func() {
		<-exitChannel
		fmt.Println("Shutting down")
		s.Stop()
		os.Exit(0)
	}()

	s.Wait()
}

func createServer(cfg *static.Config) (*server.Server, error) {
	return server.NewServer(cfg), nil
}

func configureLogging(cfg *static.Config) {
	stdlog.SetFlags(stdlog.Lshortfile | stdlog.LstdFlags)

	if cfg.Log != nil {
		level, err := log.ParseLevel(cfg.Log.Level)
		if err != nil {
			log.WithField("logLevel", cfg.Log.Level).Error("error parsing log level: %v", err.Error())
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

package main

import (
	"fmt"
	stdlog "log"
	"mokapi/config"
	"mokapi/config/decoders"
	"mokapi/config/static"
	"mokapi/server"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

func main() {

	cfg := static.NewConfig()
	configDecoders := []decoders.ConfigDecoder{&decoders.FileDecoder{}, &decoders.FlagDecoder{}}
	error := config.Load(configDecoders, cfg)
	if error != nil {
		fmt.Println("Error", error)
	}

	configureLogging(cfg)

	server, error := createServer(cfg)
	if error != nil {
		log.WithField("error", error).Error("error creating server")
	}

	server.Start()

	fmt.Println("Hello World")
}

func createServer(cfg *static.Config) (*server.Server, error) {
	manager := server.NewManager()
	entryPoints := manager.Build(cfg)

	api := server.NewApiServer()
	api.SetRouters(entryPoints)

	return server.NewServer(api), nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello www.example1.com!") // send data to client sid
	w.Write([]byte("<h1>Hello World!</h1>"))
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

package runtime

import (
	"fmt"
	"mokapi/config/static"
	"mokapi/runtime/events"
	"strings"

	stdlog "log"

	log "github.com/sirupsen/logrus"
)

type LogHook struct {
	sm      events.Handler
	enabled bool
}

type LogData struct {
	Message string `json:"message"`
	Level   string `json:"level"`
}

func NewLogHook(sm events.Handler) *LogHook {
	return &LogHook{sm: sm, enabled: true}
}

func (h *LogHook) Levels() []log.Level {
	return log.AllLevels
}

func (h *LogHook) Fire(entry *log.Entry) error {
	if !h.enabled {
		return nil
	}

	data := &LogData{
		Message: entry.Message,
		Level:   entry.Level.String(),
	}

	traits := extractTraits(entry)

	return h.sm.Push(data, traits)
}

func (h *LogHook) Disable() {
	h.enabled = false
}

func (d *LogData) Title() string {
	msg := d.Message
	if len(d.Message) > 50 {
		msg = msg[:47] + "..."
	}
	return fmt.Sprintf("%v: %v", d.Level, msg)
}

func extractTraits(entry *log.Entry) events.Traits {
	traits := events.NewTraits()
	if ns, ok := entry.Data["namespace"].(string); ok {
		traits.WithNamespace(ns)
	}
	if api, ok := entry.Data["api"]; ok {
		traits.WithName(api.(string))
	}
	traits.With("level", entry.Level.String())
	traits.With("type", "log")
	return traits
}

func configureLogging(cfg *static.Config) {
	stdlog.SetFlags(stdlog.Lshortfile | stdlog.LstdFlags)

	level := log.InfoLevel
	if cfg.Log.Level != "" {
		var err error
		level, err = log.ParseLevel(cfg.Log.Level)
		if err != nil {
			log.Errorf("error parsing log level: %v", err.Error())

		}
	}
	log.SetLevel(level)

	if strings.ToLower(cfg.Log.Format) == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		formatter := &log.TextFormatter{DisableColors: false, FullTimestamp: true, DisableSorting: true}
		log.SetFormatter(formatter)
	}
}

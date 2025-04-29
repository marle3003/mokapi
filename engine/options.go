package engine

import (
	"github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
)

func WithScriptLoader(loader ScriptLoader) Options {
	return func(e *Engine) {
		e.loader = loader
	}
}

func WithReader(reader dynamic.Reader) Options {
	return func(e *Engine) {
		e.reader = reader
	}
}

func WithKafkaClient(client common.KafkaClient) Options {
	return func(e *Engine) {
		e.kafkaClient = client
	}
}

func WithScheduler(scheduler Scheduler) Options {
	return func(e *Engine) {
		e.scheduler = scheduler
	}
}

func WithDefaultLogger() Options {
	return func(e *Engine) {
		e.logger = newLogger(logrus.StandardLogger())
	}
}

func WithLogger(logger common.Logger) Options {
	return func(e *Engine) {
		e.logger = logger
	}
}

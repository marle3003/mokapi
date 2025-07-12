package engine

import (
	"github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/runtime/events"
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

func WithApp(app *runtime.App) Options {
	return func(e *Engine) {
		e.jobCounter = app.Monitor.JobCounter
		e.sm = app.Events
	}
}

func WithStoreManager(eventManager *events.StoreManager) Options {
	return func(e *Engine) {
		e.sm = eventManager
	}
}

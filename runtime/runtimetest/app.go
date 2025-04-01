package runtimetest

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/providers/openapi"
	"mokapi/runtime"
)

func NewHttpApp(configs ...*openapi.Config) *runtime.App {
	app := runtime.New()
	for i, cfg := range configs {
		app.Http.Add(&dynamic.Config{
			Info: dynamictest.NewConfigInfo(dynamictest.WithUrl(fmt.Sprintf("%d", i))),
			Data: cfg,
		})
	}
	return app
}

type Options func(app *runtime.App)

type HttpInfoOptions func(hi *runtime.HttpInfo)

type KafkaInfoOptions func(ki *runtime.KafkaInfo)

func NewKafkaApp(configs ...*asyncapi3.Config) *runtime.App {
	app := runtime.New()
	for i, cfg := range configs {
		app.Kafka.Add(&dynamic.Config{
			Info: dynamictest.NewConfigInfo(dynamictest.WithUrl(fmt.Sprintf("%d", i))),
			Data: cfg,
		}, enginetest.NewEngine())
	}
	return app
}

func NewApp(opts ...Options) *runtime.App {
	app := runtime.New()
	for _, opt := range opts {
		opt(app)
	}
	return app
}

func WithKafka(name string, opts ...KafkaInfoOptions) Options {
	return func(app *runtime.App) {
		ki := &runtime.KafkaInfo{}
		for _, opt := range opts {
			opt(ki)
		}
		app.Kafka.Set(name, ki)
	}
}

func WithKafkaInfo(name string, ki *runtime.KafkaInfo) Options {
	return func(app *runtime.App) {
		app.Kafka.Set(name, ki)
	}
}

func WithKafkaConfig(c *asyncapi3.Config) KafkaInfoOptions {
	return func(ki *runtime.KafkaInfo) {
		ki.Config = c
	}
}

func WithKafkaStore(store *store.Store) KafkaInfoOptions {
	return func(ki *runtime.KafkaInfo) {
		ki.Store = store
	}
}

func WithLdapInfo(name string, li *runtime.LdapInfo) Options {
	return func(app *runtime.App) {
		app.Ldap.Set(name, li)
	}
}

func WithMailInfo(name string, mi *runtime.MailInfo) Options {
	return func(app *runtime.App) {
		app.Mail.Set(name, mi)
	}
}

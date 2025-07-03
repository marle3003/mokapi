package runtimetest

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/openapi"
	"mokapi/runtime"
)

func NewHttpApp(configs ...*openapi.Config) *runtime.App {
	cfg := &static.Config{}
	app := runtime.New(cfg)
	for i, cfg := range configs {
		app.AddHttp(&dynamic.Config{
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
	cfg := &static.Config{}
	app := runtime.New(cfg)
	for i, cfg := range configs {
		_, _ = app.Kafka.Add(&dynamic.Config{
			Info: dynamictest.NewConfigInfo(dynamictest.WithUrl(fmt.Sprintf("%d", i))),
			Data: cfg,
		}, enginetest.NewEngine())
	}
	return app
}

func NewApp(opts ...Options) *runtime.App {
	cfg := &static.Config{}
	app := runtime.New(cfg)
	for _, opt := range opts {
		opt(app)
	}
	return app
}

func WithKafkaInfo(name string, ki *runtime.KafkaInfo) Options {
	return func(app *runtime.App) {
		app.Kafka.Set(name, ki)
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

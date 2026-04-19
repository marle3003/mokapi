package runtimetest

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/directory"
	"mokapi/providers/mail"
	"mokapi/providers/openapi"
	"mokapi/runtime"
	"mokapi/runtime/events"
)

func NewHttpApp(configs ...*openapi.Config) *runtime.App {
	return NewApp(WithHttp(configs...))
}

type Options func(app *runtime.App)

type HttpInfoOptions func(hi *runtime.HttpInfo)

type KafkaInfoOptions func(ki *runtime.KafkaInfo)

func NewKafkaApp(configs ...*asyncapi3.Config) *runtime.App {
	return NewApp(WithKafka(configs...))
}

func NewApp(opts ...Options) *runtime.App {
	cfg := &static.Config{}
	app := runtime.New(cfg, &dynamictest.Reader{})
	app.Engine = enginetest.NewEngine()
	for _, opt := range opts {
		opt(app)
	}
	return app
}

func WithHttp(configs ...*openapi.Config) Options {
	return func(app *runtime.App) {
		for i, cfg := range configs {
			app.AddHttp(&dynamic.Config{
				Info: dynamictest.NewConfigInfo(dynamictest.WithUrl(fmt.Sprintf("%d", i))),
				Data: cfg,
			})
		}
	}
}

func WithKafka(configs ...*asyncapi3.Config) Options {
	return func(app *runtime.App) {
		for i, cfg := range configs {
			c := &dynamic.Config{
				Info: dynamictest.NewConfigInfo(dynamictest.WithUrl(fmt.Sprintf("%d", i))),
				Data: cfg,
			}

			_, err := app.Kafka.Add(c, app.Engine)
			if err != nil {
				panic(err)
			}
		}
	}
}

func WithKafkaInfo(name string, ki *runtime.KafkaInfo) Options {
	return func(app *runtime.App) {
		app.Kafka.Set(name, ki)
	}
}

func WithLdap(configs ...*directory.Config) Options {
	return func(app *runtime.App) {
		for i, cfg := range configs {
			c := &dynamic.Config{
				Info: dynamictest.NewConfigInfo(dynamictest.WithUrl(fmt.Sprintf("%d", i))),
				Data: cfg,
			}

			app.Ldap.Add(c, app.Engine)
		}
	}
}

func WithLdapInfo(name string, li *runtime.LdapInfo) Options {
	return func(app *runtime.App) {
		app.Ldap.Set(name, li)
	}
}

func WithMail(configs ...*mail.Config) Options {
	return func(app *runtime.App) {
		for i, cfg := range configs {
			c := &dynamic.Config{
				Info: dynamictest.NewConfigInfo(dynamictest.WithUrl(fmt.Sprintf("%d", i))),
				Data: cfg,
			}

			app.Mail.Add(c)
		}
	}
}

func WithMailInfo(name string, mi *runtime.MailInfo) Options {
	return func(app *runtime.App) {
		app.Mail.Set(name, mi)
	}
}

func WithEvent(traits events.Traits, data events.EventData) Options {
	return func(app *runtime.App) {
		err := app.Events.Push(data, traits)
		if err != nil {
			panic(err)
		}
	}
}

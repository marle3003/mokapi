package enginetest

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/engine/common"
	"mokapi/js"
	"mokapi/js/jstest"
	"mokapi/runtime"
	"mokapi/smtp"
	"path"
)

func NewEngine(opts ...engine.Options) *engine.Engine {
	loader := engine.NewDefaultScriptLoader(&static.Config{})

	app := runtime.New(&static.Config{}, &dynamictest.Reader{})

	opts = append([]engine.Options{
		engine.WithScriptLoader(engine.ScriptLoaderFunc(func(file *dynamic.Config, host common.Host) (common.Script, error) {
			ext := path.Ext(file.Info.Kernel().Path())
			if ext == ".js" || ext == ".ts" {
				// use this loader to ensure not to reuse the JavaScript Registry which is a singleton
				return jstest.New(js.WithFile(file), js.WithHost(host))
			}
			return loader.Load(file, host)
		})),
		engine.WithLogger(&Logger{}),
		engine.WithApp(app),
	}, opts...)
	return engine.NewEngine(opts...)
}

type HttpEventHandlerFunc func(request *common.HttpEventRequest, response *common.HttpEventResponse) []*common.Action

func (f HttpEventHandlerFunc) EmitHttp(request *common.HttpEventRequest, response *common.HttpEventResponse) []*common.Action {
	if f == nil {
		return nil
	}
	return f(request, response)
}

type KafkaEventHandlerFunc func(record *common.KafkaEventRecord) []*common.Action

func (f KafkaEventHandlerFunc) EmitKafka(record *common.KafkaEventRecord) []*common.Action {
	return f(record)
}

type MqttEventHandlerFunc func(message *common.MqttMessageEvent) []*common.Action

func (f MqttEventHandlerFunc) EmitMqtt(message *common.MqttMessageEvent) []*common.Action {
	return f(message)
}

type WebsocketEventHandlerFunc func(message *common.WebsocketEvent) []*common.Action

func (f WebsocketEventHandlerFunc) EmitWebsocket(message *common.WebsocketEvent) []*common.Action {
	return f(message)
}

type MailEventHandlerFunc func(message *smtp.Message, status *smtp.Status) []*common.Action

func (f MailEventHandlerFunc) EmitSmtp(message *smtp.Message, status *smtp.Status) []*common.Action {
	return f(message, status)
}

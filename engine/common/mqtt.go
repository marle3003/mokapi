package common

type MqttFilter struct{}

type MqttMessageEvent struct {
	Api    string
	Topic  string
	Retain bool
	Value  string
}

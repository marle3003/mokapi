package common

type MqttFilter struct {
	Api   string
	Topic string
}

type MqttMessageEvent struct {
	Api    string
	Topic  string
	Retain bool
	Value  string
}

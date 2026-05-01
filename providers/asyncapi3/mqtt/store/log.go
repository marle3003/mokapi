package store

type LogMessage struct {
	Topic      string   `json:"topic"`
	Value      LogValue `json:"value"`
	Api        string   `json:"api"`
	ClientId   string   `json:"clientId"`
	ScriptFile string   `json:"script"`
}

type LogValue struct {
	Value  string `json:"value"`
	Binary []byte `json:"binary"`
}

func (l *LogMessage) Title() string {
	return l.Topic
}

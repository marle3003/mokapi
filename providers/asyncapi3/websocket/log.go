package websocket

import "mokapi/engine/common"

type Log struct {
	Message Message          `json:"message"`
	Api     string           `json:"api"`
	Actions []*common.Action `json:"actions"`
}

func (l *Log) Title() string {
	return ""
}

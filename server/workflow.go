package server

import (
	"mokapi/engine"
)

func (s *Server) Run(event string, args ...interface{}) []*engine.Summary {
	return s.engine.Run(event, args...)
}

//func (s *Server) AddScript(key string, code string) {
//	if script, ok := s.scripts[key]; ok {
//		script.Close()
//	}
//
//	s.scripts[key] = lua.NewScript(key, code, kafka.NewKafka(s.writeKafkaMessage), s)
//}

package middlewares

import (
	"fmt"
	"mokapi/models"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

type replaceContent struct {
	config *models.ReplaceContent
	next   Middleware
}

func NewReplaceContent(config *models.ReplaceContent, next Middleware) Middleware {
	m := &replaceContent{config: config, next: next}
	return m
}

func (m *replaceContent) ServeData(data *Data, context *Context) {
	dataString, ok := data.Content.(string)
	if !ok {
		log.Errorf("Middleware replaceContent does only support string data")
		return
	}

	r, error := regexp.Compile(m.config.Regex)
	if error != nil {
		log.Errorf("Error in parsing regex '%v': %v", m.config.Regex, error.Error())
		return
	}
	replacement := ""
	switch strings.ToLower(m.config.Replacement.From) {
	case "requestbody":
		s, error := context.Body.Select(m.config.Replacement.Selector)
		if error != nil {
			log.Errorf("Error in selecting replacement: %v", error.Error())
			return
		}
		replacement = s
	}

	data.Content = r.ReplaceAllString(dataString, replacement)
	fmt.Print(data.Content)

	m.next.ServeData(data, context)
}

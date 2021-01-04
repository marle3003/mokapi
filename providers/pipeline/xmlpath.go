package pipeline

import (
	"github.com/pkg/errors"
	"gopkg.in/xmlpath.v2"
	"strings"
)

type XmlPathStep struct {
}

type XmlPathExecution struct {
	Selector string `step:"selector,position=0,required"`
	Text     string `step:"text,position=1,required"`
}

func (e *XmlPathStep) Start() StepExecution {
	return &XmlPathExecution{}
}

func (e *XmlPathExecution) Run(_ StepContext) (interface{}, error) {
	path, err := xmlpath.Compile(e.Selector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to compile selector '%v'", e.Selector)
	}
	reader := strings.NewReader(e.Text)
	node, err := xmlpath.Parse(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse text '%v'", e.Text)
	}
	if s, ok := path.String(node); ok {
		return s, nil
	}
	return nil, nil
}

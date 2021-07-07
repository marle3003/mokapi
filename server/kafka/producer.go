package kafka

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/workflow/runtime"
	"strconv"
)

type Producer struct {
	addMessage func(topic string, partition int, key, message interface{}) (interface{}, interface{}, error)
}

func newProducer(addMessage func(topic string, partition int, key, message interface{}) (interface{}, interface{}, error)) *Producer {
	return &Producer{
		addMessage: addMessage,
	}
}

func (p *Producer) Run(ctx *runtime.ActionContext) error {
	topic, ok := ctx.GetInputString("topic")
	if !ok {
		return fmt.Errorf("missing required parameter 'topic'")
	}

	key, _ := ctx.GetInput("key")
	message, _ := ctx.GetInput("message")

	partition := -1
	if p, ok := ctx.GetInput("partition"); ok {
		switch p := p.(type) {
		case int:
			partition = p
		case string:
			if i, err := strconv.Atoi(p); err != nil {
				return errors.Wrap(err, "partition parameter must be an integer")
			} else {
				partition = i
			}
		}

	}

	var err error
	key, message, err = p.addMessage(topic, partition, key, message)
	if err != nil {
		return err
	}

	ctx.Log("produced message to topic %q with key %q", topic, key)

	return nil
}

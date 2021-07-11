package kafka

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/workflow/runtime"
	"strconv"
)

type WriteMessage func(broker, topic string, partition int, key, message interface{}) (interface{}, interface{}, error)

type Producer struct {
	write WriteMessage
}

func NewProducer(write WriteMessage) *Producer {
	return &Producer{
		write: write,
	}
}

func (p *Producer) Run(ctx *runtime.ActionContext) error {
	topic, ok := ctx.GetInputString("topic")
	if !ok {
		return fmt.Errorf("missing required parameter 'topic'")
	}

	broker, ok := ctx.GetInputString("broker")

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
	key, message, err = p.write(broker, topic, partition, key, message)
	if err != nil {
		return err
	}

	ctx.Log("produced message to topic %q with key %q", topic, key)

	return nil
}

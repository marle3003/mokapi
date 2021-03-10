package asyncApi

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"strconv"
)

type KafkaBinding struct {
	Group Group
}

type Group struct {
	Initial Initial
}

type Initial struct {
	Rebalance Rebalance
}

type Rebalance struct {
	Delay int
}

func (b KafkaBinding) UnmarshalYAML(value *yaml.Node) error {
	m := make(map[string]string)
	value.Decode(m)

	if v, ok := m["group.initial.rebalance.delay.ms"]; ok {
		if i, err := strconv.Atoi(v); err != nil {
			return errors.Wrapf(err, "unable to convert 'group.initial.rebalance.delay.ms' to int")
		} else {
			b.Group.Initial.Rebalance.Delay = i
		}
	}

	return nil
}

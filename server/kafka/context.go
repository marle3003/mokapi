package kafka

import (
	"fmt"
	"mokapi/server/kafka/protocol"
	"regexp"
)

type GroupState int

const (
	Stable       GroupState = iota
	Joining      GroupState = 1
	AwaitingSync GroupState = 2
)

const (
	legalTopicChars    = "[a-zA-Z0-9\\._\\-]"
	maxTopicNameLength = 249
)

var topicNamePattern = regexp.MustCompile("^" + legalTopicChars + "+$")

type Cluster interface {
	Topic(string) Topic
	AddTopic(string) (Topic, error)
	Topics() []Topic
	Brokers() []Broker
	Groups() []Group
	Group(string) Group
	NewGroup(string) (Group, error)
}

type Topic interface {
	Name() string
	Partition(int) Partition
	Partitions() []Partition
}

type Partition interface {
	Index() int
	Read(offset int64, maxBytes int) (protocol.RecordBatch, protocol.ErrorCode)
	Write(protocol.RecordBatch)
	Offset() int64
	StartOffset() int64

	Leader() Broker
	Replicas() []Broker
}

type Broker interface {
	Id() int
	Host() string
	Port() int
}

type Group interface {
	Name() string
	Coordinator() (Broker, error)
	State() GroupState
	SetState(state GroupState)

	Generation() *Generation
	SetGeneration(generation *Generation)
	NewGeneration() *Generation

	Commit(topic string, partition int, offset int64)
	Offset(topic string, partition int) int64
}

type Generation struct {
	Id       int
	Protocol string
	Members  map[string]*Member
}

type Member struct {
	Client     *ClientContext
	Partitions []Partition
}

func validateTopicName(s string) error {
	switch {
	case len(s) == 0:
		return fmt.Errorf("topic name can not be empty")
	case s == "." || s == "..":
		return fmt.Errorf("topic name can not be %v", s)
	case len(s) > maxTopicNameLength:
		return fmt.Errorf("topic name can not be longer than %v", maxTopicNameLength)
	case !topicNamePattern.Match([]byte(s)):
		return fmt.Errorf("topic name is not valid, valid characters are ASCII alphanumerics, '.', '_', and '-'")
	}

	return nil
}

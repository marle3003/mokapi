package kafka

import (
	"fmt"
	"regexp"
)

const (
	legalTopicChars    = "[a-zA-Z0-9\\._\\-]"
	maxTopicNameLength = 249
)

var topicNamePattern = regexp.MustCompile("^" + legalTopicChars + "+$")

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

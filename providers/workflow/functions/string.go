package functions

import (
	"fmt"
	"mokapi/providers/utils"
	"strings"
)

func HasPrefix(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("invalid number of arguments provided: expected 2 got %v", len(args))
	}
	s := utils.ToString(args[0])
	prefix := utils.ToString(args[1])
	return strings.HasPrefix(s, prefix), nil
}

func HasSuffix(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("invalid number of arguments provided: expected 2 got %v", len(args))
	}
	s := utils.ToString(args[0])
	prefix := utils.ToString(args[1])
	return strings.HasSuffix(s, prefix), nil
}

func Contains(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("invalid number of arguments provided: expected 2 got %v", len(args))
	}
	s := utils.ToString(args[0])
	substr := utils.ToString(args[1])
	return strings.Contains(s, substr), nil
}

func ToLower(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of arguments provided: expected 1 got %v", len(args))
	}
	s := utils.ToString(args[0])
	return strings.ToLower(s), nil
}

func ToUpper(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of arguments provided: expected 1 got %v", len(args))
	}
	s := utils.ToString(args[0])
	return strings.ToUpper(s), nil
}

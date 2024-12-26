package parser

import (
	"errors"
	"fmt"
	"strings"
)

type PathErrorString interface {
	ToString(path string) string
}

type Error struct {
	NumErrors int
	Err       error
}

type PathError struct {
	Err  error
	Path string
}

type PathErrors []error

type PathCompositionError struct {
	Path    string
	Message string
	Errs    []error
}

func wrapError(path string, err error) *PathError {
	return &PathError{Err: err, Path: path}
}

func wrapErrorPath(path []string, err error) error {
	n := len(path) - 1
	// loop reverse
	for i := range path {
		err = wrapError(path[n-i], err)
	}
	return err
}

func Errorf(path string, format string, args ...interface{}) *PathError {
	return &PathError{Err: fmt.Errorf(format, args...), Path: path}
}

func NumErrors(err error) int {
	var comp *PathCompositionError
	if errors.As(err, &comp) {
		return len(comp.Errs)
	}
	var list *PathErrors
	if errors.As(err, &list) {
		return len(*list)
	}

	return 1
}

func (e *PathError) Error() string {
	return e.ToString("#")
}

func (e *PathError) ToString(path string) string {
	path = fmt.Sprintf("%s/%s", path, e.Path)

	var target PathErrorString
	if errors.As(e.Err, &target) {
		return target.ToString(path)
	}
	return fmt.Sprintf("%s\nschema path %s", e.Err, path)
}

func (e *Error) Error() string {
	num := NumErrors(e.Err)
	if num > 1 {
		return fmt.Sprintf("found %v errors:\n%s", num, e.Err)
	}
	return fmt.Sprintf("found %v error:\n%s", num, e.Err)
}

func (e *PathCompositionError) Error() string {
	return e.ToString("#")
}

func (e *PathCompositionError) ToString(path string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s:\n", e.Message))
	path = fmt.Sprintf("%s/%s", path, e.Path)
	for i, err := range e.Errs {
		if i > 0 {
			sb.WriteString("\n")
		}

		var target PathErrorString
		if errors.As(err, &target) {
			sb.WriteString(fmt.Sprintf("%v", target.ToString(path)))
		} else {
			sb.WriteString(fmt.Sprintf("%v", err.Error()))
		}
	}
	return sb.String()
}

func (e *PathCompositionError) append(err error) {
	var target *PathErrors
	if errors.As(err, &target) {
		e.Errs = append(e.Errs, *target...)
	} else {
		e.Errs = append(e.Errs, err)
	}
}

func (e *PathErrors) Error() string {
	return e.ToString("#")
}

func (e *PathErrors) ToString(path string) string {
	var sb strings.Builder
	for i, err := range *e {
		if i > 0 {
			sb.WriteString("\n")
		}
		var target PathErrorString
		if errors.As(err, &target) {
			sb.WriteString(fmt.Sprintf("%s", target.ToString(path)))
		} else {
			sb.WriteString(fmt.Sprintf("%v", err.Error()))
		}

	}
	return sb.String()
}

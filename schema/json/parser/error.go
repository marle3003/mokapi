package parser

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

type Error struct {
	err error
}

type ErrorList []error

type ErrorDetail struct {
	Message string    `json:"message,omitempty"`
	Field   string    `json:"field,omitempty"`
	Errors  ErrorList `json:"errors,omitempty"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	num := countErrors(e.err)
	s := toString2(e.err, []string{"#"})

	return fmt.Sprintf("error count %d:%s", num, s)
}

func (e *ErrorList) Error() string {
	return ""
}

func (e *ErrorDetail) Error() string {
	return e.toString([]string{"#"})
}

func (e *ErrorList) toString(path []string) string {
	var sb strings.Builder
	for _, err := range *e {
		sb.WriteString(toString2(err, path))
	}
	return sb.String()
}

func (e *ErrorDetail) toString(path []string) string {
	var sb strings.Builder
	path = append(path, e.Field)
	s := strings.Join(path, "/")

	if len(e.Errors) > 0 {
		if e.Field != "" {
			if e.Message != "" {
				sb.WriteString(fmt.Sprintf("%s: %s", s, e.Message))
			} else {
				sb.WriteString(fmt.Sprintf("%s:", s))
			}
		} else {
			sb.WriteString(e.Message)
		}

		sb.WriteString(e.Errors.toString(path))

		return sb.String()
	} else {
		return fmt.Sprintf("%s: %s", s, e.Message)
	}
}

func (e *ErrorDetail) count() int {
	if len(e.Errors) == 0 {
		return 1
	}
	return countErrors(&e.Errors)
}

func (e *ErrorList) count() int {
	sum := 0
	for _, err := range *e {
		sum += countErrors(err)
	}
	return sum
}

func countErrors(err error) int {
	var detail *ErrorDetail
	if errors.As(err, &detail) {
		return detail.count()
	}

	var list *ErrorList
	if errors.As(err, &list) {
		return list.count()
	}

	return 1
}

func toString2(err error, path []string) string {
	prefix := strings.Repeat("\t", len(path)-1)

	var detail *ErrorDetail
	if errors.As(err, &detail) {
		return fmt.Sprintf("\n%s- %s", prefix, detail.toString(path))
	}
	var list *ErrorList
	if errors.As(err, &list) {
		return list.toString(path)
	}

	return fmt.Sprintf("\n%s- %s", prefix, err.Error())
}

func wrapErrorDetail(err error, detail *ErrorDetail) error {
	var base *ErrorDetail
	if errors.As(err, &base) {
		if detail.Field != "" {
			base.Field = fmt.Sprintf("%s/%s", detail.Field, base.Field)
		}
		if detail.Message != "" {
			base.Message = fmt.Sprintf("%s: %s", detail.Message, base.Message)
		}
		return base
	}

	var baseList *ErrorList
	if errors.As(err, &baseList) {
		for _, item := range *baseList {
			_ = wrapErrorDetail(item, detail)
		}
		return baseList
	}

	if detail.Message != "" {
		detail.Message = fmt.Sprintf("%s: %s", detail.Message, err.Error())
	} else {
		detail.Message = err.Error()
	}
	return detail
}

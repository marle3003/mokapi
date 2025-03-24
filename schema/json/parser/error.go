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

type ErrorComposition struct {
	Message string    `json:"message,omitempty"`
	Field   string    `json:"field,omitempty"`
	Errors  ErrorList `json:"errors,omitempty"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	items := toString(e.err, []string{"#"}, "")

	var sb strings.Builder
	for _, item := range items {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("\t%s", item))
	}

	return fmt.Sprintf("error count %d:\n%s", len(items), sb.String())
}

func (e *ErrorList) Error() string {
	list := e.toString([]string{"#"}, "")
	var sb strings.Builder
	for _, item := range list {
		if sb.Len() > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(item)
	}
	return sb.String()
}

func (e *ErrorDetail) Error() string {
	list := e.toString([]string{"#"}, "")
	var sb strings.Builder
	for _, item := range list {
		if sb.Len() > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(item)
	}
	return sb.String()
}

func (e *ErrorComposition) Error() string {
	list := e.toString([]string{"#"}, "")
	var sb strings.Builder
	for _, item := range list {
		if sb.Len() > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(item)
	}
	return sb.String()
}

func (e *ErrorList) toString(path []string, output string) []string {
	var list []string
	for _, err := range *e {
		list = append(list, toString(err, path, output)...)
	}
	return list
}

func (e *ErrorDetail) toString(path []string, output string) []string {
	path = append(path, e.Field)

	if len(e.Errors) > 0 {
		return e.Errors.toString(path, output)
	} else {
		s := strings.Join(path, "/")
		if output == "json" {
			return []string{fmt.Sprintf(`{"schema":"%s","message":"%s"}`, s, e.Message)}
		}
		return []string{fmt.Sprintf("- %s: %s", s, e.Message)}
	}
}

func (e *ErrorComposition) toString(path []string, output string) []string {
	path = append(path, e.Field)
	var result []string
	s := strings.Join(path, "/")

	if output == "json" {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(`{"schema":"%s","message":"%s"`, s, e.Message))
		for _, err := range e.Errors {
			items := toString(err, path, output)
			for _, item := range items {
				result = append(result, item)
			}
		}
		if len(result) > 0 {
			sb.WriteString(fmt.Sprintf(`,"errors":[%s]`, strings.Join(result, ",")))
		}
		return []string{sb.String() + "}"}
	}

	result = append(result, fmt.Sprintf("- %s: %s", s, e.Message))
	for _, err := range e.Errors {
		items := toString(err, path, output)
		for _, item := range items {
			result = append(result, fmt.Sprintf("\t%s", item))
		}
	}
	return result
}

func toString(err error, path []string, output string) []string {
	var detail *ErrorDetail
	if errors.As(err, &detail) {
		return detail.toString(path, output)
	}
	var list *ErrorList
	if errors.As(err, &list) {
		return list.toString(path, output)
	}
	var comp *ErrorComposition
	if errors.As(err, &comp) {
		return comp.toString(path, output)
	}

	if output == "json" {
		return []string{fmt.Sprintf(`"%s"`, err.Error())}
	}

	return []string{fmt.Sprintf("- %s", err.Error())}
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

func (e *ErrorDetail) append(err error) {
	e.Errors = append(e.Errors, err)
}

func Marshal(err error) string {
	var e *Error
	if errors.As(err, &e) {
		items := toString(e.err, []string{"#"}, "json")
		return fmt.Sprintf("[%s]", strings.Join(items, ","))
	}
	return fmt.Sprintf(`[%s]`, err.Error())
}

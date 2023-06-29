package imap

import (
	"fmt"
)

type responseStatus string

type responseCode string

const (
	ok  responseStatus = "OK"
	no  responseStatus = "NO"
	bad responseStatus = "BAD"

	alert responseCode = "ALERT"
)

type response struct {
	status responseStatus
	code   responseCode
	text   string
}

func (r *response) String() string {
	if len(r.code) > 0 {
		return fmt.Sprintf("%v [%v] %v", r.status, r.code, r.text)
	} else {
		return fmt.Sprintf("%v %v", r.status, r.text)
	}
}

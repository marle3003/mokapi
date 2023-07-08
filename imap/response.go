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

	alert     responseCode = "ALERT"
	readWrite responseCode = "READ-WRITE"
	cannot    responseCode = "CANNOT"
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
		if len(r.status) > 0 {
			return fmt.Sprintf("%v %v", r.status, r.text)
		}
		return r.text
	}
}

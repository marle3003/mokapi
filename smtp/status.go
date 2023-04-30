package smtp

import "fmt"

// https://www.iana.org/assignments/smtp-enhanced-status-codes/smtp-enhanced-status-codes.xhtml

type StatusCode int16

const (
	StatusClose                 StatusCode = 221
	StatusAuthSucceeded         StatusCode = 235
	StatusOk                    StatusCode = 250
	StatusAuthMethodAccepted    StatusCode = 334
	StatusActionAborted         StatusCode = 451
	StatusUnknownCommand        StatusCode = 500
	StatusSyntaxError           StatusCode = 501
	StatusCommandNotImplemented StatusCode = 502
	BadSequenceOfCommands       StatusCode = 503
	StatusReject                StatusCode = 521
	AuthenticationRequire       StatusCode = 530
	StatusStartMailInput        StatusCode = 354
)

type EnhancedStatusCode [3]int8

var Undefined = EnhancedStatusCode{-1, -1, -1}
var Success = EnhancedStatusCode{2, 0, 0}
var InvalidCommand = EnhancedStatusCode{5, 5, 1}
var SyntaxError = EnhancedStatusCode{5, 5, 2}
var UndefinedError = EnhancedStatusCode{5, 0, 0}
var SecurityError = EnhancedStatusCode{5, 7, 0}

func (e *EnhancedStatusCode) String() string {
	return fmt.Sprintf("%v.%v.%v", e[0], e[1], e[2])
}

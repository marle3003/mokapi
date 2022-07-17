package smtp

import "fmt"

type StatusCode int16

const (
	StatusClose          StatusCode = 221
	StatusOk             StatusCode = 250
	StatusSyntaxError    StatusCode = 501
	StatusStartMailInput StatusCode = 354
)

type EnhancedStatusCode [3]int8

var Undefined = EnhancedStatusCode{-1, -1, -1}
var Success = EnhancedStatusCode{2, 0, 0}
var SyntaxError = EnhancedStatusCode{5, 5, 2}

func (e *EnhancedStatusCode) String() string {
	return fmt.Sprintf("%v.%v.%v", e[0], e[1], e[2])
}

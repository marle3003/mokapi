package ldap

import (
	"github.com/dop251/goja"
)

type searchScope struct {
	BaseObject   int `json:"BaseObject"`
	SingleLevel  int `json:"SingleLevel"`
	WholeSubtree int `json:"WholeSubtree"`
}

type resultCode struct {
	Success                int `json:"Success"`
	OperationsError        int `json:"OperationsError"`
	ProtocolError          int `json:"ProtocolError"`
	SizeLimitExceeded      int `json:"SizeLimitExceeded"`
	AuthMethodNotSupported int `json:"AuthMethodNotSupported"`
	CannotCancel           int `json:"CannotCancel"`
}

var (
	scope = searchScope{
		BaseObject:   1,
		SingleLevel:  2,
		WholeSubtree: 3,
	}
	code = resultCode{
		Success:                0,
		OperationsError:        1,
		ProtocolError:          2,
		SizeLimitExceeded:      4,
		AuthMethodNotSupported: 7,
		CannotCancel:           121,
	}
)

func Require(_ *goja.Runtime, module *goja.Object) {
	obj := module.Get("exports").(*goja.Object)
	obj.Set("SearchScope", scope)
	obj.Set("ResultCode", code)
}

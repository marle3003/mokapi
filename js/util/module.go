package util

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/lib"
	"time"

	"github.com/dop251/goja"
)

func JsType(v any) string {
	return lib.TypeFrom(v)
}

func GetScriptFile(vm *goja.Runtime) *dynamic.Config {
	return vm.Get("mokapi/internal").(*goja.Object).Get("file").Export().(*dynamic.Config)
}

func DefaultRetryArgs() common.RetryArgs {
	return common.RetryArgs{
		MaxRetryTime:     3 * time.Minute,
		InitialRetryTime: 500 * time.Millisecond,
		Retries:          10,
		Factor:           2,
	}
}

func ConvertToRetryArgs(m map[string]any) (common.RetryArgs, error) {
	retryArgs := DefaultRetryArgs()
	if i, ok := m["maxRetryTime"]; ok {
		switch v := i.(type) {
		case int64:
			retryArgs.MaxRetryTime = time.Duration(v) * time.Millisecond
		case string:
			d, err := time.ParseDuration(v)
			if err != nil {
				return retryArgs, fmt.Errorf("parse maxRetryTime failed: %w", err)
			}
			retryArgs.MaxRetryTime = d
		default:
			return retryArgs, fmt.Errorf("type %T for maxRetryTime not supported", v)
		}

	}
	if i, ok := m["initialRetryTime"]; ok {
		switch v := i.(type) {
		case int64:
			retryArgs.InitialRetryTime = time.Duration(v) * time.Millisecond
		case string:
			d, err := time.ParseDuration(v)
			if err != nil {
				return retryArgs, fmt.Errorf("parse initialRetryTime failed: %w", err)
			}
			retryArgs.InitialRetryTime = d
		default:
			return retryArgs, fmt.Errorf("type %T for initialRetryTime not supported", v)
		}
	}
	if v, ok := m["retries"]; ok {
		retryArgs.Retries = int(v.(int64))
	}
	if v, ok := m["factor"]; ok {
		retryArgs.Factor = int(v.(int64))
	}

	return retryArgs, nil
}

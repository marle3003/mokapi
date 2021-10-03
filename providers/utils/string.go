package utils

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

func ToString(i interface{}) string {
	if i == nil {
		return ""
	}

	v := reflect.ValueOf(i)
	var ptr reflect.Value
	if v.Type().Kind() == reflect.Ptr {
		ptr = v
		v = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(i))
		temp := ptr.Elem()
		temp.Set(v)
	}

	if v.Kind() == reflect.Map {
		var sb strings.Builder
		for _, k := range v.MapKeys() {
			if sb.Len() > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%v=%v", k, ToString(v.MapIndex(k).Interface())))
		}
		return fmt.Sprintf("{%v}", sb.String())
	} else if v.Kind() == reflect.Struct {
		var sb strings.Builder
		for i := 0; i < v.NumField(); i++ {
			if sb.Len() > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%v=%v", v.Type().Field(i).Name, ToString(v.Field(i).Interface())))
		}
		return fmt.Sprintf("[%v]", sb.String())
	}

	return fmt.Sprintf("%v", i)
}

func ParseUrl(s string) (host string, port int, err error) {
	u, err := url.Parse("//" + s)
	if err != nil {
		err = fmt.Errorf("invalid format in url %q", s)
		return
	}
	host = u.Hostname()

	portString := u.Port()
	if len(portString) == 0 {
		port = 80
	} else {
		port, err = strconv.Atoi(portString)
		if err != nil {
			err = fmt.Errorf("invalid port format in url %q", s)
		}
	}

	return
}

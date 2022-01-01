package test

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

var TestError = errors.New("TESTING ERROR")

// Assert fails the test if the condition is false.
func Assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// Ok fails the test if err is not nil.
func Ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// Error fails the test if err is nil
func Error(tb testing.TB, err error) {
	if err == nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: expected an error\n\n", filepath.Base(file), line)
		tb.FailNow()
	}
}

// EqualError fails the test if error message is not equal
func EqualError(tb testing.TB, errMsg string, err error) {
	Error(tb, err)

	if errMsg != err.Error() {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: expected error: %s\033[39m\n\ngot: %s\u001B[39m\n\n", filepath.Base(file), line, errMsg, err.Error())
		tb.FailNow()
	}
}

// Equals fails the test if exp is not equal to act.
func Equals(tb testing.TB, exp, act interface{}) {
	if isNil(exp) && isNil(act) {
		return
	}
	if !reflect.DeepEqual(exp, act) {
		if !equal(exp, act) {
			_, file, line, _ := runtime.Caller(1)
			fmt.Printf("\033[31m%s:%d:\n\n\texp: %v\n\n\tgot: %v\033[39m\n\n", filepath.Base(file), line, exp, act)
			tb.FailNow()
		}
	}
}

func IsTrue(t *testing.T, b bool) {
	Equals(t, true, b)
}

func isNil(x interface{}) bool {
	if x == nil {
		return true
	}
	v := reflect.ValueOf(x)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func NewNullLogger() *test.Hook {
	logrus.SetOutput(ioutil.Discard)
	return test.NewGlobal()
}

func equal(exp, act interface{}) bool {
	if _, ok := exp.(error); ok {
		return false
	}
	if _, ok := act.(error); ok {
		return false
	}

	v1 := reflect.ValueOf(exp)
	v2 := reflect.ValueOf(act)

	if v1.Kind() != v2.Kind() {
		return false
	}

	if v1.Kind() == reflect.Ptr {
		v1 = v1.Elem()
		v2 = v2.Elem()
	}

	if v1.Kind() != reflect.Struct {
		return false
	}

	if v1.NumField() != v2.NumField() {
		return false
	}
	t := v1.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		// skip private fields
		if f.Name != strings.Title(f.Name) {
			continue
		}
		f1 := v1.FieldByName(f.Name)
		f2 := v2.FieldByName(f.Name)

		i1 := f1.Interface()
		i2 := f2.Interface()

		if !reflect.DeepEqual(i1, i2) && !equal(i1, i2) {
			return false
		}
	}

	return true
}

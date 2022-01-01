package common

import (
	"encoding/json"
	"github.com/pkg/errors"
	"net/url"
	"reflect"
)

var UnknownFile = errors.New("unknown file")

type File struct {
	Url  *url.URL
	Data interface{}

	Listeners []func(*File)

	AllowParsingUnknownType bool
	AsPlainText             bool
}

func (f *File) Changed() {
	for _, l := range f.Listeners {
		l(f)
	}
}

type FileOptions func(file *File)

func WithListener(f func(file *File)) FileOptions {
	return func(file *File) {
		file.Listeners = append(file.Listeners, f)
	}
}

func WithData(data interface{}) FileOptions {
	return func(file *File) {
		if file.Data == nil {
			file.Data = data
		}
	}
}

func WithParent(parent *File) FileOptions {
	return func(file *File) {
		file.Listeners = append(file.Listeners, func(_ *File) {
			parent.Changed()
		})
	}
}

func AllowParsingAny() FileOptions {
	return func(file *File) {
		file.AllowParsingUnknownType = true
	}
}

func AsPlaintext() FileOptions {
	return func(file *File) {
		file.AsPlainText = true
	}
}

func (f *File) UnmarshalYAML(unmarshal func(interface{}) error) error {
	data := make(map[string]string)
	_ = unmarshal(data)

	for _, ct := range configTypes {
		if _, ok := data[ct.header]; ok {
			if f.Data != nil {
				i := reflect.New(ct.configType).Interface()
				err := unmarshal(i)
				v := reflect.ValueOf(f.Data).Elem()
				v.Set(reflect.ValueOf(i).Elem())
				return err
			} else {
				f.Data = reflect.New(ct.configType).Interface()
				return unmarshal(f.Data)
			}
		}
	}

	if f.Data == nil {
		if f.AllowParsingUnknownType {
			f.Data = make(map[string]interface{})
		} else {
			return nil
		}
	}

	err := unmarshal(f.Data)
	if err != nil {
		return err
	}

	return nil
}

func (f *File) UnmarshalJSON(b []byte) error {
	data := make(map[string]string)
	_ = json.Unmarshal(b, &data)

	for _, ct := range configTypes {
		if _, ok := data[ct.header]; ok {
			f.Data = reflect.New(ct.configType).Interface()
			return json.Unmarshal(b, f.Data)
		}
	}

	if f.Data == nil {
		if f.AllowParsingUnknownType {
			f.Data = make(map[string]interface{})
		} else {
			return nil
		}
	}

	err := json.Unmarshal(b, f.Data)
	if err != nil {
		return err
	}

	return nil
}

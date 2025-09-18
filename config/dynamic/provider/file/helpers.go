package file

import (
	"fmt"
	"net/url"
	"path/filepath"
)

func ParseUrl(path string) (*url.URL, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return url.ParseRequestURI(fmt.Sprintf("file:%v", abs))
}

package file

import (
	"net/url"
	"os"
	"path/filepath"
)

func ParseUrl(path string) (*url.URL, error) {
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(wd, path)
	}

	return url.ParseRequestURI("file:" + path)
}

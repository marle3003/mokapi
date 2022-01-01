package file

import (
	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func isDir(path string) (bool, error) {
	if fi, err := os.Stat(path); err != nil {
		return false, err
	} else if fi.IsDir() {
		return true, nil
	}
	return false, nil
}

func isValidConfigFile(path string) bool {
	if skipPath(path) {
		return false
	}
	switch filepath.Ext(path) {
	case ".yml", ".yaml", ".json", ".tmpl", ".lua":
		return true
	default:
		return false
	}
}

func skipPath(path string) bool {
	name := filepath.Base(path)
	// TODO: make skip char configurable
	if strings.HasPrefix(name, "_") {
		log.Infof("skipping config %v", name)
		return true
	}
	return false
}

func ParseUrl(path string) (*url.URL, error) {
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(wd, path)
	}

	path = filepath.ToSlash(path)

	// windows needs that?
	//return url.ParseRequestURI("file:///" + path)
	return url.ParseRequestURI("file://" + path)
}

func MustParseUrl(path string) *url.URL {
	u, err := ParseUrl(path)
	if err != nil {
		panic(err)
	}
	return u
}

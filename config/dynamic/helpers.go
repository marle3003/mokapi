package dynamic

import (
	log "github.com/sirupsen/logrus"
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
	case ".yml", ".yaml", ".json", ".tmpl":
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

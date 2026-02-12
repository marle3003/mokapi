package mokapi

import (
	"fmt"
	"mokapi/config/static"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func writeSkeleton(section string) {
	var skeleton interface{}
	if section != "" {
		paths := parsePath(section)
		current := reflect.ValueOf(static.NewConfig())
		for _, path := range paths {
			if current.Kind() == reflect.Pointer {
				current = current.Elem()
			}
			field := current.FieldByNameFunc(func(f string) bool {
				return strings.ToLower(f) == path
			})
			if !field.IsValid() {
				log.Errorf("unable to find config element: %v", section)
				return
			}
			current = field
		}
		skeleton = current.Interface()
	} else {
		skeleton = static.NewConfig()
	}

	b, err := yaml.Marshal(skeleton)
	if err != nil {
		log.Errorf("unable to write skeleton: %v", err)
		return
	}
	fmt.Print(string(b))
}

func parsePath(key string) []string {
	var paths []string
	split := strings.FieldsFunc(key, func(r rune) bool {
		return r == '.' || r == '-'
	})

	for _, v := range split {
		if strings.HasSuffix(v, "]") {
			index := strings.Index(v, "[")
			paths = append(paths, v[:index], v[index:])
		} else {
			paths = append(paths, v)
		}
	}
	return paths
}

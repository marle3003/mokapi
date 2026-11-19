package cli

import (
	"reflect"
	"strings"
	"unicode"
)

const (
	noConfigKey = "-"
	listItem    = "[]"
)

// configKey returns the key under which a flag is set in the config file, e.g. log.level,
// providers.git.repositories[].url or event.store.<name>.size. It returns "-" if there is none.
// It follows the same field matching as the binder, so the documentation can't drift from it.
func configKey(config any, flagName string) string {
	if config == nil {
		return noConfigKey
	}
	segments, ok := walkConfigKey(reflect.TypeOf(config), ParsePath(flagName))
	if !ok || len(segments) == 0 {
		return noConfigKey
	}
	return joinKey(segments)
}

// joinKey separates segments with a dot. A list index belongs to the segment before it: repositories[].url.
func joinKey(segments []string) string {
	key := ""
	for _, s := range segments {
		if key != "" && s != listItem {
			key += "."
		}
		key += s
	}
	return key
}

func walkConfigKey(t reflect.Type, paths []string) ([]string, bool) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if len(paths) == 0 {
		return nil, true
	}

	switch t.Kind() {
	case reflect.Struct:
		return walkStructKey(t, paths)
	case reflect.Slice:
		if !isIndexSegment(paths[0]) {
			return nil, false
		}
		rest, ok := walkConfigKey(t.Elem(), paths[1:])
		return append([]string{listItem}, rest...), ok
	case reflect.Map:
		rest, ok := walkConfigKey(t.Elem(), paths[1:])
		return append([]string{paths[0]}, rest...), ok
	}
	return nil, false
}

func walkStructKey(t reflect.Type, paths []string) ([]string, bool) {
	index, consumed, ok := matchField(t, paths)
	if !ok {
		// the flag adds a single item to a list, e.g. providers-git-url adds to urls
		i := fieldIndexByTag(t, strings.Join(paths, "-"), "explode")
		if i < 0 {
			return nil, false
		}
		index, consumed = []int{i}, len(paths)
	}

	field := t.FieldByIndex(index)
	name, ok := yamlName(field)
	if !ok {
		return nil, false
	}
	rest, ok := walkConfigKey(field.Type, paths[consumed:])
	return append([]string{name}, rest...), ok
}

// yamlName is the name of the field in the config file. ok is false for fields excluded from it.
func yamlName(f reflect.StructField) (name string, ok bool) {
	tag, _, _ := strings.Cut(f.Tag.Get("yaml"), ",")
	switch tag {
	case "-":
		return "", false
	case "":
		runes := []rune(f.Name)
		runes[0] = unicode.ToLower(runes[0])
		return string(runes), true
	}
	return tag, true
}

func isIndexSegment(path string) bool {
	return strings.HasPrefix(path, "[") || hasIndexPath(path)
}

package decoders

import (
	"fmt"
	"os"
	"strings"
)

const DefaultEnvNamePrefix = "MOKAPI_"

func parseFlags() (map[string]string, error) {
	flags, error := parseArgs(os.Args[1:]) // first argument is the program path
	if error != nil {
		return nil, error
	}

	envs, error := parseEnv(os.Environ())
	if error != nil {
		return nil, error
	}

	// merge maps. env flags overwrites cli flags
	for k, v := range envs {
		flags[k] = v
	}

	return flags, nil
}

func parseEnv(environ []string) (map[string]string, error) {
	dictionary := make(map[string]string)

	for _, s := range environ {
		kv := strings.SplitN(s, "=", 2)
		if strings.HasPrefix(strings.ToUpper(kv[0]), DefaultEnvNamePrefix) {
			key := strings.Replace(kv[0], DefaultEnvNamePrefix, "", 1)
			name := strings.ReplaceAll(strings.ToLower(key), "_", ".")
			dictionary[name] = kv[1]
		}
	}

	return dictionary, nil
}

func parseArgs(args []string) (map[string]string, error) {
	dictionary := make(map[string]string)
	for i := 0; i < len(args); i++ {
		s := args[i]
		if len(s) < 2 || s[0] != '-' {
			return nil, fmt.Errorf("Invalid argument %v", s)
		}

		index := 1

		if s[1] == '-' {
			index++
			if len(s) == 2 {
				return nil, fmt.Errorf("Invalid argument %v", s)
			}
		}

		name := s[index:]
		value := ""
		hasValue := false

		if len(name) == 0 || name[0] == '-' || name[0] == '=' {
			return nil, fmt.Errorf("Invalid argument %v", s)
		}

		// search for =
		for i := 1; i < len(name); i++ { // = can not be first
			if name[i] == '=' {
				value = name[i+1:]
				name = name[0:i]
				hasValue = true
			}
		}

		if hasValue {
			dictionary[strings.ToLower(name)] = value
			continue
		}

		// value is next arg
		i++
		if i >= len(args) {
			return nil, fmt.Errorf("argument %v need a value", name)
		}
		value = args[i]
		dictionary[strings.ToLower(name)] = value
	}

	return dictionary, nil
}

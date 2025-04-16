package decoders

import (
	"fmt"
	"os"
	"strings"
)

const DefaultEnvNamePrefix = "MOKAPI_"

func parseFlags() (map[string][]string, error) {
	flags, err := parseArgs(os.Args[1:]) // first argument is the program path
	if err != nil {
		return nil, err
	}

	envs := parseEnv(os.Environ())

	// merge maps. env flags overwrites cli flags
	for k, v := range envs {
		flags[k] = []string{v}
	}

	return flags, nil
}

func parseEnv(environ []string) map[string]string {
	dictionary := make(map[string]string)

	for _, s := range environ {
		kv := strings.SplitN(s, "=", 2)
		if strings.HasPrefix(strings.ToUpper(kv[0]), DefaultEnvNamePrefix) {
			key := strings.Replace(kv[0], DefaultEnvNamePrefix, "", 1)
			name := strings.ReplaceAll(strings.ToLower(key), "_", "-")
			dictionary[name] = kv[1]
		}
	}

	return dictionary
}

func parseArgs(args []string) (map[string][]string, error) {
	dictionary := make(map[string][]string)
	inPositionalArgs := false
	for i := 0; i < len(args); i++ {
		s := args[i]
		if len(s) < 2 || s[0] != '-' {
			dictionary["args"] = append(dictionary["args"], s)
			continue
		} else if inPositionalArgs {
			// currently, no positional argument with prefix -- are defined
			return nil, fmt.Errorf("unknown positional argument: '%s'", s)
		}

		index := 1

		if s[1] == '-' {
			index++
			if len(s) == 2 {
				inPositionalArgs = true
				continue
			}
		}

		name := s[index:]
		value := ""
		hasValue := false

		if len(name) == 0 || name[0] == '-' || name[0] == '=' {
			return nil, fmt.Errorf("invalid argument %v", s)
		}

		// search for =
		for i := 1; i < len(name); i++ { // = can not be first
			if name[i] == '=' {
				value = name[i+1:]
				name = name[0:i]
				hasValue = true
				break
			}
		}

		param := strings.ToLower(name)
		if hasValue {
			dictionary[param] = append(dictionary[param], value)
			continue
		}

		// value is next args
		for i++; i < len(args); i++ {
			if strings.HasPrefix(args[i], "--") {
				i--
				break
			}
			value = args[i]
			dictionary[param] = append(dictionary[param], value)
		}

		if len(dictionary[param]) == 0 {
			dictionary[param] = append(dictionary[param], "")
		}
	}

	return dictionary, nil
}

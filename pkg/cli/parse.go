package cli

import (
	"fmt"
	"os"
	"strings"
)

func parseFlags(args []string, envNamePrefix string, flags *FlagSet) ([]string, error) {
	// env vars
	if envNamePrefix != "" {
		for _, s := range os.Environ() {
			kv := strings.SplitN(s, "=", 2)
			if strings.HasPrefix(strings.ToUpper(kv[0]), envNamePrefix) {
				key := strings.Replace(kv[0], envNamePrefix, "", 1)
				name := strings.ReplaceAll(strings.ToLower(key), "_", "-")
				if err := flags.setValue(name, []string{kv[1]}, SourceEnv); err != nil {
					return nil, fmt.Errorf("unknown environment variable '%s' (value '%s')", kv[0], kv[1])
				}
			}
		}
	}

	// CLI args
	inPositionalArgs := false
	var positionalArgs []string
	for i := 0; i < len(args); i++ {
		s := args[i]
		if len(s) < 2 || s[0] != '-' {
			positionalArgs = append(positionalArgs, s)
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
			} else if name[i] == ' ' {
				value = name[i+1:]
				name = name[0:i]
				break
			}
		}

		param := strings.ToLower(name)
		if hasValue {
			if err := flags.setValue(param, []string{value}, SourceCli); err != nil {
				return nil, err
			}
			continue
		}

		// value is next args
		for i++; i < len(args); i++ {
			if strings.HasPrefix(args[i], "--") || strings.HasPrefix(args[i], "-") {
				i--
				break
			}
			value = args[i]
		}

		if err := flags.setValue(param, []string{value}, SourceCli); err != nil {
			return nil, err
		}
	}

	return positionalArgs, nil
}

package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"mokapi/config/decoders"
	"mokapi/config/static"
	"reflect"
	"strings"
)

func printHelp() {
	fmt.Print("\nMokapi is an easy, modern and flexible API mocking tool using Go and Javascript.\n")
	fmt.Print("\nUsage:\n  mokapi [flags]\n")
	fmt.Print("\nFlags:")
	fmt.Print("\n  --log-level (string)\n\tLog level (default: info)")
	fmt.Print("\n  --log-format string\n\tLog format (default: text)")
	fmt.Print("\n")

	fmt.Print("\n  --api-port (integer)\n\tApi port (default: 8080)")
	fmt.Print("\n  --api-dashboard | --api-no-dashboard (boolean)\n\tActivate dashboard (default: true)")
	fmt.Print("\n  --api-path (string)\n\tThe path prefix where dashboard is served (default: empty)")
	fmt.Print("\n  --api-base (string)\n\tThe base path of the dashboard useful in case of url rewriting (default: empty)")
	fmt.Print("\n")

	fmt.Print("\n  --providers-file-filename (string)\n\tLoad configuration from this URL")
	fmt.Print("\n  --providers-file-directory (string)\n\tLoad one or more dynamic configuration from a directory")
	fmt.Print("\n  --providers-file-skip-prefix (list)\n\tOne or more prefixes that indicate whether a file or directory should be skipped.")
	fmt.Print("\n\n\t(string)\n\t    The prefix of the files to skip")
	fmt.Print("\n  --providers-file-include (list)\n\tOne or more patterns that a file must match, except when empty")
	fmt.Print("\n\n\t(string)\n\t    The pattern that a file must match")
	fmt.Print("\n")

	fmt.Print("\n  --providers-http-url (string)\n\tLoad the dynamic configuration from file")
	fmt.Print("\n  --providers-http-poll-interval (string)\n\tLPolling interval for URL (default: 3m)")
	fmt.Print("\n  --providers-http-poll-timeout (string)\n\tPolling timeout for URL (default is 5s)")
	fmt.Print("\n  --providers-http-proxy (string)\n\tSpecifies a proxy server for the request")
	fmt.Print("\n  --providers-http-tls-skip-verify (boolean)\n\tSpecifies a proxy server for the request")
	fmt.Print("\n  --providers-http-ca (string)\n\tPath to the certificate authority used for secure connection (default: system certification pool)")
	fmt.Print("\n")

	fmt.Print("\n  --providers-git-url (string)\n\tLoad one or more dynamic configuration from a GIT repository")
	fmt.Print("\n  --providers-git-pull-interval (string)\n\tPulling interval for URL (default: 3m)")
	fmt.Print("\n  --providers-git-temp-dir (string)\n\tSpecifies the folder to checkout all GIT repositories")
	fmt.Print("\n")

	fmt.Print("\n  --providers-npm-package (string)\n\tLoad one or more dynamic configuration from a GIT repository")
	fmt.Print("\n")

	fmt.Print("\n  --generate-cli-skeleton [string]\n\tGenerates the skeleton configuration file")

	fmt.Print("\n\nGet help with Mokapi CLI: https://mokapi.io/docs/configuration/static/cli")
}

func writeSkeleton(cfg *static.Config) {
	var skeleton interface{}
	if s, ok := cfg.GenerateSkeleton.(string); ok {
		paths := decoders.ParsePath(s)
		current := reflect.ValueOf(static.NewConfig())
		for _, path := range paths {
			if current.Kind() == reflect.Pointer {
				current = current.Elem()
			}
			field := current.FieldByNameFunc(func(f string) bool {
				return strings.ToLower(f) == path
			})
			if !field.IsValid() {
				log.Errorf("unable to find config element: %v", cfg.GenerateSkeleton.(string))
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

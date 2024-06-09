package main

import "fmt"

func printHelp() {
	fmt.Printf("\nMokapi is an easy, modern and flexible API mocking tool using Go and Javascript.\n")
	fmt.Printf("\nUsage:\n  mokapi [flags]\n")
	fmt.Printf("\nFlags:")
	fmt.Printf("\n  --log-level (string)\n\tLog level (default: info)")
	fmt.Printf("\n  --log-format string\n\tLog format (default: text)")
	fmt.Printf("\n")

	fmt.Printf("\n  --api-port (integer)\n\tApi port (default: 8080)")
	fmt.Printf("\n  --api-dashboard | --api-no-dashboard (boolean)\n\tActivate dashboard (default: true)")
	fmt.Printf("\n  --api-path (string)\n\tThe path prefix where dashboard is served (default: empty)")
	fmt.Printf("\n  --api-base (string)\n\tThe base path of the dashboard useful in case of url rewriting (default: empty)")
	fmt.Printf("\n")

	fmt.Printf("\n  --providers-file-filename (string)\n\tLoad configuration from this URL")
	fmt.Printf("\n  --providers-file-directory (string)\n\tLoad one or more dynamic configuration from a directory")
	fmt.Printf("\n  --providers-file-skip-prefix (list)\n\tOne or more prefixes that indicate whether a file or directory should be skipped.")
	fmt.Printf("\n\n\t(string)\n\t    The prefix of the files to skip")
	fmt.Printf("\n  --providers-file-include (list)\n\tOne or more patterns that a file must match, except when empty")
	fmt.Printf("\n\n\t(string)\n\t    The pattern that a file must match")
	fmt.Printf("\n")

	fmt.Printf("\n  --providers-http-url (string)\n\tLoad the dynamic configuration from file")
	fmt.Printf("\n  --providers-http-poll-interval (string)\n\tLPolling interval for URL (default: 5s)")
	fmt.Printf("\n  --providers-http-poll-timeout (string)\n\tPolling timeout for URL (default is 5s)")
	fmt.Printf("\n  --providers-http-proxy (string)\n\tSpecifies a proxy server for the request")
	fmt.Printf("\n  --providers-http-tls-skip-verify (boolean)\n\tSpecifies a proxy server for the request")
	fmt.Printf("\n  --providers-http-ca (string)\n\tPath to the certificate authority used for secure connection (default: system certification pool)")
	fmt.Printf("\n")

	fmt.Printf("\n  --providers-git-url (string)\n\tLoad one or more dynamic configuration from a GIT repository")
	fmt.Printf("\n  --providers-git-pull-interval (string)\n\tPulling interval for URL (default: 5s)")
	fmt.Printf("\n  --providers-git-temp-dir (string)\n\tSpecifies the folder to checkout all GIT repositories")
	fmt.Printf("\n")

	fmt.Printf("\n  --providers-npm-package (string)\n\tLoad one or more dynamic configuration from a GIT repository")
}

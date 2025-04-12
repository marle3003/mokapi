---
title: Complete Options Reference
description: A complete list of all Mokapi options and how to set the option in your config files, environment variables, or CLI.
---
# Complete Options Reference

Options define Mokapi's run behavior that can be passed in multiple places.
Mokapi chooses the value from the [highest order of precedence](/docs/configuration/introduction.md).

## Log
Mokapi log level (default is info)
```bash tab=CLI
--log-level=warn
```
```bash tab=Env
MOKAPI_LOG_LEVEL=warn
```
```yaml tab=File (YAML)
log:
  level: warn
```

Mokapi log format: json|text (default is text)
```bash tab=CLI
--log-format=json
```
```bash tab=Env
MOKAPI_LOG_FORMAT=json
```
```yaml tab=File (YAML)
log:
  format: json
```

## API & Dashboard
API port (Default 8080). The API is available on the path `/api`
```bash tab=CLI
--api-port 5000
```
```bash tab=Env
MOKAPI_API_PORT=5000
```
```yaml tab=File (YAML)
api:
  port: 5000
```

Activate dashboard (default true). The dashboard is available at the same port as the API but on the path `/` by default.
```bash tab=CLI
--api-dashboard true
--api-dashboard
--api-no-dashboard
```
```bash tab=Env
MOKAPI_API_DASHBOARD=true
```
```yaml tab=File (YAML)
api:
  dashboard: true
```

The path prefix where dashboard is served (default empty)
```bash tab=CLI
--api-path /mokapi/dashboard
```
```bash tab=Env
MOKAPI_API_PATH=/mokapi/dashboard
```
```yaml tab=File (YAML)
api:
  path: /mokapi/dashboard
```

The base path of the dashboard useful in case of url rewriting (default empty)
```bash tab=CLI
--api-base /mokapi/dashboard
```
```bash tab=Env
MOKAPI_API_BASE=/mokapi/dashboard
```
```yaml tab=File (YAML)
api:
  base: /mokapi/dashboard
```

## File Provider
Load the dynamic configuration from file
```bash tab=CLI
--providers-file-filename foobar.yaml
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_FILENAME=foobar.yaml
```
```yaml tab=File (YAML)
providers:
  file:
    filename: foobar.yaml
```

Load one or more dynamic configuration from a directory
```bash tab=CLI
--providers-file-directory /foobar
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_DIRECTORY=/foobar
```
```yaml tab=File (YAML)
providers:
  file:
    directory: /foobar
```

One or more prefixes that indicate whether a file or directory should be skipped. (default is ["_"])
```bash tab=CLI
--providers-file-skip-prefix foo
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_SKIP_PREFIX=foo
```
```yaml tab=File (YAML)
providers:
  file:
    skipPrefix: foo
```

One or more patterns that a file must match, except when empty.
```bash tab=CLI
--providers-file-include *.json *.yaml
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_INCLUDE="*.json *.yaml"
```
```yaml tab=File (YAML)
providers:
  file:
    include:
      - "*.json"
      - "*.yaml"
```

## HTTP Provider
Load configuration from this URL
```bash tab=CLI
--providers-http-url https://foo.bar/file.yaml
--providers-http-url https://foo.bar/file1.yaml --providers-http-url https://foo.bar/file2.yaml
--providers-http-urls https://foo.bar/file1.yaml https://foo.bar/file2.yaml
--providers-http-urls '["https://foo.bar/file1.yaml","https://foo.bar/file2.yaml"]'
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_URL=https://foo.bar/file.yaml
```
```yaml tab=File (YAML)
providers:
  http:
    url: https://foo.bar/file.yaml
    urls:
      - https://foo.bar/file2.yaml
```

Polling interval for URL (default is 3m)
```bash tab=CLI
--providers-http-poll-interval 10s
--providers-http-poll-interval 5m
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_POLL_INTERVAL=10s
```
```yaml tab=File (YAML)
providers:
  http:
    pollInterval: 10s
```

Polling timeout for URL (default is 5s)
```bash tab=CLI
--providers-http-poll-timeout 10s
--providers-http-poll-timeout 5m
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_POLL_TIMEOUT=10s
```
```yaml tab=File (YAML)
providers:
  http:
    polTimeout: 10s
```

Specifies a proxy server for the request, rather than connecting directly to the URL.
```bash tab=CLI
--providers-http-proxy http://localhost:3128
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_PROXY=http://localhost:3128
```
```yaml tab=File (YAML)
providers:
  http:
    proxy: http://localhost:3128
```

Skip certificate validation checks.
```bash tab=CLI
--providers-http-tls-skip-verify true
--providers-http-tls-skip-verify
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_TLS_SKIP_VERIFY=true
```
```yaml tab=File (YAML)
providers:
  http:
    tlsSkipVerify: true
```

Path to the certificate authority used for secure connection. By default, the system
certification pool is used.
```bash tab=CLI
--providers-http-ca /path/to/mycert.pem
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_CA=/path/to/mycert.pem
```
```yaml tab=File (YAML)
providers:
  http:
    ca: /path/to/mycert.pem
```

## GIT Provider
Load one or more dynamic configuration from a GIT repository
```bash tab=CLI
--providers-git-url=https://github.com/foo/foo.git
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_URL=https://github.com/foo/foo.git
```
```yaml tab=File (YAML)
providers:
  git:
    url: https://github.com/foo/foo.git
```

Pulling interval for URL in seconds (default 3m)
```bash tab=CLI
--providers-git-pull-interval=10s
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_PULL_IINTERVAL=10s
```
```yaml tab=File (YAML)
providers:
  git:
    pullInterval: 10s
```

Pulling interval for URL in seconds (default 3m)
```bash tab=CLI
--providers-git-pul-interval=10s
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_PULL_IINTERVAL=10s
```
```yaml tab=File (YAML)
providers:
  git:
    pullInterval: 10s
```

Specifies the folder to check out all GIT repositories (default uses the default directory for temporary files)
```bash tab=CLI
--providers-git-temp-dir=/tempdir
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_TEMP_DIR=/tempdir
```
```yaml tab=File (YAML)
providers:
  git:
    tempDir: /tempdir
```

## NPM Provider

Specifies NPM package that Mokapi looks for.
```bash tab=CLI
--providers-npm-package=foo
```
```bash tab=Env
MOKAPI_PROVIDERS_NPM_PACKAGE=foo
```
```yaml tab=File (YAML)
providers:
  npm:
    package: foo
```

## Certificates
CA Certificate for signing certificate generated at runtime
```bash tab=CLI
--rootCaCert=/path/to/caCert.pem
```
```bash tab=Env
MOKAPI_RootCaCert=/path/to/caCert.pem
```
```yaml tab=File (YAML)
providers:
  rootCaCert: /path/to/caCert.pem
```

Private Key of CA for signing certificate generated at runtime
```bash tab=CLI
--rootCaKey=/path/to/caKey.pem
```
```bash tab=Env
MOKAPI_RootCaKey=/path/to/caKey.pem
```
```yaml tab=File (YAML)
providers:
  rootCaKey: /path/to/caKey.pem
```

## Events Store

Mokapi stores request and event history in memory. You can control the memory size (in number of events) for 
each event store using the following CLI flags.

Sets the default maximum number of events stored for each event type (e.g., HTTP, Kafka), unless overridden individually.
(default 100)
```bash tab=CLI
--event-store-default-size 200
```
```bash tab=Env
MOKAPI_EVENT_STORE_DEFAULT=200
```
```yaml tab=File (YAML)
event:
  store:
    default: 200
```

Overrides the default event store size for a specific API by name.
```bash tab=CLI
--event-store-<name>-size 200
```
```bash tab=Env
MOKAPI_EVENT_STORE_<NAME>=200
```
```yaml tab=File (YAML)
event:
  store:
    <name>: 200
```

Provides advanced configuration for a specific API using a JSON-style syntax.
This format is required if the API name contains spaces or special characters.
```bash tab=CLI
--event-store <api-name>={"size": 250}
```
```bash tab=Env
MOKAPI_EVENT_STORE=<api-name>={"size": 250}
```
```yaml tab=File (YAML)
event:
  store:
    "<api-name>": 250
```
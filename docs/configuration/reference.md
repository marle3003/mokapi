# Static Configuration

## Log
Mokapi log level (Default info)
```bash tab=CLI
--providers.log.level=warn
```
```bash tab=Env
MOKAPI_Log_Level=warn
```
```yaml tab=File (YAML)
providers:
  log:
    level: warn
```

Mokapi log format: json|default (Default default)
```bash tab=CLI
--providers.log.format=json
```
```bash tab=Env
MOKAPI_Log_Formatl=json
```
```yaml tab=File (YAML)
providers:
  log:
    format: json
```

## API & Dashboard
API port (Default 8080)
```bash tab=CLI
--providers.api.port=5000
```
```bash tab=Env
MOKAPI_API_Port=5000
```
```yaml tab=File (YAML)
providers:
  api:
    port: 5000
```

Activate dashboard (Default true)
```bash tab=CLI
--providers.api.dashboard=true
```
```bash tab=Env
MOKAPI_API_Dashboard=true
```
```yaml tab=File (YAML)
providers:
  api:
    dashboard: true
```

The path prefix where dashboard is served (Default empty)
```bash tab=CLI
--providers.api.path=/mokapi/dashboard
```
```bash tab=Env
MOKAPI_API_Path=/mokapi/dashboard
```
```yaml tab=File (YAML)
providers:
  api:
    path: /mokapi/dashboard
```

The base path of the dashboard useful in case of url rewriting (Default empty)
```bash tab=CLI
--providers.api.base=/mokapi/dashboard
```
```bash tab=Env
MOKAPI_API_Base=/mokapi/dashboard
```
```yaml tab=File (YAML)
providers:
  api:
    base: /mokapi/dashboard
```

## File Provider
Load the dynamic configuration from file
```bash tab=CLI
--providers.file.filename=foobar.yaml
```
```bash tab=Env
MOKAPI_Providers_File_Filename=foobar.yaml
```
```yaml tab=File (YAML)
providers:
  file:
    filename: foobar.yaml
```

Load one or more dynamic configuration from a directory
```bash tab=CLI
--providers.file.directory=/foobar
```
```bash tab=Env
MOKAPI_Providers_File_Filename=/foobar
```
```yaml tab=File (YAML)
providers:
  file:
    filename: /foobar
```

## HTTP Provider
Load configuration from this URL
```bash tab=CLI
--providers.http.url=http://foo.bar/file.yaml
```
```bash tab=Env
MOKAPI_Providers_HTTP_URL=http://foo.bar/file.yaml
```
```yaml tab=File (YAML)
providers:
  http:
    url: http://foo.bar/file.yaml
```

Polling interval for URL in seconds (default 5)
```bash tab=CLI
--providers.http.pollInterval=10
```
```bash tab=Env
MOKAPI_Providers_HTTP_PollInterval=10
```
```yaml tab=File (YAML)
providers:
  http:
    pollInterval: 10
```

Specifies a proxy server for the request, rather than connecting directly to the URL.
```bash tab=CLI
--providers.http.proxy=http://localhost:3128
```
```bash tab=Env
MOKAPI_Providers_HTTP_Proxy=http://localhost:3128
```
```yaml tab=File (YAML)
providers:
  http:
    proxy: http://localhost:3128
```

Skip certificate validation checks.
```bash tab=CLI
--providers.http.tlsSkipVerify=true
```
```bash tab=Env
MOKAPI_Providers_HTTP_TlsSkipVerify=true
```
```yaml tab=File (YAML)
providers:
  http:
    tlsSkipVerify: true
```

## GIT Provider
Load one or more dynamic configuration from a GIT repository
```bash tab=CLI
--providers.git.url=https://github.com/foo/foo.git
```
```bash tab=Env
MOKAPI_Providers_Git_URL=https://github.com/foo/foo.git
```
```yaml tab=File (YAML)
providers:
  git:
    url: https://github.com/foo/foo.git
```

Pulling interval for URL in seconds (default 5)
```bash tab=CLI
--providers.git.pullInterval=10
```
```bash tab=Env
MOKAPI_Providers_Git_PullInterval=10
```
```yaml tab=File (YAML)
providers:
  git:
    pullInterval: 10
```

## Certificates
CA Certificate for signing certificate generated at runtime
```bash tab=CLI
--providers.git.rootCaCert=/path/to/caCert.pem
```
```bash tab=Env
MOKAPI_Providers_RootCaCert=/path/to/caCert.pem
```
```yaml tab=File (YAML)
providers:
  rootCaCert: /path/to/caCert.pem
```

Private Key of CA for signing certificate generated at runtime
```bash tab=CLI
--providers.rootCaKey=/path/to/caKey.pem
```
```bash tab=Env
MOKAPI_Providers_RootCaKey=/path/to/caKey.pem
```
```yaml tab=File (YAML)
providers:
  rootCaKey: /path/to/caKey.pem
```
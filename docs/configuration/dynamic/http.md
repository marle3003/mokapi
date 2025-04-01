---
title: HTTP Provider for dynamic configuration
description: Reads your dynamic configuration from an HTTP(S) source.
---
# HTTP

Reads your dynamic configuration from an HTTP(S) source. The HTTP provider downloads your
configuration, which is then read by the [File Provider](/docs/configuration/dynamic/file.md).
By default, it uses HTTP proxies as directed by the environment variables HTTP_PROXY, HTTPS_PROXY
and NO_PROXY (or the lowercase versions thereof).

## Configuration Example

```yaml tab=File (YAML)
providers:
  http:
    url: http://127.0.0.1/api
```
```bash tab=CLI
--providers-file-directory http://127.0.0.1/api
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_DIRECTORY=http://127.0.0.1/api
```

## Provider Configuration

A list of all options that can be used with the HTTP provider, refer to
the [reference page](/docs/configuration/reference.md).

``` box=tip
You can also use CLI JSON or shorthand syntax, see [CLI](/docs/configuration/static/cli.md)
```

``` box=tip
HTTP provider is also used to get resources defined by `$ref` using HTTP scheme.
For example in OpenAPI 3.0 "$ref: 'http://path/to/your/resource.json#/myElement'"
```

### URL
Defines the URL where to get your configuration.

```yaml tab=File (YAML)
providers:
  http:
    url: http://127.0.0.1/api
```
```bash tab=CLI
--providers-http-url "http://127.0.0.1/api"
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_URL=http://127.0.0.1/api
```

### Poll Interval
Defines in which interval possible changes are checked, default 3 minutes. 
Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

```yaml tab=File (YAML)
providers:
  http:
    pollInterval: 3m30s
```
```bash tab=CLI
--providers-http-poll-interval 3m30s
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_POLL_INTERVAL=3m30s
```

### Poll Timeout
Defines the polling timeout, default 5 seconds.
Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

```yaml tab=File (YAML)
providers:
  http:
    pollTimeout: 10s
```
```bash tab=CLI
--providers-http-poll-timeout 10s
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_POLL_TIMEOUT=10s
```

### Proxy
Specifies a proxy server for the request, rather than downloading directly from the
resource. 

```yaml tab=File (YAML)
providers:
  http:
    proxy: http://localhost:3128
```
```bash tab=CLI
--providers-http-proxy http://localhost:3128
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_PROXY=http://localhost:3128
```

### CA (Certificate Authority)
Path to the certificate authority used for secure connection. By default, the system 
certification pool is used.

``` box=warning
This switch is only intended to be used against known hosts using self-signed certificate.
Use at your own risk.
```

```yaml tab=File (YAML)
providers:
  http:
    ca: /path/to/mycert.pem
```
```bash tab=CLI
--providers-http-ca=/path/to/mycert.pem
```
```bash tab=Env
MOKAPI_Providers_HTTP_CA=/path/to/mycert.pem
```

### Skip TLS Verification
Skips certificate validation checks. This includes all validations such as expiration,
trusted root authority, revocation, etc. Default is `false`

``` box=warning
This switch is only intended to be used against known hosts using self-signed certificate.
Use at your own risk.
```

```yaml tab=File (YAML)
providers:
  http:
    tlsSkipVerify: true
```
```bash tab=CLI
--providers-http-tls-skip-verify
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_TLS_SKIP_VERIFY=true
```
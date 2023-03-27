# HTTP

Reads your dynamic configuration from an HTTP(S) source. The HTTP provider downloads yours
configuration and is read by the [File Provider](/docs/configuration/file.md).

## Configuration

A list of all options that can be used with the HTTP provider, refer to
the [reference page](/docs/references/static-configuration.md).

``` box=tip
HTTP provider is also used to get resources defined by `$ref` using HTTP scheme.
For example in OpenAPI 3.0 "$ref: 'http://path/to/your/resource.json#/myElement'"
```

### URL
Defines the URL where to get your configuration.

```bash tab=CLI
--providers.http.url="http://foo.bar/api"
```
```bash tab=Env
MOKAPI_Providers_HTTP_URL=http://foo.bar/api
```
```yaml tab=File (YAML)
providers:
  http:
    url: http://foo.bar/api
```

### Poll Interval
Defines in which interval possible changes are checked, default 5 seconds. 
Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

```bash tab=CLI
--providers.http.pollInterval="2h45m"
```
```bash tab=Env
MOKAPI_Providers_HTTP_PollInterval=2h45m
```
```yaml tab=File (YAML)
providers:
  http:
    pollInterval: 2h45m
```

### Poll Timeout
Defines the polling timeout, default 5 seconds.
Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

```bash tab=CLI
--providers.http.pollTimeout="10s"
```
```bash tab=Env
MOKAPI_Providers_HTTP_PollTimeout=10s
```
```yaml tab=File (YAML)
providers:
  http:
    pollTimeout: 10s
```

### Proxy
Specifies a proxy server for the request, rather than downloading directly from the
resource.

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

### Skip TLS Verification
Skips certificate validation checks. This includes all validations such as expiration,
trusted root authority, revocation, etc. Default is `false`

``` box=warning
This switch is only intended to be used against known hosts using self-signed certificate.
Use at your own risk.
```

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
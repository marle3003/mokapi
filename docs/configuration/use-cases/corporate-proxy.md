---
title: Running Mokapi Behind a Corporate Proxy
description: Fetch specs over HTTP(S) through a proxy and trust an internal certificate authority, with the proxy, CA and TLS skip-verify options of the HTTP provider.
---
# Running Mokapi Behind a Corporate Proxy

Mokapi loads configuration from a URL, for example an OpenAPI spec on an internal server or a `$ref` that
points to an `http(s)` address, with the [HTTP Provider](/docs/configuration/dynamic/http.md). Behind a
corporate proxy this can fail for two reasons: the request must go through the proxy, and the proxy may
re-sign HTTPS traffic with an internal certificate that Mokapi doesn't trust.

This page covers the HTTP provider only. Local files and mocked APIs are not affected, and Mokapi doesn't
need internet access to serve a mock. Requests sent by scripts with the `mokapi/http` module use the proxy
environment variables, but none of the options below.

## Use the proxy from your environment

If `HTTP_PROXY`, `HTTPS_PROXY` and `NO_PROXY` (or the lowercase variants) are set, the HTTP provider uses them
without further configuration.

```bash
export HTTPS_PROXY=http://proxy.example.com:3128
export NO_PROXY=localhost,127.0.0.1
mokapi https://specs.example.com/petstore.yaml
```

The [GIT provider](/docs/configuration/dynamic/git.md) uses the same environment variables for repositories
cloned over `http(s)`. It has no proxy option of its own.

## Set the proxy explicitly

Use [`--providers-http-proxy`](/docs/configuration/static/mokapi.md#providers-http-proxy) if you don't want to
change the environment. When it is set, it replaces the environment variables for the HTTP provider, so
`NO_PROXY` no longer applies to it.

```yaml tab=File
providers:
  http:
    urls: [https://specs.example.com/petstore.yaml]
    proxy: http://proxy.example.com:3128
```
```bash tab=CLI
--providers-http-url https://specs.example.com/petstore.yaml \
--providers-http-proxy http://proxy.example.com:3128
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_URL=https://specs.example.com/petstore.yaml
MOKAPI_PROVIDERS_HTTP_PROXY=http://proxy.example.com:3128
```

If the proxy URL can't be parsed, Mokapi logs `invalid proxy url` and falls back to the environment
variables.

## Trust the certificate of the proxy

A proxy that inspects HTTPS traffic presents certificates signed by your company's certificate authority.
If that CA isn't in the system certificate pool, requests fail with a certificate error. The cleanest fix is to
install the CA in the system pool. Otherwise pass it to Mokapi with
[`--providers-http-ca`](/docs/configuration/static/mokapi.md#providers-http-ca). Mokapi adds it to the system
pool.

```yaml tab=File
providers:
  http:
    urls: [https://specs.example.com/petstore.yaml]
    proxy: http://proxy.example.com:3128
    ca: /etc/ssl/corp-ca.pem
```
```bash tab=CLI
--providers-http-url https://specs.example.com/petstore.yaml \
--providers-http-proxy http://proxy.example.com:3128 \
--providers-http-ca /etc/ssl/corp-ca.pem
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_URL=https://specs.example.com/petstore.yaml
MOKAPI_PROVIDERS_HTTP_PROXY=http://proxy.example.com:3128
MOKAPI_PROVIDERS_HTTP_CA=/etc/ssl/corp-ca.pem
```

The file can be PEM or DER. A PEM file may contain several certificates, so you can pass a bundle with the
intermediate and root CA of your proxy.

If the CA can't be read or parsed, Mokapi logs `failed to use CA certification for http provider` and continues
with the system pool only.

## Skip certificate verification

For a quick test you can turn verification off with
[`--providers-http-tls-skip-verify`](/docs/configuration/static/mokapi.md#providers-http-tls-skip-verify).

```yaml tab=File
providers:
  http:
    urls: [https://specs.example.com/petstore.yaml]
    tlsSkipVerify: true
```
```bash tab=CLI
--providers-http-url https://specs.example.com/petstore.yaml \
--providers-http-tls-skip-verify
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_URL=https://specs.example.com/petstore.yaml
MOKAPI_PROVIDERS_HTTP_TLS_SKIP_VERIFY=true
```

``` box=warning
Skipping verification disables all certificate checks, including expiry and the trusted authority, and takes
precedence over `providers-http-ca`. Use it only against hosts you know.
```

## Not related: root-ca-cert

[`--root-ca-cert`](/docs/configuration/static/mokapi.md#root-ca-cert) and `--root-ca-key` configure the CA
that Mokapi uses to sign the certificates of its own mocked services. They have no effect on outgoing
requests through a proxy.

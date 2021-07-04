# HTTPS and TLS
Mokapi is able to mock the behaviour of multiple hostnames and present a valid X.509 Certificate
for them. This guide shows you how to configure TLS.

```yaml
openapi: 3.0.0
servers:
  - url: https://demo
  - url: https://demo:8443
```

## Certificate Authority
By default, Mokapi signs all generated certificate with its own CA certificate in the [git repo](https://github.com/marle3003/mokapi/tree/master/assets).

This example demonstrates how to reference a custom CA certificate as an environment variable.
```
MOKAPI_RootCaCert: /path/to/caCert.pem
MOKAPI_RootCaKey: /path/to/caKey.pem
```

## Server Certificate
You can set your custom server certificate or let Mokapi generate it on runtime. Mokapi
generates a new certificate at first request, if you not provide a suitable certificate.

```yaml
mokapi: 1.0
certificates:
  - certFile: ./domainCert.pem
    keyFile: ./domainKey.pem
  - certFile: /path/to/other-domain.cert
    keyFile: /path/to/other-domain.key
```
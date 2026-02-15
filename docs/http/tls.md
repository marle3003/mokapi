---
title: HTTPS, TLS and Certificates
description: Mokapi can mock the behaviour of multiple hostnames and present valid X.509 certificates for them. This guide explains how to configure TLS in Mokapi.
---
# HTTPS, TLS and Certificates
Mokapi can simulate multiple hostnames and present a valid X.509 certificate for each of
them, making it possible to test HTTPS endpoints with realistic TLS behavior. 
This guide explains how TLS works in Mokapi and how you can configure certificates to suit 
your needs.

## Example: Mocking HTTPS Endpoints

When defining your API, you can specify HTTPS URLs in the `servers` section. Mokapi will
automatically generate certificates for these hostnames at runtime if none are provided.

```yaml
openapi: 3.0.0
info:
  title: Petstore
  version: "1"
servers:
  - url: https://demo
  - url: https://demo:8443
paths: {}
```

In this example:
- Mokapi will respond over HTTPS for https://demo and https://demo:8443.
- If no certificate is provided, Mokapi will generate one signed by its own Certificate Authority.

## Certificate Authority (CA)
By default, Mokapi uses its own built-in Certificate Authority to sign any certificates it
generates. The Root CA certificate and key are stored in the [Mokapi GitHub repository](https://github.com/marle3003/mokapi/tree/master/assets).

If you want to use your own CA for signing, you can specify it, for example, by setting environment variables.
```yaml
MOKAPI_RootCaCert: /path/to/caCert.pem
MOKAPI_RootCaKey: /path/to/caKey.pem
```

Why use your own CA?
- To avoid browser or client certificate warnings (if your CA is trusted by the OS or application). 
- To align with an existing internal PKI (Public Key Infrastructure). 
- To test certificate chains with your organization’s security policies.

## Server Certificates

Mokapi supports two ways of providing server certificates:

### Option A — Let Mokapi Generate Certificates

If you do not supply a certificate for a hostname, Mokapi will generate one dynamically the first
time that hostname is requested. The certificate will be valid for the requested hostname and 
signed by the configured CA (Mokapi’s default or your custom CA).

### Option B — Provide Your Own Certificates

If you already have certificates, you can configure Mokapi to use them instead of generating new ones.
Add them to the static configuration:

```yaml
certificates:
  static:
    - cert: ./domainCert.pem
      key: ./domainKey.pem
    - cert: /path/to/other-domain.cert
      key: /path/to/other-domain.key
```

You can list multiple certificate/key pairs to support multiple hostnames.

## Tips for Working with Certificates

- **Trust the CA** — If you use Mokapi’s default CA, importing its root certificate into your OS or browser’s trust store will prevent certificate warnings. 
- **Match Hostnames** — The certificate’s Common Name (CN) or Subject Alternative Name (SAN) must match the hostname in your API definition (servers list). 
- **Use Custom Domains** — For local testing, you can map custom domains in your /etc/hosts file or Windows hosts file to 127.0.0.1 so the hostname matches the certificate.

## Summary

- Mokapi can generate certificates dynamically or use certificates you provide. 
- Certificates are signed by Mokapi’s default CA unless you specify a custom one. 
- You can configure both the CA and the server certificates via environment variables or the static configuration. 
- Trusting the CA in your OS avoids browser or client security warnings.
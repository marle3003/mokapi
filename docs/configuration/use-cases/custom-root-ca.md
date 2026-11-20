---
title: Use Your Own Root CA for HTTPS
description: Sign the certificates Mokapi generates for HTTPS mocks with your own root CA, so browsers, curl, Node.js and test runners trust them without warnings.
---
# Use Your Own Root CA for HTTPS

When an API in Mokapi uses an `https://` server URL, Mokapi generates a certificate for the requested hostname
on the fly. By default it signs these certificates with its built-in CA, which no client trusts. Browsers show a
warning, and curl, Node.js, Java or your test runner fail with a certificate error.

If you sign with a root CA that your clients already trust, or that you add to their trust store once, every
generated certificate is trusted, whatever hostnames your mocks use. See
[HTTPS, TLS and Certificates](/docs/http/tls.md) for how certificates are generated and how to provide
certificates for individual hostnames instead.

## Create a CA

You need a CA certificate and its private key, both PEM encoded. The following commands create a CA that is valid
for ten years.

```bash
openssl genrsa -traditional -out ca-key.pem 2048
openssl req -x509 -new -key ca-key.pem -sha256 -days 3650 \
  -subj "/CN=My Mokapi Test CA" \
  -addext "basicConstraints=critical,CA:TRUE" \
  -addext "keyUsage=critical,keyCertSign,cRLSign" \
  -out ca.pem
```

Mokapi reads the key as an RSA key in PKCS#1 format. The file must start with
`-----BEGIN RSA PRIVATE KEY-----`. Since OpenSSL 3, `openssl genrsa` writes PKCS#8
(`BEGIN PRIVATE KEY`) unless you pass `-traditional`. EC keys are not supported for the root CA. If you already
have an RSA key in PKCS#8 format, convert it with `openssl rsa -in key.pem -traditional -out ca-key.pem`.
The `-traditional` option needs OpenSSL 3. LibreSSL, which macOS ships, doesn't know it but writes PKCS#1 by
default, so omit the option there.

If your company already provides a root or intermediate CA for development, use that certificate and key instead.

## Start Mokapi with the CA

Pass the certificate and the key with [`--root-ca-cert`](/docs/configuration/static/mokapi.md#root-ca-cert) and
[`--root-ca-key`](/docs/configuration/static/mokapi.md#root-ca-key).

```yaml tab=File
rootCaCert: ./ca.pem
rootCaKey: ./ca-key.pem
```
```bash tab=CLI
--root-ca-cert ./ca.pem \
--root-ca-key ./ca-key.pem
```
```bash tab=Env
MOKAPI_ROOT_CA_CERT=./ca.pem
MOKAPI_ROOT_CA_KEY=./ca-key.pem
```

The value is a file path or the PEM content itself. Inline content is handy for environment variables in a CI
pipeline where secrets are injected as text.

Set both options. If only one of them is set, Mokapi ignores it and signs with its built-in CA. A path that
doesn't exist is treated as inline PEM content and fails to parse, so check the path if Mokapi doesn't start.

## Trust the CA on the client

Your clients must trust `ca.pem`. Only the certificate is needed on the client side. The private key stays with
Mokapi.

```bash tab=curl
curl --cacert ca.pem https://demo:8443/pets
```
```bash tab=Node.js
NODE_EXTRA_CA_CERTS=./ca.pem node client.js
```
```bash tab=Docker
docker run --rm -v "$(pwd)/ca.pem:/certs/ca.pem:ro" \
  -e NODE_EXTRA_CA_CERTS=/certs/ca.pem my-test-image
```

For browsers and tools that use the operating system, import `ca.pem` into the system trust store: Keychain
Access on macOS, `update-ca-certificates` on Debian and Ubuntu, or the Certificate Manager (`certmgr.msc`) on
Windows. Java uses its own store. Add the CA with `keytool -importcert -trustcacerts -file ca.pem -alias
mokapi-ca -keystore $JAVA_HOME/lib/security/cacerts`.

## Verify

Use an OpenAPI spec with an HTTPS server URL. The hostname must resolve to the machine where Mokapi runs, for
example through an entry in your `hosts` file.

```yaml
openapi: 3.0.0
info:
  title: Petstore
  version: "1"
servers:
  - url: https://demo:8443
paths:
  /pets:
    get:
      responses:
        "200":
          description: OK
```

```bash
mokapi --root-ca-cert ./ca.pem --root-ca-key ./ca-key.pem ./petstore.yaml
curl --cacert ca.pem https://demo:8443/pets
```

To check who signed the certificate, inspect the chain.

```bash
openssl s_client -connect demo:8443 -servername demo -CAfile ca.pem </dev/null | grep -E "issuer|Verify"
```

The output shows `My Mokapi Test CA` as the issuer and `Verify return code: 0 (ok)`.

Mokapi caches a generated certificate per hostname for the lifetime of the process. Restart Mokapi after you
change the CA.

## Keep the CA key private

Anyone with the CA key can create certificates that every client with your CA in its trust store accepts. Use a
CA created for testing, don't commit `ca-key.pem` to a repository, and remove the CA from your trust store when
you no longer need it. Don't install it on machines that are used for anything besides development and tests.

## Trusting a CA when fetching configuration

[`--providers-http-ca`](/docs/configuration/static/mokapi.md#providers-http-ca) is the opposite direction. It
tells Mokapi which CA to trust when it downloads configuration over HTTPS. See
[Running Mokapi Behind a Corporate Proxy](/docs/configuration/use-cases/corporate-proxy.md).

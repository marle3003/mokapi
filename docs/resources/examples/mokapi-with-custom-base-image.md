---
title: Mokapi with custom base image
description: Build your own Docker image for Mokapi on a base image of your choice, for example Ubuntu with curl for a health check, and verify that it works.
icon: bi-box-seam-fill
tech: core
---

# Use Mokapi with a custom base Docker image

The official `mokapi/mokapi` image is built on Alpine and contains only the `mokapi` binary. For most setups that is
all you need, see [Running Mokapi in Docker and Docker Compose](/docs/configuration/use-cases/docker.md). A custom
image makes sense when the official one doesn't fit:

- Your company only allows images based on an approved base image.
- You need tools next to Mokapi. The `ubuntu` tag ships neither `wget` nor `curl`, so a Compose health check can't
  run there. The Alpine-based tags (`latest` and version tags) include BusyBox `wget` and don't need this.
- You want to add certificates, for example a company root CA that your scripts need to trust when they call
  other HTTPS services.
- You need a Windows container.

Mokapi is a single static binary (built with `CGO_ENABLED=0`), so it runs on any Linux base image without
additional libraries. On Linux there are two ways to get the binary into your image, and Windows needs its own
build.

| Approach | Pick it when |
|----------|--------------|
| [Multi-stage build](#copy-the-binary-from-the-official-image) | You can pull `mokapi/mokapi` during the build. It's the shortest Dockerfile and works for any Linux base image. |
| [Download from GitHub](#download-the-release-package) | The build can reach GitHub but not Docker Hub, or you want to install the `.deb` package. |
| [Windows](#windows-container) | Your containers run on a Windows host in Windows container mode. |

All examples use the same conventions as the official image. The entrypoint is `mokapi`, so everything after the
image name is passed to Mokapi as arguments. Mokapi listens on port `8080` for the dashboard, REST API and health
check, and on the ports defined in your specs for the mocks. Specs and scripts aren't part of the image. Mount
them, for example to `/data`, and pass the path as an argument.

## Copy the binary from the official image

The multi-stage build takes the binary from the official image and puts it into your base image:

```Dockerfile tab=Dockerfile
FROM mokapi/mokapi:0.57.0 AS mokapi

FROM ubuntu:resolute

COPY --from=mokapi /usr/local/bin/mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
```

The first stage only exists to provide the file `/usr/local/bin/mokapi`, which is where the official images place
the binary. Nothing else from that stage ends up in your image. Pin the version tag so that rebuilds are
reproducible, and replace `ubuntu:resolute` with your own base image.

## Download the release package

If you can't pull the official image, install the package from the
[GitHub releases](https://github.com/marle3003/mokapi/releases). The Debian package is named
`mokapi_<version>_linux_<arch>.deb` (without a `v` in the version) and the URL contains the tag with a `v`:

```Dockerfile tab=Dockerfile
FROM ubuntu:resolute

ARG MOKAPI_VERSION=0.57.0

RUN apt-get update && \
    apt-get install -y wget ca-certificates && \
    rm -rf /var/lib/apt/lists/*

RUN MOKAPI_DEB="$(mktemp)" && \
    wget -O "$MOKAPI_DEB" "https://github.com/marle3003/mokapi/releases/download/v${MOKAPI_VERSION}/mokapi_${MOKAPI_VERSION}_linux_$(dpkg --print-architecture).deb" && \
    dpkg -i "$MOKAPI_DEB" && \
    rm -f "$MOKAPI_DEB"

ENTRYPOINT ["mokapi"]
```

`dpkg --print-architecture` returns `amd64` or `arm64`, which are the two architectures published for Linux. The
package installs the binary as `mokapi` on the `PATH`, so the entrypoint doesn't need a path. On a base image
without `dpkg`, download the archive `mokapi_v<version>_linux_<arch>.tar.gz` instead and move the extracted
`mokapi` binary to `/usr/local/bin`.

If a corporate proxy re-signs HTTPS traffic, `wget` fails with a certificate error until the proxy's CA is in the
image. Add the CA before the download step instead of turning off certificate verification.

## Windows container

For Windows containers, download the zip archive `mokapi_v<version>_windows_amd64.zip`. Only the amd64
build is published for Windows.

```Dockerfile tab=Dockerfile
FROM mcr.microsoft.com/windows/nanoserver:ltsc2019

ADD https://github.com/marle3003/mokapi/releases/download/v0.57.0/mokapi_v0.57.0_windows_amd64.zip C:/mokapi.zip

RUN tar -xf C:\mokapi.zip -C C:\

ENTRYPOINT ["C:\\mokapi.exe"]
```

The base image tag must match your Windows host version unless you run the container with Hyper-V isolation.

## A complete example: Ubuntu with a health check

This example builds an Ubuntu-based image with `curl` and uses it in Docker Compose with a health check. It
builds on the multi-stage approach and adds the tool that the `ubuntu` image lacks.

```Dockerfile tab=Dockerfile
FROM mokapi/mokapi:0.57.0 AS mokapi

FROM ubuntu:resolute

RUN apt-get update && \
    apt-get install -y --no-install-recommends curl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=mokapi /usr/local/bin/mokapi /usr/local/bin/mokapi

EXPOSE 8080

ENTRYPOINT ["mokapi"]
```
```yaml tab=docker-compose.yml
services:
  mokapi:
    build: .
    command: /data
    ports:
      - "8080:8080"
      - "80:80"
    volumes:
      - ./mocks:/data:ro
    healthcheck:
      test: ["CMD", "curl", "-fsS", "http://localhost:8080/health"]
      interval: 5s
      timeout: 2s
      retries: 10
```

`-f` makes `curl` exit with an error on an HTTP error status, so the check fails whenever `/health` doesn't
answer with `200`. Mokapi serves the endpoint on the dashboard port by default. If you changed `--api-port`,
`--health-port` or `--health-path`, use the same values in the URL, see
[Health Check](/docs/configuration/health-check). Port `80` in the example is for HTTP mocks whose server URL has no
port. Publish the ports your own specs define.

## Verify the image

Build the image and check that the binary starts:

```bash
docker build -t my-mokapi .
docker run --rm my-mokapi --version
```

Then start it with your mocks and look at the log. Mokapi logs the hosts and ports it binds, which shows whether
your specs were loaded:

```bash
docker compose up -d
docker compose logs mokapi
docker compose ps
```

`docker compose ps` shows `healthy` once the health check passes. You can also call the endpoints from your
host:

```bash
curl http://localhost:8080/health
```

If the container stops right after the start, the log usually says why. A common reason is a wrong path in
`command`, because the specs are read from the mounted volume and not from the image.

## Related pages

- [Running Mokapi in Docker and Docker Compose](/docs/configuration/use-cases/docker.md): ports, environment
  variables, volumes and health check of the official image
- [Use Your Own Root CA for HTTPS](/docs/configuration/use-cases/custom-root-ca.md): sign the certificates that
  Mokapi generates with your own CA
- [Installation](/docs/get-started/installation.md): install Mokapi without Docker

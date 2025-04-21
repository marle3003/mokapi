---
title: Mokapi with custom base image
description: How to use Mokapi with your custom base Docker image
icon: bi-box-seam-fill
---

# Use Mokapi with a custom base Docker image

## Download Mokapi package from GitHub

```Dockerfile tab=Dockerfile  
FROM ubuntu:noble

RUN apt-get update && \
	apt-get install -y wget && \
	rm -rf /var/lib/apt/lists/*

RUN MOKAPI_DEB="$(mktemp)" && \
	wget -O "$MOKAPI_DEB" 'https://github.com/marle3003/mokapi/releases/download/v0.9.25/mokapi_0.9.25_linux_amd64.deb' --no-check-certificate && \
	dpkg -i "$MOKAPI_DEB" && \
	rm -f "$MOKAPI_DEB"
	
ENTRYPOINT ["mokapi"]
```

## Using Multi-stage build

```Dockerfile tab=Dockerfile  
FROM mokapi/mokapi:v0.9.25 as mokapi

FROM ubuntu:noble

COPY --from=mokapi /usr/local/bin/mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
```

## Using Windows as base image

```Dockerfile tab=Dockerfile  
FROM mcr.microsoft.com/windows/nanoserver:ltsc2019

RUN curl --location --output mokapi.zip --request GET https://github.com/marle3003/mokapi/releases/download/v0.9.25/mokapi_v0.9.25_windows_amd64.zip 

RUN tar -xf mokapi.zip
	
ENTRYPOINT ["mokapi"]
```
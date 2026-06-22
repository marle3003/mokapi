ARG VERSION

FROM node:26.3.1@sha256:3c05c2cf0f6a5795dfb7abefb2a4e31a78d6271a99962531c48315ced17d618a as webui

COPY ./webui ./webui

WORKDIR webui

COPY ./docs ./src/assets/docs

RUN npm install
RUN npm run build-dashboard

FROM golang:1.25.5-alpine@sha256:ac09a5f469f307e5da71e766b0bd59c9c49ea460a528cc3e6686513d64a6f1fb AS gobuild

ARG VERSION=dev

ARG BUILD_TIME=dev

# Install git for GIT tests
RUN apk add --no-cache git

COPY . /go/src/github.com/mokapi

WORKDIR /go/src/github.com/mokapi

RUN rm -rf ./webui
COPY --from=webui /webui webui

RUN go test -v ./...

RUN go build -o mokapi -ldflags="-X mokapi/version.BuildVersion=$VERSION -X mokapi/version.BuildTime=$BUILD_TIME" ./cmd/mokapi

FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

COPY --from=gobuild /go/src/github.com/mokapi/mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
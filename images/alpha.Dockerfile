ARG VERSION

FROM node:26.12.1@sha256:97fa0f1b59035c6072721a88d55be1371c7cb1cc942b521e1cc5c72ec87655f4 as webui

COPY ./webui ./webui

WORKDIR webui

COPY ./docs ./src/assets/docs

RUN npm install
RUN npm run build-dashboard

FROM golang:1.27.3-alpine@sha256:b8c9b82affbbb82a020bbcf998dad2ba9516944ac7e883524642a481a9505e9c AS gobuild

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

FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a

COPY --from=gobuild /go/src/github.com/mokapi/mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
ARG VERSION

FROM node:26.12.1@sha256:a98e6fda54fdbb22d4238054cca9a2a113bcd1191c09193a535aae22f400aa8e as webui

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

FROM alpine:3.24.3@sha256:b3413572228e34bd1c9245f5565e6cd127005655e8a7899aef33d8e83255c9e6

COPY --from=gobuild /go/src/github.com/mokapi/mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
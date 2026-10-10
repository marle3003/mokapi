ARG VERSION

FROM node:26.11.1@sha256:32fa97f3363975684b08bf4e8a68a47c7905175cc20275b50b974bbd02aba731 as webui

COPY ./webui ./webui

WORKDIR webui

COPY ./docs ./src/assets/docs

RUN npm install
RUN npm run build-dashboard

FROM golang:1.27.2-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS gobuild

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

FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

COPY --from=gobuild /go/src/github.com/mokapi/mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
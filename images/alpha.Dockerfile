ARG VERSION

FROM node:26.13.0@sha256:438603bbb5cda404b9416f0b5c8026f61501ab911760d861ec28c36b5d675725 as webui

COPY ./webui ./webui

WORKDIR webui

COPY ./docs ./src/assets/docs

RUN npm install
RUN npm run build-dashboard

FROM golang:1.27.3-alpine@sha256:fa3d596898d9f126af7c01d14ee7597bab3f39cccf392ef36b69fe77cd681f18 AS gobuild

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

FROM alpine:3.24.4@sha256:93b41ef40b1677fce8fcc64ff4af09daddfacf5040afde618bc46c1e2d1cc7d4

COPY --from=gobuild /go/src/github.com/mokapi/mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
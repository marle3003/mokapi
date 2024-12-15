ARG VERSION

FROM node:20.11.1 as webui

COPY ./webui ./webui

WORKDIR webui

COPY ./docs ./src/assets/docs

RUN npm install
RUN npm run build

FROM golang:1.23.4-alpine AS gobuild

ARG VERSION=dev

COPY . /go/src/github.com/mokapi

WORKDIR /go/src/github.com/mokapi

RUN rm -rf ./webui
COPY --from=webui /webui webui

RUN go test -v ./...

RUN go build -o mokapi -ldflags="-X mokapi/version.BuildVersion=$VERSION" ./cmd/mokapi

FROM alpine

COPY --from=gobuild /go/src/github.com/mokapi/mokapi /

ENTRYPOINT ["/mokapi"]
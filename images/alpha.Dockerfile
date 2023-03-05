ARG VERSION

FROM node:16.13.1 as webui

COPY ./webui ./webui

WORKDIR webui

COPY ./docs ./src/assets/docs

RUN npm install
RUN npm run build

FROM golang:1.20.1 AS gobuild

ARG VERSION=dev

COPY . /go/src/github.com/mokapi

WORKDIR /go/src/github.com/mokapi

RUN rm -rf ./webui
COPY --from=webui /webui/dist webui

RUN go install -a -v github.com/go-bindata/go-bindata/...@latest
RUN go-bindata -nomemcopy -pkg api -o api/bindata.go -prefix webui/ webui/...

RUN go test -v ./...

RUN go build -o mokapi -ldflags="-X mokapi/version.BuildVersion=$VERSION" ./cmd/mokapi

FROM alpine

COPY --from=gobuild /go/src/github.com/mokapi/mokapi /

ENV MOKAPI_Log.Level=info

#ADD mokapi /

ENTRYPOINT ["/mokapi"]
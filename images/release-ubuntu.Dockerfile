FROM ubuntu:resolute@sha256:53958ec7b67c2c9355df922dd08dbf0360611f8c3cdb656875e81873db9ffdba

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
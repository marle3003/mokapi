FROM ubuntu:resolute@sha256:d66b366dac61b575bee2aef07e9c26e8070096fbe84a3d4d3eb0da069b5d76db

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
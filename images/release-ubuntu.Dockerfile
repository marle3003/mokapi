FROM ubuntu:resolute@sha256:065aa8626841f16057b78b027b464cd59c4fac36c204c72dcbf48d66f5fa8a96

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
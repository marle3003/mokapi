FROM ubuntu:resolute@sha256:678c6550cc43645e08669028bc177f50be4e7c5b8cca677067b1914d4afc7a03

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
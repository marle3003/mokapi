FROM ubuntu:resolute@sha256:f144425ff09be612d6d9ad965196e9cdc23dae1f42110a8a11a3e9a8198759f7

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
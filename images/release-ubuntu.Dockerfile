FROM ubuntu:noble

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
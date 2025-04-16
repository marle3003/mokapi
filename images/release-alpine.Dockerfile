FROM alpine

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
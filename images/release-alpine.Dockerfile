FROM alpine

ADD mokapi /

ENTRYPOINT ["/mokapi"]
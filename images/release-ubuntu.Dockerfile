FROM ubuntu:noble

ADD mokapi /

ENTRYPOINT ["/mokapi"]
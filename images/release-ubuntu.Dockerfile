FROM ubuntu:resolute@sha256:2260313b31c8c011cd2eebe728008efac1b3982be73eb71348ea2648d2c0e09b

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
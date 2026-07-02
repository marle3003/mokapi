FROM ubuntu:resolute@sha256:b7f48194d4d8b763a478a621cdc81c27be222ba2206ca3ca6bc42b49685f3d9e

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
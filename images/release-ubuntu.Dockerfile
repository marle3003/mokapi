FROM ubuntu:resolute@sha256:3131b4cc82a783df6c9df078f86e01819a13594b865c2cad47bd1bca2b7063bb

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
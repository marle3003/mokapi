FROM ubuntu:resolute@sha256:da6fc2be547864451aa253836dd926da33623312df4a9a243e35dc877c378a78

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
FROM ubuntu:resolute@sha256:513c074113a871b51a8d16ab445c88779d6452d937a164fb5cc479f32668a41d

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
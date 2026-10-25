FROM ubuntu:resolute@sha256:c098726dd9b9733f3e885f8de0ef375206a311b0a317eadc40d347b1115fb944

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
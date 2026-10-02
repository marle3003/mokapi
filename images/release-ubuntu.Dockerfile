FROM ubuntu:resolute@sha256:3595d7fc4286a33fad0fd853a4063e654287a9c3787437d7937c94ca3f7a804e

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
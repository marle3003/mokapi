FROM ubuntu:resolute@sha256:6a06111c3b0a8600feadb270aec771dc239bf407b25625f93f44ab00e11f4d3c

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
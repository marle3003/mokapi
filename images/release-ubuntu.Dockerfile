FROM ubuntu:noble@sha256:786a8b558f7be160c6c8c4a54f9a57274f3b4fb1491cf65146521ae77ff1dc54

ADD mokapi /usr/local/bin/mokapi

ENTRYPOINT ["mokapi"]
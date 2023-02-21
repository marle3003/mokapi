FROM servercore-ltsc2022

ARG MOKAPI_VERSION

SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]

RUN Invoke-WebRequest \
        -Uri "https://github.com/marle3003/mokapi/releases/download/$MOKAPI_VERSION/mokapi_$MOKAPI_VERSION_windows_amd64.zip" \
        -OutFile "/mokapi.zip"; \
    Expand-Archive -Path "/mokapi.zip" -DestinationPath "/" -Force; \
    Remove-Item "/mokapi.zip" -Force

ENTRYPOINT ["/mokapi"]
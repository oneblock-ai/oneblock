FROM golang:1.21.4-bookworm
RUN apt-get update -qq && apt-get install -y --no-install-recommends git curl unzip tar tini bash && \
    rm -rf /var/lib/apt/lists/* && \
    adduser --disabled-password oneblock && su -l oneblock && \
    mkdir -p /var/lib/oneblock-ai/oneblock && \
    chown -R oneblock /var/lib/oneblock-ai/oneblock /usr/local/bin

WORKDIR /var/lib/oneblock-ai/oneblock

ENV ONEBLOCK_UI_VERSION latest
ENV ONEBLOCK_UI_PATH /usr/share/oneblock-ai/oneblock
# Please update the api-ui-version in pkg/settings/settings.go when updating the version here.
ENV ONEBLOCK_API_UI_VERSION 1.1.10

ARG VERSION=dev
ENV ONEBLOCK_SERVER_VERSION ${VERSION}

RUN mkdir -p /usr/share/oneblock-ai/oneblock && \
    cd /usr/share/oneblock-ai/oneblock && \
    curl -sL https://releases.rancher.com/dashboard/${ONEBLOCK_UI_VERSION}.tar.gz | tar xvzf - --strip-components=2 && \
    mkdir -p /usr/share/oneblock-ai/oneblock/api-ui && \
    cd /usr/share/oneblock-ai/oneblock/api-ui && \
    curl -sL https://releases.rancher.com/api-ui/${ONEBLOCK_API_UI_VERSION}.tar.gz | tar xvzf - --strip-components=1 && \
    cd /var/lib/oneblock-ai/oneblock

COPY package/entrypoint.sh oneblock /usr/bin/
RUN chmod +x /usr/bin/entrypoint.sh

VOLUME /var/lib/oneblock-ai/oneblock
ENTRYPOINT ["entrypoint.sh"]

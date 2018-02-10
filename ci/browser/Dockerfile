FROM ubuntu:17.10

# Golang
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    git \
    default-jre \
    wget

ENV VERSION=1.9.2
ENV OS=linux
ENV ARCH=amd64
RUN wget https://storage.googleapis.com/golang/go$VERSION.$OS-$ARCH.tar.gz
RUN tar -C /usr/local -xzf go$VERSION.$OS-$ARCH.tar.gz

ENV GOPATH=/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

RUN go get -u github.com/alecthomas/gometalinter && gometalinter --install
RUN go get -u github.com/shuLhan/go-bindata/...

# Node
RUN curl -sL https://deb.nodesource.com/setup_8.x | bash -
RUN apt-get install -y nodejs
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
RUN apt-get update && apt-get install -y yarn
RUN npm install -g gulp

# WebDrivers
RUN  mkdir -p /opt/selenium \
  && wget --no-verbose https://selenium-release.storage.googleapis.com/3.8/selenium-server-standalone-3.8.1.jar \
    -O /opt/selenium/selenium-server-standalone.jar

# firefox
RUN apt-get update && apt-get install -y \
	firefox  \
	xvfb
ENV DISPLAY=localhost:0.0
RUN export BASE_URL=https://github.com/mozilla/geckodriver/releases/download \
  && export VERSION=$(curl -sL \
    https://api.github.com/repos/mozilla/geckodriver/releases/latest | \
    grep tag_name | cut -d '"' -f 4) \
  && curl -sL \
  $BASE_URL/$VERSION/geckodriver-$VERSION-linux64.tar.gz | tar -xz \
  && mv geckodriver /usr/local/bin/geckodriver


# # chrome
# RUN export DEBIAN_FRONTEND=noninteractive \
#   && apt-get update \
#   && apt-get install --no-install-recommends --no-install-suggests -y \
#     chromium-chromedriver \
#   # Start chromium via wrapper script with --no-sandbox argument:
#   && mv /usr/lib/chromium/chromium /usr/lib/chromium/chromium-original \
#   && printf '%s\n' '#!/bin/sh' \
#     'exec /usr/lib/chromium/chromium-original --no-sandbox "$@"' \
#     > /usr/lib/chromium/chromium && chmod +x /usr/lib/chromium/chromium \
#   # Remove obsolete files:
#   && apt-get clean \
#   && rm -rf \
#     /tmp/* \
#     /usr/share/doc/* \
#     /var/cache/* \
#     /var/lib/apt/lists/* \
#     /var/tmp/*

# # Make chromedriver available in the PATH:
# RUN ln -s /usr/lib/chromium/chromedriver /usr/local/bin/

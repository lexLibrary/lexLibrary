FROM ubuntu:17.10

# Golang
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    git \
    tzdata \
    wget

ENV VERSION=1.10
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

# Set Non-UTC timezone
ENV TZ=America/Chicago
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

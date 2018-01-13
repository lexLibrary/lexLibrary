FROM ubuntu:17.04

# Golang
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    git \
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
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list
RUN yarn global install gulp

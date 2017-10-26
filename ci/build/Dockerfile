FROM ubuntu:17.04

# Golang
RUN apt-get update && apt-get install -y \
    curl \
    wget

ENV VERSION=1.9.2
ENV OS=linux
ENV ARCH=amd64
RUN wget https://storage.googleapis.com/golang/go$VERSION.$OS-$ARCH.tar.gz
RUN tar -C /usr/local -xzf go$VERSION.$OS-$ARCH.tar.gz

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=$HOME/workspace


# Node
RUN curl -sL https://deb.nodesource.com/setup_6.x | bash -
RUN apt-get install -y nodejs
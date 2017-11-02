FROM postgres:latest

# Golang
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    wget

ENV VERSION=1.9.2
ENV OS=linux
ENV ARCH=amd64
RUN wget https://storage.googleapis.com/golang/go$VERSION.$OS-$ARCH.tar.gz
RUN tar -C /usr/local -xzf go$VERSION.$OS-$ARCH.tar.gz

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/go
ENV POSTGRES_PASSWORD=password

ENV LLDATABASE=postgres
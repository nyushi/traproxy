#!/bin/bash -x

cd traproxy
export GOOS=linux
for arch in amd64 386 arm; do
    GOARCH=$arch go build
    tar zcf "traproxy_linux_${arch}.tar.gz" traproxy
    rm -f traproxy
done

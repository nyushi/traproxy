# traproxy

[![wercker status](https://app.wercker.com/status/5c6300ff7a8ca6e33d941d8eb55916cd/s "wercker status")](https://app.wercker.com/project/bykey/5c6300ff7a8ca6e33d941d8eb55916cd)
[![Build Status](https://travis-ci.org/nyushi/traproxy.svg?branch=develop)](https://travis-ci.org/nyushi/traproxy)
[![Coverage Status](https://coveralls.io/repos/nyushi/traproxy/badge.png?branch=develop)](https://coveralls.io/r/nyushi/traproxy?branch=develop)


Traproxy is transparent http/https proxy.

Currently only supports linux.

# Description

<img src="./diagram.png" width="442" />

 - `http_proxy` and `https_proxy` environment variable is not necessary
 - Work with Docker containers

# Installation

```
go get github.com/nuyshi/traproxy/traproxy
```

# How to use

```
traproxy -proxyaddr <proxy_host>:<proxy_port>
```

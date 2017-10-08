# Logpipe

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GoDoc](https://godoc.org/github.com/arsham/logpipe?status.svg)](http://godoc.org/github.com/arsham/logpipe)
[![Build Status](https://travis-ci.org/arsham/logpipe.svg?branch=master)](https://travis-ci.org/arsham/logpipe)
[![Coverage Status](https://coveralls.io/repos/github/arsham/logpipe/badge.svg?branch=master)](https://coveralls.io/github/arsham/logpipe?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/arsham/logpipe)](https://goreportcard.com/report/github.com/arsham/logpipe)

Logpipe can redirect your application's `logs` to a generic logfile, or to [ElasticSearch][elasticsearch] for aggregate and view with [kibana][kibana]. It can receive the logs in `JSON` format or plain line.

1. [Features](#features)
    * [Upcoming Features](#upcoming-features)
2. [Installation](#installation)
3. [LICENSE](#license)

## Features

* Very lightweight and fast.
* Can receive from multiple inputs.
* Buffers the recording and passes them to the destination in batch.

### Upcoming Features

* Tail log files.
* Record to more repositories:
    * InfluxDB

## Installation

I will provide a docker image soon, but for now it needs to be installed. You need golang >= 1.7 and [glide][glide] installed. Simply do:

```bash
go get github.com/arsham/logpipe
```

You also need elasticsearch and kibana, here is a couple of docker images you can start with:

```bash
docker volume create logpipe
docker run -d --name logpipe --restart always --ulimit nofile=98304:98304 -v logpipe:/usr/share/elasticsearch/data -e ES_JAVA_OPTS='-Xms10G -Xmx10G' -e "xpack.security.enabled=false" -e "xpack.monitoring.enabled=true" -e "xpack.graph.enabled=true" -e "xpack.watcher.enabled=false" -p 9200:9200 -e "http.cors.enabled=true" -e 'http.cors.allow-origin=*' docker.elastic.co/elasticsearch/elasticsearch:5.6.2
docker run -d --name kibana --restart always -p 80:5601 --link logpipe:elasticsearch docker.elastic.co/kibana/kibana:5.6.2
```

## LICENSE

Use of this source code is governed by the Apache 2.0 license. License that can be found in the [LICENSE](./LICENSE) file.

`Enjoy!`


[glide]: https://github.com/Masterminds/glide
[elasticsearch]: https://github.com/elastic/elasticsearch
[kibana]: https://github.com/elastic/kibana

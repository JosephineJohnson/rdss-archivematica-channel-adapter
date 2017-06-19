[![Travis CI](https://travis-ci.org/JiscRDSS/rdss-archivematica-channel-adapter.svg?branch=master)](https://travis-ci.org/JiscRDSS/rdss-archivematica-channel-adapter) [![GoDoc](https://godoc.org/github.com/JiscRDSS/rdss-archivematica-channel-adapter?status.svg)](https://godoc.org/github.com/JiscRDSS/rdss-archivematica-channel-adapter) [![Coverage Status](https://coveralls.io/repos/github/JiscRDSS/rdss-archivematica-channel-adapter/badge.svg?branch=master)](https://coveralls.io/github/JiscRDSS/rdss-archivematica-channel-adapter?branch=master) [![Go Report Card](https://goreportcard.com/badge/JiscRDSS/rdss-archivematica-channel-adapter)](https://goreportcard.com/report/JiscRDSS/rdss-archivematica-channel-adapter) [![Sourcegraph](https://sourcegraph.com/github.com/JiscRDSS/rdss-archivematica-channel-adapter/-/badge.svg)](https://sourcegraph.com/github.com/JiscRDSS/rdss-archivematica-channel-adapter?badge)

# RDSS Archivematica Channel Adapter

- [Introduction](#introduction)
- [Usage](#usage)
  - [Consumer](#consumer)
  - [Publisher](#publisher)
- [Configuration](#configuration)
  - [Environment variables](#environment-variables)
  - [Configuration file](#configuration-file)
- [Reusability](#reusability)
- [Diagram](#diagram)

## Introduction

**THIS IS A PROTOTYPE!**

This repository holds the source code of the channel adapter that connects Archivematica to [RDSS's messaging API](https://github.com/JiscRDSS/rdss-message-api-docs).

The adapter is written in Go as a standalone application that runs next to Archivematica. Its main role is to abstract the complexities and specifics of the underlying queuing system from its users.

## Usage

We're producing a single binary executable file: **rdss-archivematica-channel-adapter** with two subcommands: **consumer** and **publisher**. You can build the Docker image locally running the following command:

    $ docker build -t rdss-archivematica-channel-adapter .

The **consumer** is the component that brings Archivematica functionality to RDSS. It consumes the messages that come from the stream and convert them into Archivematica-specific calls. You can start it running:

    $ docker run rdss-archivematica-channel-adapter consumer

On the other hand,  the **publisher** is the component that brings RDSS functionality to Archivematica. It is implemented as a gRPC server which and it encapsulates the asynchronous nature of the messaging interaction, exposing regular synchronous methods to the application logic or the client. You can start it running:

    $ docker run rdss-archivematica-channel-adapter publisher

Certain configuration parameter are required though. See the configuration section in this document for more details.

## Configuration

The adapter is not configurable via command-line flags. You can choose between environment variables and the configuration file, having the former method precedence over the latter.

### Environment variables

The following is a list of supported environment variables. They need to be prefixed with the string `RDSS_ARCHIVEMATICA_ADAPTER_`, e.g. `RDSS_ARCHIVEMATICA_ADAPTER_LOGGING.LEVEL=INFO`. Notice that the dot is used to separate nested attributes.

| String                             | Default              |
| ---------------------------------- | -------------------- |
| `LOGGING.LEVEL`                    | `INFO`               |
| `AMCLIENT.URL`                     | ``                   |
| `AMCLIENT.USER`                    | ``                   |
| `AMCLIENT.KEY`                     | ``                   |
| `PUBLISHER.LISTEN`                 | `0.0.0.0:8000`       |
| `PUBLISHER.TLS`                    | `false`              |
| `PUBLISHER.TLS_CERT_FILE`          | ``                   |
| `PUBLISHER.TLS_KEY_FILE`           | ``                   |
| `BROKER.BACKEND`                   | `kinesis`            |
| `BROKER.QUEUES.MAIN`               | `main`               |
| `BROKER.QUEUES.INVALID`            | `invalid`            |
| `BROKER.QUEUES.ERROR`              | `error`              |
| `BROKER.KINESIS.ENDPOINT`          | ``                   |
| `BROKER.KINESIS.DYNAMODB_ENDPOINT` | ``                   |

### Configuration file

The adapter will try to read config from `$HOME/.rdss-archivematica-channel-adapter.toml` and `/etc/archivematica/rdss-archivematica-channel-adapter.toml`. Alternatively, you can pass a different location with the global `--config string` flag.

Notice how the environment variables in the previous section map to the nested configuration sections in the configuration file, e.g.:

```toml
# RDSS Archivematica Channel Adapter

[logging]
level = "INFO"

[amclient]
url = "https://archivematica.internal:9000"
user = "demo"
key = "eid3Aitheijoo1ohce2pho4eiDei0lah"

[publisher]
listen = "0.0.0.0:8000"
tls = false
tls_cert_file = "/foo.crt"
tls_key_file = "/foo.key"

[broker]
backend = "kinesis"

    [broker.kinesis]
    # This adapter uses aws-sdk-go. The credentials must be defined using the
    # canonical environment variables, read more at https://goo.gl/xsWyS9.

    # The name of the Amazon Kinesis stream.
    stream = "rdss-archivematica"

    # Kinesis endpoint. It can be used when you want the client to speak to a
    # server not run by AWS, e.g. a local kinesalite instance used during testing.
    #endpoint = "https://127.0.0.1:4567"
```

## Code reusability

A few Go packages found in this repository are agnostic to Archivematica and could be used by other vendors:

- `github.com/JiscRDSS/rdss-archivematica-channel-adapter/amclient` is a Archivematica HTTP API client.
- `github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker` is a RDSS client conforming to the RDSS messaging API. Both `consumer` and `publisher` packages in this repository are use cases. If you want to know more, there is a comprehensive example in [broker_test.go](broker/broker_test.go).
- `github.com/JiscRDSS/rdss-archivematica-channel-adapter/s3` is a small S3 wrapper used to download files.

## Diagram

This diagram is not up to date but it's close to the current design:

![RDSS Archivematica Channel Adapter Diagram](hack/diagram.png)

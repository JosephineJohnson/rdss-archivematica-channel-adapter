# RDSS Archivematica Channel Adapter [![Build Status](https://travis-ci.com/JiscRDSS/rdss-archivematica-channel-adapter.svg?token=XEKi3UuVjsxnJD1KeZsi&branch=master)](https://travis-ci.com/JiscRDSS/rdss-archivematica-channel-adapter)

- [Introduction](#introduction)
- [Usage](#usage)
- [Diagram](#diagram)

## Introduction

**THIS IS A PROTOTYPE!**

This repository holds the source code of the channel adapter that connects Archivematica to [RDSS's messaging system](https://github.com/JiscRDSS/rdss-message-api-docs).

The adapter is written in Go as a standalone application that runs next to Archivematica. Its main role is to abstract the complexities and specifics of the underlying queuing system from its users.

## Usage

We're producing a single binary executable file: `rdss-archivematica-channel-adapter` with two subcommands: `publisher` and `consumer`.

The advantage of this approach is that they can be deployed separately.

#### `rdss-archivematica-channel-adapter publisher`

The `publisher` is the component that brings RDSS functionality to Archivematica. It is implemented as a gRPC server which and it encapsulates the asynchronous nature of the messaging interaction, exposing regular synchronous methods to the application logic or the client.

The origin of its name is that it publishes or produces messages and puts them into the stream.

The full help message can be obtained running: `rdss-archivematica-channel-adapter publisher --help`.

```
rdss-archivematica-channel-adapter publisher \
  --bind 0.0.0.0 \
  --port 8000 \
  --tls \
  --tls-cert-file /tmp/my.crt \
  --tls-key-file /tmp/my.key \
  --kinesis-stream the-name-of-the-stream
```

The Kinesis backend is based on `aws-sdk-go` which takes some extra environment variables, e.g.: `AWS_REGION`, `AWS_ACCESS_KEY` or `AWS_SECRET_KEY`.

#### `rdss-archivematica-channel-adapter consumer`

The `consumer` is the component that brings Archivematica functionality to RDSS.

It consumes the messages that come from the stream and convert them into Archivematica-specific calls.

The full help message can be obtained running: `rdss-archivematica-channel-adapter consumer --help`.

## Diagram

**This diagram may be out of date!**

![RDSS Archivematica Channel Adapter Diagram](hack/diagram.png)

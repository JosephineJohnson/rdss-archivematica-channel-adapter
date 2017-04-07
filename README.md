# RDSS Archivematica Channel Adapter [![Build Status](https://travis-ci.com/JiscRDSS/rdss-archivematica-channel-adapter.svg?token=XEKi3UuVjsxnJD1KeZsi&branch=master)](https://travis-ci.com/JiscRDSS/rdss-archivematica-channel-adapter)

- [Introduction](#introduction)
- [Usage](#usage)
- [Diagram](#diagram)

## Introduction

This repository holds the source code of the channel adapter that connects Archivematica to [RDSS's messaging system](https://github.com/JiscRDSS/rdss-message-api-docs).

The adapter is written in Go as a standalone application that runs next to Archivematica. Its main role is to abstract the complexities and specifics of the underlying queuing system from its users.

## Usage

We're producing a single binary executable file: `rdss-archivematica-channel-adapter` with two subcommands: `publisher` and `consumer`.

They can be deployed separately as they have different needs. You may want to have one single `publisher` and many instances of `consumer`. They have different needs from the operational perspective, e.g. you may want to run one single instance of `publisher` but multiple instances of `consumer` and set up a persistent data store to hold the application state (e.g. checkpoints, shard association, etc...).

**`rdss-archivematica-channel-adapter publisher`**

The `publisher` is the component that brings RDSS functionality to Archivematica. It is implemented as a gRPC server which allows us to present both a HTTP 1.1 RET/JSON API and an efficient gRPC interface on a single TPC port. This interface encapsulates the asynchronous nature of the messaging interaction, exposing regular synchronous methods to the application logic or the client.

The origin of its name is that it publishes or produces messages and puts them into the stream.

```
$ rdss-archivematica-channel-adapter publisher --help
Outbound server (Archivematica » RDSS)

Usage:
  rdss-archivematica-channel-adapter publisher [flags]

Flags:
  -b, --bind string   interface to which the gRPC server will bind (default "0.0.0.0")
  -p, --port int      port on which the gRPC server will listen (default 8000)

Global Flags:
      --config string   config file (default is $HOME/.rdss-archivematica-channel-adapter.yaml)
```

**`rdss-archivematica-channel-adapter consumer`**

The `consumer` is the component that brings Archivematica functionality to RDSS.

It consumes the messages that come from the stream and convert them into Archivematica-specific calls.

```
rdss-archivematica-channel-adapter consumer --help
Inbound server (RDSS » Archivematica)

Usage:
  rdss-archivematica-channel-adapter consumer [flags]

Global Flags:
      --config string   config file (default is $HOME/.rdss-archivematica-channel-adapter.yaml)
```

## Diagram

**This diagram may be out of date!**

![RDSS Archivematica Channel Adapter Diagram](hack/diagram.png)

## Known issues

It is currently not possible to run multiple instances of the `consumer` if you are using the Kinesis backend. A single `consumer` will do its best effort to consume all the shards available and listen to split and merges. This limitation can be addressed later if needed using solutions like KCL (Java-based daemon) or the [kinesumer](https://github.com/remind101/kinesumer) Go library.

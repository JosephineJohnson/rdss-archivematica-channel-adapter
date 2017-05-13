FROM golang:1.8.1-alpine

# This Dockerfile is less useful in production environments!
# One specific to production should be created.

ARG APP_PATH=github.com/JiscRDSS/rdss-archivematica-channel-adapter

RUN addgroup -g 333 -S archivematica && adduser -u 333 -S -G archivematica archivematica

ADD ./ /go/src/$APP_PATH

WORKDIR /go/src/$APP_PATH

RUN set -x \
	&& apk add --no-cache --virtual .build-deps make gcc musl-dev \
	&& make \
	&& make build
	# Sometimes useful during development
	# && apk del .build-deps

USER archivematica

ENTRYPOINT ["/go/bin/rdss-archivematica-channel-adapter"]

CMD ["help"]

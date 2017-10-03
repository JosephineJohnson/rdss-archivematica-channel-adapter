FROM golang:1.9.0-alpine
WORKDIR /go/src/github.com/JiscRDSS/rdss-archivematica-channel-adapter
COPY . .
RUN set -x \
	&& apk add --no-cache ca-certificates \
	&& apk add --no-cache --virtual .build-deps make gcc musl-dev git \
	&& make \
	&& make install
RUN set -x \
	&& addgroup -g 333 -S archivematica \
	&& adduser -u 333 -h /var/lib/archivematica -S -G archivematica archivematica
USER archivematica
ENTRYPOINT ["/go/bin/rdss-archivematica-channel-adapter"]
CMD ["help"]

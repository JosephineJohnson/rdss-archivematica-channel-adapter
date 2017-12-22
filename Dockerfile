FROM golang:1.9.1-alpine
ARG GITHUB_API_TOKEN
WORKDIR /go/src/github.com/JiscRDSS/rdss-archivematica-channel-adapter
COPY . .
# Don't use `make testrace`, it won't work in Alpine Linux!
RUN set -x \
	&& apk add --no-cache ca-certificates \
	&& apk add --no-cache --virtual .build-deps \
		make gcc musl-dev git bash curl \
	&& ./hack/download-schemas.sh $GITHUB_API_TOKEN \
	&& make test vet \
	&& make install
RUN set -x \
	&& addgroup -g 333 -S archivematica \
	&& adduser -u 333 -h /var/lib/archivematica -S -G archivematica archivematica
ENV RDSS_ARCHIVEMATICA_ADAPTER_BROKER.SCHEMAS_DIR /go/src/github.com/JiscRDSS/rdss-archivematica-channel-adapter/hack/schemas
USER archivematica
ENTRYPOINT ["/go/bin/rdss-archivematica-channel-adapter"]
CMD ["help"]

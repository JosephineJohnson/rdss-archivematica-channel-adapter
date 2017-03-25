build:
	go install $(go list ./... | grep -v /vendor/)

test:
	go test $(go list ./... | grep -v /vendor/)

vendor-status:
	dep status

proto:
	@hack/build-proto.sh

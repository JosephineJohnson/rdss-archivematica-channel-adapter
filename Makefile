PKGS?=$$(go list ./... | grep -v '/vendor/')
FILES?=$$(find . -name '*.go' | grep -v vendor)

default: test vet

tools:
	go get -u github.com/golang/dep/...
	go get -u golang.org/x/tools/cmd/cover
	go get -u github.com/golang/protobuf/protoc-gen-go

build:
	@env CGO_ENABLED=0 go install $(PKGS)

test:
	@go test -i $(PKGS)
	@go test $(PKGS)

testrace:
	@go test -i -race $(PKGS)
	@go test -race $(PKGS)

vet:
	@go vet $(PKGS); if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

cover:
	go test $(PKGS) -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out

vendor-status:
	dep status

proto:
	hack/build-proto.sh

.NOTPARALLEL:

.PHONY: default tools build test testrace vendor-status proto

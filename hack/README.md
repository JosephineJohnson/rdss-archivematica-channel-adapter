### Vendoring

We're using `dep` which is not stable yet but it's good enough for what we
need. Run `dep status` to know more about our dependencies and constraints
established.

### Development dependencies

You're going to need:

```
go get -u github.com/golang/dep/...
go get -u github.com/golang/protobuf/protoc-gen-go
```

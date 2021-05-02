# gRPC -> REST via gRPC Gateway Example

## Install the dependencies

For Go,

```shell
go install \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
  google.golang.org/protobuf/cmd/protoc-gen-go \
  google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

[Buf](https://buf.build) is used to generate proto files.

```shell
buf generate
```


## Deploy to Heroku

```shell
heroku container:push web
heroku container:release web

# Check the app
heroku open
```

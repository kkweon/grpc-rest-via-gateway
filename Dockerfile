FROM golang:alpine AS build_base

RUN apk add --no-cache git curl

RUN BIN="/usr/local/bin" && \
   VERSION="0.41.0" && \
   BINARY_NAME="buf" && \
     curl -sSL \
       "https://github.com/bufbuild/buf/releases/download/v${VERSION}/${BINARY_NAME}-$(uname -s)-$(uname -m)" \
       -o "${BIN}/${BINARY_NAME}" && \
     chmod +x "${BIN}/${BINARY_NAME}"

# Set the Current Working Directory inside the container
WORKDIR /app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download && go install \
                         github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
                         github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
                         google.golang.org/protobuf/cmd/protoc-gen-go \
                         google.golang.org/grpc/cmd/protoc-gen-go-grpc

COPY . .

RUN buf generate

# Build the Go app
RUN go build -o server cmd/main.go

# Start fresh from a smaller image
FROM alpine:latest
RUN apk add ca-certificates

COPY --from=build_base /app/server /app/server
COPY --from=build_base /app/swagger-ui /app/swagger-ui
COPY --from=build_base /app/gen/openapiv2 /app/gen/openapiv2

RUN adduser -D kkweon
USER kkweon

WORKDIR /app
CMD ./server --port $PORT

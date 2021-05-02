package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/kkweon/grpc-rest-via-gateway/gen/go/blog/v1"
	"google.golang.org/grpc"
)

type blogImpl struct{}

const addr = ":80"

func allHandler(grpcServer *grpc.Server, httpHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpHandler.ServeHTTP(w, r)
		}
	})
}

func main() {
	grpcServer := grpc.NewServer()
	v1.RegisterBlogServiceServer(grpcServer, blogImpl{})

	gwmux := runtime.NewServeMux()
	v1.RegisterBlogServiceHandlerFromEndpoint(context.Background(), gwmux, addr)

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)

	conn, err := net.Listen("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}

	server := http.Server{
		Addr:    addr,
		Handler: allHandler(grpcServer, mux),
	}

	if err := server.Serve(conn); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kkweon/grpc-rest-via-gateway/gen/go/blog/v1"
	"google.golang.org/grpc"
)

type blogImpl struct {
	posts []*v1.Post
	v1.UnimplementedBlogServiceServer
}

func (b *blogImpl) CreatePost(ctx context.Context, request *v1.CreatePostRequest) (*v1.CreatePostResponse, error) {
	post := &v1.Post{
		Id:        b.getNewId(),
		Content:   request.GetContent(),
		CreatedAt: timestamppb.Now(),
	}
	b.posts = append(b.posts, post)
	return &v1.CreatePostResponse{Post: post}, nil
}

func (b *blogImpl) GetPosts(ctx context.Context, request *v1.GetPostsRequest) (*v1.GetPostsResponse, error) {
	if request.GetPostId() > 0 {
		for _, post := range b.posts {
			if post.GetId() == request.GetPostId() {
				return &v1.GetPostsResponse{Posts: []*v1.Post{post}}, nil
			}
		}

		return nil, fmt.Errorf("unable to find post_id = %d", request.GetPostId())
	}

	return &v1.GetPostsResponse{Posts: b.posts}, nil
}

func (b *blogImpl) DeletePost(ctx context.Context, request *v1.DeletePostRequest) (*v1.DeletePostResponse, error) {
	for i, post := range b.posts {
		if post.GetId() == request.GetPostId() {
			b.posts = append(b.posts[:i], b.posts[i+1:]...)
			return &v1.DeletePostResponse{}, nil
		}
	}

	return nil, fmt.Errorf("unable to find post_id = %d", request.GetPostId())
}

func (b *blogImpl) getNewId() int64 {
	return time.Now().UnixNano()
}

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
	v1.RegisterBlogServiceServer(grpcServer, &blogImpl{})

	gwmux := runtime.NewServeMux()
	err := v1.RegisterBlogServiceHandlerFromEndpoint(context.Background(), gwmux, addr, []grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)

	conn, err := net.Listen("tcp", ":80")
	if err != nil {
		panic(err)
	}

	server := http.Server{
		Addr:    addr,
		Handler: allHandler(grpcServer, mux),
	}

	if err := server.Serve(conn); err != nil {
		panic(err)
	}
}

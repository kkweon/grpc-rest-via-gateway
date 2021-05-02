package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"strings"
	"time"

	"github.com/kkweon/grpc-rest-via-gateway/gen/go/blog/v1"
	"google.golang.org/grpc"
)

var port = flag.Int("port", 80, "--port 80")

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

func allHandler(grpcServer *grpc.Server, httpHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithField("request", fmt.Sprintf("%+v", r)).Info("hit Handler")
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

func main() {
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	logrus.WithField("addr", addr).Info("flag parsed")

	grpcServer := grpc.NewServer()
	v1.RegisterBlogServiceServer(grpcServer, &blogImpl{})
	reflection.Register(grpcServer)

	gwmux := runtime.NewServeMux()
	err := v1.RegisterBlogServiceHandlerFromEndpoint(context.Background(), gwmux, addr, []grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "gen/openapiv2/blog/v1/blog.swagger.json")
	})
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("swagger-ui/dist"))))
	mux.Handle("/", gwmux)

	err = http.ListenAndServe(addr, allHandler(grpcServer, mux))
	if err != nil {
		panic(err)
	}
}

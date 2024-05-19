package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"kdev/pkg/server"
	"kdev/pkg/service"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

var (
	grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:50051", "gRPC server endpoint")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := service.RegisterExampleServiceHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return err
	}

	exampleServer := server.NewExampleServer()

	httpMux := http.NewServeMux()
	httpMux.Handle("/", server.JwtAuthentication(mux, exampleServer))

	return http.ListenAndServe(":8080", httpMux)
}

func main() {
	flag.Parse()

	exampleServer := server.NewExampleServer()

	go func() {
		lis, err := net.Listen("tcp", *grpcServerEndpoint)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer(grpc.UnaryInterceptor(server.UnaryInterceptor(exampleServer)))
		service.RegisterExampleServiceServer(s, exampleServer)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to serve: %v\n", err)
		os.Exit(1)
	}
}

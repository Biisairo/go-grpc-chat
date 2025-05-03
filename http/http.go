package main

import (
	"context"
	"flag"
	gen "grpc-example/chatting"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
)

var (
	port = flag.String("p", "08061", "Port")
)

func main() {
	flag.Parse()

	address := "localhost:" + *port

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := gen.RegisterChattingHandlerFromEndpoint(ctx, mux, address, opts)
	if err != nil {
		grpclog.Fatal(err)
	}

	http.ListenAndServe("localhost:8080", mux)
}

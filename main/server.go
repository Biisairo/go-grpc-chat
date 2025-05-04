package main

import (
	"flag"
	pb "grpc-example/chatting"
	chattingserver "grpc-example/chattingserver"
	"log"
	"net"

	"google.golang.org/grpc"
)

var (
	port = flag.String("p", "08061", "Port")
)

func main() {
	flag.Parse()

	address := "localhost:" + *port
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Fail to Listen: %v", err)
	}

	// server := grpc.NewServer(grpc.EmptyServerOption{})
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			chattingserver.CustomUnaryMiddleware(),
		),
		grpc.ChainStreamInterceptor(
			chattingserver.CustomStreamMiddleware(),
		),
	)
	pb.RegisterChattingServer(server, chattingserver.NewServer())
	server.Serve(lis)
}

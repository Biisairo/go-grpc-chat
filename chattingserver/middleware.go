package chattingserver

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
)

func CustomUnaryMiddleware() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		log.Print("Requested at:", time.Now())

		fmt.Println("req : ", req)
		fmt.Println("info : ", info)

		resp, err := handler(ctx, req)
		return resp, err
	}
}

func CustomStreamMiddleware() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		log.Print("Requested at:", time.Now())

		fmt.Println("srv : ", srv)
		fmt.Println("info : ", info)

		err := handler(srv, ss)
		return err
	}
}

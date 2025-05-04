package chattingserver

import (
	pb "grpc-example/chatting"
	"sync"
)

type UserInRoom struct {
	Buffer []*pb.Message
}

type Room struct {
	RoomName string
	Users    map[int32]*UserInRoom

	mu sync.Mutex
}

type chattingServer struct {
	pb.UnimplementedChattingServer

	Users map[int32]struct{}
	Rooms map[int32]*Room
}

func NewServer() *chattingServer {
	s := &chattingServer{
		Users: make(map[int32]struct{}),
		Rooms: map[int32]*Room{}}
	return s
}

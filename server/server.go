package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	pb "grpc-example/chatting"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	port = flag.String("p", "08061", "Port")
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

func newServer() *chattingServer {
	s := &chattingServer{
		Users: make(map[int32]struct{}),
		Rooms: map[int32]*Room{}}
	return s
}

func (c *chattingServer) GetUserId(ctx *context.Context) (int32, error) {
	md, ok := metadata.FromIncomingContext(*ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "no metadata")
	}

	userIdStrs := md.Get("user_id")
	userId64, err := strconv.ParseInt(userIdStrs[0], 10, 32)
	if err != nil {
		return 0, status.Errorf(codes.InvalidArgument, "invalid user-id: %v", err)
	}

	userId := int32(userId64)

	return userId, nil
}

func (c *chattingServer) GetRoomId(ctx *context.Context) (int32, error) {
	md, ok := metadata.FromIncomingContext(*ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "no metadata")
	}

	roomIdStrs := md.Get("room_id")
	roomId64, err := strconv.ParseInt(roomIdStrs[0], 10, 32)
	if err != nil {
		return 0, status.Errorf(codes.InvalidArgument, "invalid user-id: %v", err)
	}

	roomId := int32(roomId64)

	return roomId, nil
}

func (c *chattingServer) LoginUser() (int32, error) {
	for i := 0; i < 100; i++ {
		tmp := rand.Int31()
		_, ok := c.Users[tmp]
		if !ok {
			c.Users[tmp] = struct{}{}
			return tmp, nil
		}
	}
	return 0, errors.New("can not make new user")
}

func (c *chattingServer) LogoutUser(userId int32) {
	delete(c.Users, userId)
}

func (c *chattingServer) CreateRoomId(roomName string) (int32, error) {
	for i := 0; i < 100; i++ {
		tmp := rand.Int31()
		_, ok := c.Users[tmp]
		if !ok {
			c.Rooms[tmp] = &Room{
				RoomName: roomName,
				Users:    map[int32]*UserInRoom{},
			}
			return tmp, nil
		}
	}
	return 0, errors.New("can not make new user")
}

func (c *chattingServer) RemoveRoomId(roomNumber int32) {
	delete(c.Rooms, roomNumber)
}

func (c *chattingServer) FindRoom(roomId int32) (*Room, error) {
	room, ok := c.Rooms[roomId]
	if !ok {
		return nil, errors.New("room not found")
	}

	return room, nil
}

func main() {
	flag.Parse()

	address := "localhost:" + *port
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Fail to Listen: %v", err)
	}

	server := grpc.NewServer(grpc.EmptyServerOption{})
	pb.RegisterChattingServer(server, newServer())
	server.Serve(lis)
}

func (s *chattingServer) Login(_ context.Context, _ *pb.Empty) (*pb.User, error) {
	fmt.Println("login")
	userId, err := s.LoginUser()
	if err != nil {
		return &pb.User{}, err
	}

	log.Default().Printf("user %v login\n", userId)

	return &pb.User{UserId: userId}, nil
}

func (s *chattingServer) Logout(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	userId, err := s.GetUserId(&ctx)
	if err != nil {
		return nil, err
	}

	s.LogoutUser(userId)

	return nil, nil
}

func (s *chattingServer) GetChatRoom(_ *pb.Empty, stream pb.Chatting_GetChatRoomServer) error {
	for roomNumber, room := range s.Rooms {
		room := pb.Room{
			RoomId:   roomNumber,
			RoomName: room.RoomName,
		}
		if err := stream.Send(&room); err != nil {
			return err
		}
	}

	return nil
}

func (s *chattingServer) CreateRoom(_ context.Context, room *pb.CreateRoomRequest) (*pb.Room, error) {
	roomId, err := s.CreateRoomId(room.RoomName)
	if err != nil {
		return nil, err
	}

	return &pb.Room{RoomId: roomId, RoomName: room.RoomName}, nil
}

func (s *chattingServer) RemoveRoom(_ context.Context, room *pb.RemoveRoomRequest) (*pb.Empty, error) {
	s.RemoveRoomId(room.RoomId)

	return nil, nil
}

func (s *chattingServer) EnterChatRoom(ctx context.Context, room *pb.RoomRequest) (*pb.Empty, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no metadata")
	}

	userIdStrs := md.Get("user_id")
	fmt.Println(userIdStrs)
	userId64, err := strconv.ParseInt(userIdStrs[0], 10, 32)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user-id: %v", err)
	}

	userId := int32(userId64)

	targetRoom, ok := s.Rooms[room.RoomId]
	if !ok {
		return nil, errors.New("no room exist")
	}

	targetRoom.Users[userId] = &UserInRoom{}

	fmt.Println(targetRoom)

	return nil, nil
}

func (s *chattingServer) ExitChatRoom(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no metadata")
	}

	roomIdStrs := md.Get("room_id")
	roomId64, err := strconv.ParseInt(roomIdStrs[0], 10, 32)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user-id: %v", err)
	}

	roomId := int32(roomId64)

	userIdStrs := md.Get("user_id")
	userId64, err := strconv.ParseInt(userIdStrs[0], 10, 32)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user-id: %v", err)
	}

	userId := int32(userId64)

	targetRoom, ok := s.Rooms[roomId]
	if !ok {
		return nil, errors.New("no room exist")
	}

	delete(targetRoom.Users, userId)

	return nil, nil
}

func (s *chattingServer) Chatting(stream pb.Chatting_ChattingServer) error {
	waitc := make(chan struct{})

	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "no metadata")
	}

	roomIdStrs := md.Get("room_id")
	roomId64, err := strconv.ParseInt(roomIdStrs[0], 10, 32)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid user-id: %v", err)
	}

	roomId := int32(roomId64)
	room, ok := s.Rooms[roomId]
	if !ok {
		return errors.New("room not found")
	}

	userIdStrs := md.Get("user_id")
	userId64, err := strconv.ParseInt(userIdStrs[0], 10, 32)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid user-id: %v", err)
	}

	userId := int32(userId64)
	user, ok := room.Users[userId]
	if !ok {
		return errors.New("user not found")
	}

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("client.Chatting failed: %v", err)
				close(waitc)
				return
			}

			fmt.Println(in)

			room.mu.Lock()
			for id, buffer := range room.Users {
				if userId != id {
					buffer.Buffer = append(buffer.Buffer, in)
				}
				fmt.Println(buffer)
			}
			room.mu.Unlock()

			fmt.Println(room)
		}
	}()

	for {
		room.mu.Lock()
		for _, buffer := range user.Buffer {
			if err := stream.Send(buffer); err != nil {
				return err
			}
		}
		user.Buffer = []*pb.Message{}
		room.mu.Unlock()
	}

	// <-waitc
}

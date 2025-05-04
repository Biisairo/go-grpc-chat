package chattingserver

import (
	"context"
	"errors"
	"fmt"
	pb "grpc-example/chatting"
	"io"
	"log"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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

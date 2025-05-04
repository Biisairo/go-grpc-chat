package chattingserver

import (
	"context"
	"errors"
	"math/rand"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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

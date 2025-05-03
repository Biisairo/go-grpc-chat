package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	pb "grpc-example/chatting"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	port = flag.String("p", "08061", "Port")
)

type chattingClient struct {
	Cl pb.ChattingClient

	UserId int32
	RoomId int32
}

func main() {
	// get params
	flag.Parse()

	// connect
	address := "localhost:" + *port
	fmt.Printf("Open Client to addr \"%v\"\n", address)

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		log.Fatalf("Fail to Create Client: %v", err)
	}
	defer conn.Close()

	client := pb.NewChattingClient(conn)

	chattingClient := chattingClient{Cl: client}

	for {
		reader := bufio.NewScanner(os.Stdin)
		fmt.Print("% ")

		if reader.Scan() {
			input := reader.Text()

			token := strings.Split(input, " ")

			argc := len(token)
			if argc == 0 {
				continue
			}

			firstToken := token[0]

			if !strings.HasPrefix(firstToken, "/") {
				continue
			}

			cmd := strings.ToLower(firstToken[1:])
			switch cmd {
			case "login":
				Login(&chattingClient)
			case "logout":
				Logout(&chattingClient)
			case "create":
				if argc > 1 {
					CreateRoom(&chattingClient, token[1])
				}
			case "remove":
				if argc > 1 {
					roomId, err := strconv.Atoi(token[1])
					if err != nil {
						continue
					}
					DeleteRoom(&chattingClient, int32(roomId))
				}
			case "rooms":
				rooms, err := GetChatRoom(&chattingClient)
				if err != nil {
					continue
				}

				for _, room := range rooms {
					fmt.Printf("| %v | %v\n", room.RoomId, room.RoomName)
				}
			case "enter":
				roomId, err := strconv.Atoi(token[1])
				if err != nil {
					continue
				}

				err = EnterChatRoom(&chattingClient, int32(roomId))
				if err != nil {
					continue
				}

				Chatting(&chattingClient)
				ExitChatRoom(&chattingClient, int32(roomId))
			}
		}
	}

}

func Login(client *chattingClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := client.Cl.Login(ctx, nil)
	if err != nil {
		log.Fatalf("client.Login failed: %v", err)
		return err
	}

	fmt.Printf("login to user %v\n", user.UserId)

	client.UserId = user.UserId
	return nil
}

func Logout(client *chattingClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	md := metadata.New(map[string]string{
		"user_id": strconv.Itoa(int(client.UserId)),
	})

	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := client.Cl.Logout(ctx, nil)
	if err != nil {
		log.Fatalf("client.Logout failed: %v", err)
		return err
	}

	client.UserId = 0

	return nil
}

func GetChatRoom(client *chattingClient) ([]*pb.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.Cl.GetChatRoom(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("client.GetChatRoom failed: %v", err)
		return nil, err
	}

	var chatRoomList []*pb.Room

	for {
		chatRoom, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("client.GetChatRoom failed: %v", err)
			return nil, err
		}

		chatRoomList = append(chatRoomList, chatRoom)
	}

	return chatRoomList, nil
}

func CreateRoom(client *chattingClient, roomName string) (int32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	room, err := client.Cl.CreateRoom(ctx, &pb.CreateRoomRequest{
		RoomName: roomName,
	})
	if err != nil {
		log.Fatalf("client.CreateRoom failed: %v", err)
		return 0, err
	}

	return room.RoomId, nil
}

func DeleteRoom(client *chattingClient, roomId int32) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Cl.RemoveRoom(ctx, &pb.RemoveRoomRequest{
		RoomId: roomId,
	})
	if err != nil {
		log.Fatalf("client.CreateRoom failed: %v", err)
		return err
	}

	return nil
}

func EnterChatRoom(client *chattingClient, roomId int32) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	md := metadata.New(map[string]string{
		"user_id": strconv.Itoa(int(client.UserId)),
	})

	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := client.Cl.EnterChatRoom(ctx, &pb.RoomRequest{RoomId: roomId})
	if err != nil {
		log.Fatalf("client.EnterChatRoom failed: %v", err)
		return err
	}

	client.RoomId = roomId

	return nil
}

func ExitChatRoom(client *chattingClient, roomNumber int32) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	md := metadata.New(map[string]string{
		"user_id": strconv.Itoa(int(client.UserId)),
		"room_id": strconv.Itoa(int(client.RoomId)),
	})

	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := client.Cl.ExitChatRoom(ctx, nil)
	if err != nil {
		log.Fatalf("client.ExitChatRoom failed: %v", err)
		return err
	}

	return nil
}

func Chatting(client *chattingClient) {
	ctx, cancel := context.WithCancel(context.Background())

	md := metadata.New(map[string]string{
		"user_id": strconv.Itoa(int(client.UserId)),
		"room_id": strconv.Itoa(int(client.RoomId)),
	})

	ctx = metadata.NewOutgoingContext(ctx, md)

	stream, err := client.Cl.Chatting(ctx)
	if err != nil {
		log.Fatalf("client.Chatting failed: %v", err)
	}

	waitc := make(chan struct{})

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("client.Chatting failed: %v", err)
			}
			fmt.Printf("|%v|[%v]|%v\n", client.RoomId, client.UserId, in.Msg)
		}
	}()

	for {
		reader := bufio.NewScanner(os.Stdin)
		fmt.Print("Chat: ")

		if reader.Scan() {
			input := reader.Text()

			token := strings.Split(input, " ")
			if len(token) == 0 {
				continue
			}

			cmd := strings.ToLower(token[0])

			if cmd == "/exit" {
				break
			}

			msg := pb.Message{
				Msg: input,
			}
			if err := stream.Send(&msg); err != nil {
				log.Fatalf("client.Chatting: stream.Send(%v) failed: %v", msg.Msg, err)
			}
		}
	}

	cancel()
	stream.CloseSend()
	<-waitc
}

package main

import (
	"context"
	"fmt"
	"time"

	pb "github.com/miska12345/DDPoll/ddpoll"
	"google.golang.org/grpc"
)

func authenticate(client pb.DDPollClient, username, password string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.DoAction(ctx, &pb.UserAction{
		Action:     pb.UserAction_Authenticate,
		Parameters: []string{username, password},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("login ok")
}

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial("localhost:8080", opts...)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	client := pb.NewDDPollClient(conn)
	authenticate(client, "admin", "666")
}

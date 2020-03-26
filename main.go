package main

import (
	"context"
	"fmt"
	"time"

	pb "github.com/miska12345/DDPoll/ddpoll"
	"google.golang.org/grpc"
)

func authenticate(client pb.DDPollClient, query *pb.AuthQuery) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.Authenticate(ctx, query)
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
	conn, err := grpc.Dial("localhost:8081", opts...)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	client := pb.NewDDPollClient(conn)
	authenticate(client, &pb.AuthQuery{
		Name:     "admin",
		Password: "666",
	})

}

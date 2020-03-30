package main

import (
	"context"
	"fmt"
	"time"

	pb "github.com/miska12345/DDPoll/ddpoll"
	"google.golang.org/grpc"
)

var authToken uint64

func establishStream(client pb.DDPollClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	config := new(pb.PollStreamConfig)
	config.RankBy = pb.PollStreamConfig_Time
	as, err := client.EstablishPollStream(ctx, config)
	for i := 0; i < 10; i++ {
		p, err := as.Recv()
		fmt.Println(p)
	}

	if err != nil {
		fmt.Println(err)
		return
	}
	authToken = as.GetToken()
	fmt.Println(authToken)
	fmt.Println("login ok")
}

func authenticate(client pb.DDPollClient, username, password string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	as, err := client.DoAction(ctx, &pb.UserAction{
		Action:     pb.UserAction_Authenticate,
		Parameters: []string{username, password},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	authToken = as.GetToken()
	fmt.Println(authToken)
	fmt.Println("login ok")
}

func createPoll(client pb.DDPollClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.DoAction(ctx, &pb.UserAction{
		Header: &pb.UserAction_Header{
			Username: "admin",
			Token:    authToken,
		},
		Action:     pb.UserAction_Create,
		Parameters: []string{"title", "content", "category", "true", "cookie", "cat"},
	})
	if err != nil {
		fmt.Println(err)
	}
}

func createUser(client pb.DDPollClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.DoAction(ctx, &pb.UserAction{
		Header: &pb.UserAction_Header{
			Username: "admin",
			Token:    authToken,
		},
		Action:     pb.UserAction_Registeration,
		Parameters: []string{"fuckj", "fff"},
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	_, err2 := client.DoAction(ctx, &pb.UserAction{
		Header: &pb.UserAction_Header{
			Username: "admin",
			Token:    authToken,
		},
		Action:     pb.UserAction_Registeration,
		Parameters: []string{"fuckj", "fff"},
	})

	if err2 != nil {
		fmt.Println(err.Error())
	}

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
	//createPoll(client)

	createUser(client)

}

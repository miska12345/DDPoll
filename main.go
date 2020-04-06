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
	ctx := context.Background()
	config := new(pb.PollStreamConfig)
	config.RankBy = pb.PollStreamConfig_Time
	as, _ := client.EstablishPollStream(ctx, config)
	defer as.CloseSend()
	for {
		p, err := as.Recv()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(p)
		}
	}
}

func authenticate(client pb.DDPollClient, username, password string) {
	ctx := context.Background()
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

// Takes a group of group ids, and username to create poll room
// Returns a passphrase to join poll room and error idicator
func createPollRoom(client pb.DDPollClient, username string, gids []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	as, err := client.DoAction(ctx, &pb.UserAction{
		Header: &pb.UserAction_Header{
			Username: "didntpay",
			Token:    authToken,
		},
		Action:     pb.UserAction_StartGroupPoll,
		Parameters: append([]string{username}, gids...),
	})
	return string(as.GetInfo()), err
}

func sendCommand(client pb.DDPollClient, roomKey string, command pb.Next_PollControl) error {
	sc, err := client.EstablishClientStream(context.Background())
	if err != nil {
		return err
	}
	err = sc.Send(&pb.Next{
		RoomKey: roomKey,
		Signal:  command,
	})
	return err
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

func testAuth(client pb.DDPollClient, n int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// for i := 0; i < n; i++ {
	// 	_, err := client.DoAction(ctx, &pb.UserAction{
	// 		Header: &pb.UserAction_Header{
	// 			Username: "admin",
	// 			Token:    authToken,
	// 		},
	// 		Action:     pb.UserAction_Registeration,
	// 		Parameters: []string{string(97 + i), "fff"},
	// 	})
	// 	if err != nil {
	// 		fmt.Println(err.Error())
	// 	}
	// }

	for i := 0; i < n; i++ {
		_, err := client.DoAction(ctx, &pb.UserAction{
			Action:     pb.UserAction_Authenticate,
			Parameters: []string{string(97 + i), "fff"},
		})

		if err != nil {
			fmt.Println(err.Error())
		}
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

	// authenticate(client, "admin", "666")
	// authenticate(client, "didntpay", "password")
	//createPoll(client)
	//createUser(client)
	roomKey, err := createPollRoom(client, "didntpay", []string{"22003698", "17129137"})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(roomKey)
	testAuth(client, 100)
}

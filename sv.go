package main

import (
	"fmt"

	"github.com/miska12345/DDPoll/server"
)

func main() {
	g, _ := server.Run("8080", 100, "mongodb+srv://ddpoll:ddpoll@test-ycw1l.mongodb.net/test?retryWrites=true&w=majority", "testDB", "mongodb+srv://ddpoll:ddpoll@test-ycw1l.mongodb.net/test?retryWrites=true&w=majority", "testDB")
	for {
		var stop string
		fmt.Scanf("%s", &stop)
		break
	}
	g.Stop()
	fmt.Println("Server shutdown...")
}

package main

import (
	"fmt"
	"time"

	"github.com/miska12345/DDPoll/server"
)

func main() {
	g, _ := server.Run("8080", 100, "mongodb+srv://ddpoll:ddpoll@test-ycw1l.mongodb.net/test?retryWrites=true&w=majority", "Polls", "mongodb+srv://ddpoll:ddpoll@test-ycw1l.mongodb.net/test?retryWrites=true&w=majority", "Users")
	time.Sleep(5 * time.Second)
	g.Stop()
	fmt.Println("Server shutdown...")
}

package main

import (
	"github.com/miska12345/DDPoll/server"
)

func main() {
	server.Run("8080", 100, "mongodb+srv://ddpoll:ddpoll@test-ycw1l.mongodb.net/test?retryWrites=true&w=majority", "Polls", "mongodb+srv://ddpoll:ddpoll@test-ycw1l.mongodb.net/test?retryWrites=true&w=majority", "Users")
}

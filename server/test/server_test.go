package svtest

import (
	"testing"
	"google.golang.org/grpc"
)

const dbLinkPoll = "mongodb+srv://ddpoll:ddpoll@test-ycw1l.mongodb.net/test?retryWrites=true&w=majority"
const dbLinkUser = "mongodb+srv://ddpoll:ddpoll@test-ycw1l.mongodb.net/test?retryWrites=true&w=majority"
const dbPollName = "Polls"
const dbUserName = "Users"
const testPort = "8080"

func initializeTestEnv() *grpc.Server, error {
	g, err := server.Run(testPort, 100, dbLinkPoll, dbPollName, dbLinkUser, dbUserName)
	if err != nil {
		return nil, err
	}
	return g, nil
}




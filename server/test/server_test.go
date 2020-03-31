package svtest

import (
	"context"
	"sync"
	"testing"
	"time"

	pb "github.com/miska12345/DDPoll/ddpoll"
	"github.com/miska12345/DDPoll/server"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const dbLinkPoll = "mongodb+srv://ddpoll:ddpoll@test-ycw1l.mongodb.net/test?retryWrites=true&w=majority"
const dbLinkUser = "mongodb+srv://ddpoll:ddpoll@test-ycw1l.mongodb.net/test?retryWrites=true&w=majority"
const dbPollName = "Polls"
const dbUserName = "Users"
const testPort = "8080"
const admin = "admin"
const adPass = "666"

func initializeTestEnv() (*grpc.Server, *grpc.ClientConn, error) {
	g, err := server.Run(testPort, 100, dbLinkPoll, dbPollName, dbLinkUser, dbUserName)
	if err != nil {
		return nil, nil, err
	}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial("localhost:8080", opts...)
	if err != nil {
		return nil, nil, err
	}
	return g, conn, nil
}

func TestBasicConnection(t *testing.T) {
	s, con, err := initializeTestEnv()
	assert.Nil(t, err)
	defer con.Close()
	defer s.Stop()
}

func TestAuthentication(t *testing.T) {
	s, con, err := initializeTestEnv()
	assert.Nil(t, err)
	defer con.Close()
	defer s.Stop()

	c := pb.NewDDPollClient(con)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = c.DoAction(ctx, &pb.UserAction{
		Action:     pb.UserAction_Authenticate,
		Parameters: []string{admin, adPass},
	})
	assert.Nil(t, err)
}

func TestCreatePoll(t *testing.T) {
	s, con, err := initializeTestEnv()
	assert.Nil(t, err)
	defer con.Close()
	defer s.Stop()

	c := pb.NewDDPollClient(con)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sum, err := c.DoAction(ctx, &pb.UserAction{
		Action:     pb.UserAction_Authenticate,
		Parameters: []string{admin, adPass},
	})
	assert.Nil(t, err)

	token := sum.GetToken()
	_, err = c.DoAction(ctx, &pb.UserAction{
		Header: &pb.UserAction_Header{
			Username: admin,
			Token:    token,
		},
		Action:     pb.UserAction_Create,
		Parameters: []string{"title", "context", "cat", "true", "A", "B"},
	})
	assert.Nil(t, err)
}

func TestStreamPolls(t *testing.T) {
	s, con, err := initializeTestEnv()
	assert.Nil(t, err)
	defer con.Close()
	defer s.Stop()

	c := pb.NewDDPollClient(con)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = c.DoAction(ctx, &pb.UserAction{
		Action:     pb.UserAction_Authenticate,
		Parameters: []string{admin, adPass},
	})
	assert.Nil(t, err)

	stream, err := c.EstablishPollStream(ctx, &pb.PollStreamConfig{
		RankBy: pb.PollStreamConfig_Time,
	})
	assert.Nil(t, err)
	for i := 0; i < 20; i++ {
		_, err := stream.Recv()
		assert.Nil(t, err)
	}
}

func TestStreamPollsMultiple(t *testing.T) {
	s, con, err := initializeTestEnv()
	assert.Nil(t, err)
	defer con.Close()
	defer s.Stop()

	var wg sync.WaitGroup
	numClients := 100
	wg.Add(numClients)
	for i := 0; i < numClients; i++ {
		go func() {
			c := pb.NewDDPollClient(con)
			ctx := context.Background()
			stream, err := c.EstablishPollStream(ctx, &pb.PollStreamConfig{
				RankBy: pb.PollStreamConfig_Time,
			})
			assert.Nil(t, err)
			for i := 0; i < 20; i++ {
				_, err := stream.Recv()
				assert.Nil(t, err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestVoteMultiple(t *testing.T) {
	s, con, err := initializeTestEnv()
	assert.Nil(t, err)
	defer con.Close()
	defer s.Stop()

	c := pb.NewDDPollClient(con)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sum, err := c.DoAction(ctx, &pb.UserAction{
		Action:     pb.UserAction_Authenticate,
		Parameters: []string{admin, adPass},
	})
	assert.Nil(t, err)

	token := sum.GetToken()

	sum2, err := c.DoAction(ctx, &pb.UserAction{
		Header: &pb.UserAction_Header{
			Username: admin,
			Token:    token,
		},
		Action:     pb.UserAction_Create,
		Parameters: []string{"title", "context", "cat", "true", "A", "B"},
	})
	assert.Nil(t, err)
	id := string(sum2.Info)
	_, err = c.DoAction(ctx, &pb.UserAction{
		Header: &pb.UserAction_Header{
			Username: admin,
			Token:    token,
		},
		Action:     pb.UserAction_VoteMultiple,
		Parameters: []string{id, "1", "0"},
	})
	assert.Nil(t, err)
}

// Package server provides an interface for DDPoll server
package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/miska12345/DDPoll/db"
	pb "github.com/miska12345/DDPoll/ddpoll"
	goLogger "github.com/phachon/go-logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var logger *goLogger.Logger

// server is a single instance of a server node - a serving entity
type server struct {
	pb.UnimplementedDDPollServer
	maxConnection int
}

func Run(port string, maxConnection int) error {
	logger = goLogger.NewLogger()
	logger.Detach("console")

	// console adapter config
	consoleConfig := &goLogger.ConsoleConfig{
		Color:      true,  // Does the text display the color
		JsonFormat: false, // Whether or not formatted into a JSON string
		Format:     "",    // JsonFormat is false, logger message output to console format string
	}

	// add output to the console
	logger.Attach("console", goLogger.LOGGER_LEVEL_DEBUG, consoleConfig)

	logger.Infof("Server is starting at port %s", port)
	ls, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	logger.Info("Server is running!")
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterDDPollServer(grpcServer, newServer(maxConnection))
	err = grpcServer.Serve(ls)
	return err
}

func newServer(maxConnection int) *server {
	s := new(server)

	// Initialize server struct
	s.maxConnection = maxConnection
	return s
}

func connectToPollsDB(URL, username, password, database, collectionName string) (dbPoll *db.PollDB, err error) {
	// TODO: Add params when release
	dbConn, err := db.Dial(URL, 2*time.Second, 5*time.Second)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	return dbConn.ToPollsDB(database, collectionName, ""), nil
}

// Authenticate verifies user login credentials
func (s *server) authenticate(username, password string) error {
	// Database stuff for authentication

	// REMOVE
	if username == "admin" && password == "666" {
		return nil
	}
	return status.Error(codes.InvalidArgument, "Authentication Failed")
}

// DoAuthenticate check the provided params and authenticate the user
func (s *server) doAuthenticate(ctx context.Context, params []string) (as *pb.ActionSummary, err error) {
	if len(params) < 2 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Expect %d but receive %d parameters for authentication", 2, len(params)))
	}
	username := params[0]
	password := params[1]
	// TODO: Do username format check(i.e. not empty, contains no special character etc)

	// Call our internal authentication routine
	err = s.authenticate(username, password)
	if err != nil {
		return
	}

	// Associate current context with the particular user
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Errorf("metadata from context failed, action aborted")
		return nil, status.Error(codes.Internal, "Internal error")
	}
	md["username"] = make([]string, 1)
	md["username"][0] = params[0]

	return &pb.ActionSummary{
		Status: pb.Status_OK,
	}, nil
}

// DoAction takes UserAction request and distribute into sub-routines for processing
func (s *server) DoAction(ctx context.Context, action *pb.UserAction) (as *pb.ActionSummary, err error) {
	switch action.GetAction() {
	case pb.UserAction_Authenticate:
		as, err = s.doAuthenticate(ctx, action.GetParameters())
	case pb.UserAction_Create:
		as, err = s.doCreatePoll(ctx, action.GetParameters())
	default:
		logger.Warningf("Unknown action type %s", action.GetAction().String())
		err = status.Error(codes.NotFound, fmt.Sprintf("Unknown action [%s]", action.GetAction().String()))
	}
	return
}

// EstablishPollStream takes polls config and stream polls to the user
func (s *server) EstablishPollStream(config *pb.PollStreamConfig, stream pb.DDPoll_EstablishPollStreamServer) error {
	// stream.Context() to get context
	panic("not implemented")
}

// FindPollByKeyWord takes a set of search criterias and return a collection of Polls
// Maximum number of polls returned is set by models.SEARCH_MAX_RESULT
func (s *server) FindPollByKeyWord(ctx context.Context, q *pb.SearchQuery) (*pb.SearchResp, error) {
	panic("not implemented")
}

/*********************************************************************************************************************************************************/

func (s *server) doCreatePoll(ctx context.Context, params []string) (as *pb.ActionSummary, err error) {
	/*
		if len(params) < models.REQUIRED_POLL_ELEMENTS+models.MIN_OPTIONS {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Expect %d but receive %d parameters for authentication", 2, len(params)))
		}
	*/
	db, err := connectToPollsDB(
		"mongodb+srv://admin:wassup@cluster0-n0w7a.mongodb.net/test?retryWrites=true&w=majority",
		"admin",
		"wassup",
		"DDPoll",
		"Polls",
	)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Error(codes.Internal, fmt.Sprintf("Cannot connect to poll database"))
	}
	/*
		owner := params[0]
		accessibility := params[1]
		title := params[2]
		body := params[3]
		category := params[4]
		optionNum := len(params) - models.REQUIRED_POLL_ELEMENTS
		options := params[7:]
	*/

	// TODO: Use db.CreatePoll to create poll...
	// If return is a string(not empty) than ok
	db.CreatePoll("miska", "title", "content", "cat", true, time.Hour, []string{"A", "B"})
	return &pb.ActionSummary{
		Status: 1,
	}, nil
}

// 	// TODO: Do username format check(i.e. not empty, contains no special character etc)

// 	// Call our internal authentication routine
// 	err = createPoll(owner, title, content, category, accessibility, options)
// 	if err != nil {
// 		return
// 	}

// 	// Associate current context with the particular user
// 	md, ok := metadata.FromIncomingContext(ctx)
// 	if !ok {
// 		logger.Errorf("metadata from context failed, action aborted")
// 		return nil, status.Error(codes.Internal, "Internal error")
// 	}
// 	md["username"] = make([]string, 1)
// 	md["username"][0] = params[0]

// 	return &pb.ActionSummary{
// 		Status: pb.Status_OK,
// 	}, nil
// }

// TO-DO
// func createPoll(host, title, content, category string, accessbility int8, choices []string) *poll.Poll {
// 	p := new(poll.Poll)

// 	// Initialize poll struct
// 	p.Owner = host
// 	p.Title = title
// 	p.Body = content
// 	p.Accessibility = accessbility
// 	p.Choices = choices
// 	p.Counts = make([]int64, len(choices))
// 	return p
// }

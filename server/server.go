// Package server provides an interface for DDPoll server
package server

import (
	"context"
	"fmt"
	"net"

	pb "github.com/miska12345/DDPoll/ddpoll"
	models "github.com/miska12345/DDPoll/models"
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

// Authenticate verifies user login credentials
func (s *server) authenticate(username, password string) error {
	// Database stuff for authentication

	// REMOVE
	if username == "admin" && password == "666" {
		return nil
	}
	s.doCreatePoll(nil, nil)
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

// Create Poll
func (s *server) doCreatePoll(ctx context.Context, params []string) (as *pb.ActionSummary, id int64, err error) {
	if len(params) < 2 {
		return nil, -1, status.Error(codes.InvalidArgument, fmt.Sprintf("Expect %d but receive %d parameters for authentication", 2, len(params)))
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
		return nil, -1, status.Error(codes.Internal, "Internal error")
	}
	md["username"] = make([]string, 1)
	md["username"][0] = params[0]

	return &pb.ActionSummary{
		Status: pb.Status_OK,
	}, 0, nil
}

func createPoll(host string, members []string, title, content string, accessbility int8, choices []string) *models.Poll {
	p := new(models.Poll)

	// Initialize poll struct
	p.HOST = host
	p.MEMBERS = members
	p.TITLE = title
	p.CONTENT = content
	p.ACCESSIBLITY = accessbility
	p.CHOICES = choices
	p.COUNTS = make([]int64, len(choices))
	return p
}

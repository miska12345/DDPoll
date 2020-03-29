// Package server provides an interface for DDPoll server
package server

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/miska12345/DDPoll/db"
	pb "github.com/miska12345/DDPoll/ddpoll"
	"github.com/miska12345/DDPoll/poll"
	goLogger "github.com/phachon/go-logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var logger *goLogger.Logger

// server is a single instance of a server node - a serving entity
type server struct {
	pb.UnimplementedDDPollServer
	maxConnection int
	pollsDB       *db.PollDB
	usersDB       *db.UserDB
}

type networkClient struct {
	userid         string
	username       string
	startTime      time.Time
	lastActiveTime time.Time
	sync.Mutex
}

type uSessionsTable struct {
	table map[string]networkClient
	sync.Mutex
}

// Run starts running the server
func Run(port string, maxConnection int, pollsDBURL, pollsBase string, userDBURL, usersBase string) error {
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

	pollsDB, perr := connectToPollsDB(pollsDBURL, pollsBase, "Polls")
	usersDB, uerr := connectToUsersDB(userDBURL, usersBase, "Users")
	if perr != nil {
		logger.Error(err.Error())
		return perr
	} else if uerr != nil {
		logger.Error(err.Error())
		return uerr
	}
	logger.Info("Server is running!")

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterDDPollServer(grpcServer, newServer(maxConnection, pollsDB, usersDB))
	err = grpcServer.Serve(ls)
	return err
}

func newServer(maxConnection int, pdb *db.PollDB, udb *db.UserDB) *server {
	s := new(server)

	// Initialize server struct
	s.maxConnection = maxConnection
	//s.uSessionsTable = make(map[string]networkClient)
	s.pollsDB = pdb
	s.usersDB = udb
	return s
}

func connectToPollsDB(URL, database string, collectionNames ...string) (dbPoll *db.PollDB, err error) {
	if len(collectionNames) < 1 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("expect 1 parameter but received %d", len(collectionNames)))
	}
	dbConn, err := db.Dial(URL, 2*time.Second, 5*time.Second)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	return dbConn.ToPollsDB(database, collectionNames[0], ""), nil
}

//TODO add comment
func connectToUsersDB(URL, database, collectionName string) (dbPoll *db.UserDB, err error) {
	// TODO: Add params when release
	dbConn, err := db.Dial(URL, 2*time.Second, 5*time.Second)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	return dbConn.ToUserDB(database, collectionName, ""), nil
}

// Authenticate verifies user login credentials
func (s *server) authenticate(username, password string) (string, error) {
	// Database stuff for authentication

	// REMOVE
	if username == "admin" && password == "666" {
		return "fakeuid", nil
	}
	return "", status.Error(codes.InvalidArgument, "Authentication Failed")
}

// DoAuthenticate check the provided params and authenticate the user
func (s *server) doAuthenticate(ctx context.Context, params []string) (as *pb.ActionSummary, err error) {
	if len(params) < 2 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Expect %d but receive %d parameters for authentication", 2, len(params)))
	}
	username := params[uParamsUsername]
	password := params[uParamsPassword]
	// TODO: Do username format check(i.e. not empty, contains no special character etc)

	// Call our internal authentication routine
	uid, err := s.authenticate(username, password)
	if err != nil {
		logger.Debugf("Useer %s failed to login because err = %s", username, err.Error)
		return
	}

	/*
		// Associate current context with the particular user
		s.uSessionsTable[uid] = networkClient{
			userid:         uid,
			username:       username,
			startTime:      time.Now(),
			lastActiveTime: time.Now(),
		}
	*/
	logger.Debugf("User %s logged in", username)
	return &pb.ActionSummary{
		Info: []byte(uid), // TODO: Update status
	}, nil
}

// DoAction takes UserAction request and distribute into sub-routines for processing
func (s *server) DoAction(ctx context.Context, action *pb.UserAction) (as *pb.ActionSummary, err error) {
	switch action.GetAction() {
	case pb.UserAction_Authenticate:
		as, err = s.doAuthenticate(ctx, action.GetParameters())
	case pb.UserAction_Create:
		as, err = s.doCreatePoll(ctx, action.GetParameters())
	case pb.UserAction_VoteMultiple:
		// as, err = s.doVoteMultiple(ctx, action.GetParameters())
	case pb.UserAction_Registeration:
		//as, err = s.doRegistration(ctx, action.GetParameters())
	default:
		logger.Warningf("Unknown action type %s", action.GetAction().String())
		err = status.Error(codes.NotFound, fmt.Sprintf("Unknown action [%s]", action.GetAction().String()))
	}
	return
}

// Establish account for user with unique usernames
func (s *server) doRegistration(ctx context.Context, params []string) (as *pb.ActionSummary, err error) {
	if len(params) < 2 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Expect %d but receive %d parameters for registration", 2, len(params)))
	}
	username := params[uParamsUsername]
	password := params[uParamsPassword]

	usernamecheck, err := s.usersDB.GetUserByName(username)

	if usernamecheck == nil {
		_, creationErr := s.usersDB.CreateNewUser(username, password)

		if creationErr != nil {
			logger.Debug(creationErr.Error())
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "User name is taken")
	}

	return &pb.ActionSummary{
		Info: []byte("New user" + username + "registered at time" + time.Now().String()), // TODO:
	}, nil
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
	logger.Debugf("Create started...len(params)=%d", len(params))
	if len(params) < poll.CreateParamLength {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Expect %d but receive %d parameters for create", poll.RequiredPollElements, len(params)))
	}
	options := params[poll.RequiredPollElements:]
	public, err := strconv.ParseBool(params[uParamsPublic])
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Argument [public] is not type true/false"))
	}

	id, err := s.pollsDB.CreatePoll("miska", params[uParamsTopic], params[uParamsContext], params[uParamsCategory], public, time.Hour, options)
	if err != nil || len(id) == 0 {
		logger.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Failed to create poll, error=%s", err))
	}
	// If return is a string(not empty) than ok
	return &pb.ActionSummary{
		Info: []byte(id),
	}, nil
}

/*********************************************************************************************************************************************************/

/*
func (s *server) doVoteMultiple(ctx context.Context, params []string) (as *pb.ActionSummary, err error) {
	db, err := connectToPollsDB(
		"mongodb+srv://admin:wassup@cluster0-n0w7a.mongodb.net/test?retryWrites=true&w=majority",
		"admin",
		"wassup",
		"DDPoll",
		"Polls",
	)

}
*/

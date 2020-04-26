// Package server provides an interface for DDPoll server
package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
	random "math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/miska12345/DDPoll/db"
	"github.com/miska12345/DDPoll/ddpoll"
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
	pollsDB       *db.PollDB
	usersDB       *db.UserDB
	maxConnection int
	authSalt      []byte
	pollgroups    pollRooms
}

type pollRooms struct {
	rooms map[string]*pollgroup
	sync.Mutex
}

type pollgroup struct {
	creator     string
	members     map[string]bool
	numEnrolled uint32
	canVote     bool
	gids        []uint32
	currentPoll *pb.Poll
	sync.Mutex
}

// Run starts running the server
func Run(port string, maxConnection int, pollsDBURL, pollsBase string, userDBURL, usersBase string) (grpcServer *grpc.Server, err error) {
	logger = goLogger.NewLogger()
	logger.Detach("console")

	// console adapter config
	consoleConfig := &goLogger.ConsoleConfig{
		Color:      false, // Does the text display the color
		JsonFormat: false, // Whether or not formatted into a JSON string
		Format:     "",    // JsonFormat is false, logger message output to console format string
	}

	// add output to the console
	logger.Attach("console", goLogger.LOGGER_LEVEL_DEBUG, consoleConfig)

	logger.Infof("Server is starting at port %s", port)
	ls, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		logger.Error(err.Error())
		return
	}

	pollsDB, perr := connectToPollsDB(pollsDBURL, pollsBase, "poll")
	usersDB, uerr := connectToUsersDB(userDBURL, usersBase, "users")
	if perr != nil || uerr != nil {
		logger.Error("Failed to initialize database connections")
		return
	}

	logger.Info("Server is running!")

	var opts []grpc.ServerOption
	grpcServer = grpc.NewServer(opts...)
	pb.RegisterDDPollServer(grpcServer, newServer(maxConnection, pollsDB, usersDB))
	go grpcServer.Serve(ls)
	return
}

func newServer(maxConnection int, pdb *db.PollDB, udb *db.UserDB) *server {
	s := new(server)

	// Initialize server struct
	s.pollsDB = pdb
	s.usersDB = udb
	s.maxConnection = maxConnection
	s.authSalt = make([]byte, 128)
	s.pollgroups.rooms = make(map[string]*pollgroup)
	rand.Read(s.authSalt)
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

// Authenticate verifies user login credentials and returns uid
func (s *server) authenticate(username, password string) (err error) {
	// REMOVE
	if username == "admin" && password == "666" {
		return nil
	}

	// Database stuff for authentication
	h := sha1.New()

	var submittedcred []byte

	authsalt, matchingcred, getErr := s.usersDB.GetUserAuthSaltAndCred(username)
	if getErr != nil {
		//error whiling getting salt
		err = getErr
		return
	}
	h.Write([]byte(password))
	h.Write(authsalt)
	submittedcred = h.Sum(nil)

	if bytes.Compare(submittedcred, matchingcred) == 0 {
		return nil
	}

	return status.Error(codes.InvalidArgument, "Authentication Failed")
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
	err = s.authenticate(username, password)
	if err != nil {
		logger.Debugf("User %s failed to login because err = %s", username, err.Error())
		return
	}
	token := s.generateAuthToken(username)
	logger.Infof("User %s logged in, token: %v", username, token)
	return &pb.ActionSummary{
		Info:  []byte(username),
		Token: token,
	}, nil
}

func (s *server) generateAuthToken(username string) uint64 {
	h := xxhash.New64()
	r := bytes.NewReader(append(s.authSalt, []byte(username)...))
	io.Copy(h, r)
	return h.Sum64()
}

func (s *server) verifyAuthToken(token uint64, username string) bool {
	logger.Debugf("generated token is %v", s.generateAuthToken(username))
	return token == s.generateAuthToken(username)
}

// DoAction takes UserAction request and distribute into sub-routines for processing
func (s *server) DoAction(ctx context.Context, action *pb.UserAction) (as *pb.ActionSummary, err error) {
	as = &pb.ActionSummary{}
	// if action.GetAction() != pb.UserAction_Authenticate {
	// 	if !s.verifyAuthToken(action.GetHeader().GetToken(), action.GetHeader().GetUsername()) {
	// 		err = status.Error(codes.Unauthenticated, "Token is invalid")
	// 		return
	// 	}
	// }
	logger.Debugf("%s", action.GetAction())
	switch action.GetAction() {
	case pb.UserAction_Authenticate:
		as, err = s.doAuthenticate(ctx, action.GetParameters())
	case pb.UserAction_Create:
		as, err = s.doCreatePoll(ctx, append([]string{action.Header.GetUsername()}, action.GetParameters()...))
	case pb.UserAction_VoteMultiple:
		as, err = s.doVoteMultiple(ctx, action.GetParameters())
	case pb.UserAction_Registeration:
		as, err = s.doRegistration(ctx, action.GetParameters())
		//TODO: print action summary
	case pb.UserAction_StartGroupPoll:
		as, err = s.doStartPollGroup(ctx, action.GetParameters())
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
	as = &pb.ActionSummary{Info: []byte("unusuall exit")}

	username := params[uParamsUsername]
	password := params[uParamsPassword]

	if _, errcreate := s.usersDB.CreateNewUser(username, password); errcreate != nil {
		err = errcreate
		logger.Debug(err.Error())
		return
	}

	return &pb.ActionSummary{
		Info: []byte("New user" + username + "registered at time" + time.Now().String()), // TODO:
	}, nil
}

// EstablishPollStream takes polls config and stream polls to the user
func (s *server) EstablishPollStream(config *pb.PollStreamConfig, stream pb.DDPoll_EstablishPollStreamServer) error {
	// stream.Context() to get context

	for {
		ch, errGetPolls := s.pollsDB.GetNewestPolls(GET_POLL_NUM)
		if errGetPolls != nil {
			logger.Errorf("[GetPolls] %s", errGetPolls)
			return errGetPolls
		}
		serverP, ok := <-ch
		for ok {
			clientP := new(ddpoll.Poll)
			clientP.Body = serverP.Content
			clientP.Category = serverP.Category
			clientP.Id = serverP.PID
			clientP.DisplayType = pb.Poll_OnReveal
			clientP.Options = serverP.Choices
			clientP.Owner = serverP.Owner
			clientP.Stars = serverP.Stars
			clientP.Tags = serverP.Tags
			if errSend := stream.Send(clientP); errSend != nil {
				logger.Errorf("[SendPolls] %s", errSend)
				return errSend
			}
			serverP, ok = <-ch
		}
	}
}

// FindPollByKeyWord takes a set of search criterias and return a collection of Polls
// Maximum number of polls returned is set by models.SEARCH_MAX_RESULT
func (s *server) FindPollByKeyWord(ctx context.Context, q *pb.SearchQuery) (*pb.SearchResp, error) {
	panic("not implemented")
}

/*********************************************************************************************************************************************************/

func (s *server) doCreatePoll(ctx context.Context, params []string) (as *pb.ActionSummary, err error) {
	logger.Debugf("Create started with params=%v", params)
	if len(params) < poll.CreateParamLength {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Expect %d but receive %d parameters for create", poll.RequiredPollElements, len(params)))
	}
	options := params[poll.RequiredPollElements:]
	public, err := strconv.ParseBool(params[uParamsPublic])
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Argument [public] is not type true/false"))
	}

	id, err := s.pollsDB.CreatePoll(params[uParamsUsername], params[uParamsTopic], params[uParamsContext], params[uParamsCategory], public, time.Hour, options)
	if err != nil || len(id) == 0 {
		logger.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Failed to create poll, error=%s", err))
	}
	logger.Debugf("New poll created! ID:%s", id)
	// If return is a string(not empty) than ok
	return &pb.ActionSummary{
		Info: []byte(id),
	}, nil
}

/*********************************************************************************************************************************************************/

func (s *server) doVoteMultiple(ctx context.Context, params []string) (as *pb.ActionSummary, err error) {
	if len(params) < VOTE_MUL_PARAM_NUM {
		return &pb.ActionSummary{}, status.Error(codes.InvalidArgument, fmt.Sprintf("Expect %d but receive %d parameters for registration", 2, len(params)))
	}

	pid := params[uParamsPollID]
	sVotes := params[uParamsPollID+1:]
	votes := make([]uint64, len(sVotes))
	for idx, val := range sVotes {
		n, err := strconv.ParseInt(val, 10, 64)
		votes[idx] = uint64(n)
		if err != nil {
			return nil, err
		}
	}
	err = s.pollsDB.UpdateNumVoted(pid, votes)
	return &pb.ActionSummary{}, err
}

/*********************************************************************************************************************************************************/

func (s *server) doStartPollGroup(ctx context.Context, params []string) (*pb.ActionSummary, error) {
	if len(params) < START_PG_PARAM_NUM {
		logger.Error("Invalid Argument Error")
		return &pb.ActionSummary{}, status.Error(codes.InvalidArgument, fmt.Sprintf("Expect %d but receive %d parameters for registration", START_PG_PARAM_NUM, len(params)))
	}
	// Randomized random generated integer
	random.Seed(time.Now().UnixNano())
	var roomKey string
	pg := new(pollgroup)
	s.pollgroups.Lock()
	for {
		// Generate random string
		var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
		b := make([]rune, 4)
		for i := range b {
			b[i] = letterRunes[random.Intn(len(letterRunes))]
		}
		roomKey = string(b)

		// Assign roomKey to pollroom
		if _, ok := s.pollgroups.rooms[roomKey]; !ok {
			s.pollgroups.rooms[roomKey] = pg
			break
		}
	}
	s.pollgroups.Unlock()
	// TODO: Add lock
	// Initilize poll room
	pg.creator = params[0]
	strGids := params[1:]
	for _, val := range strGids {
		gid, _ := strconv.ParseInt(val, 10, 64)
		pg.gids = append(pg.gids, uint32(gid))
	}

	return &pb.ActionSummary{
		Info: []byte(roomKey),
	}, error(nil)
}

func (s *server) doStopPollGroup(ctx context.Context, params []string) (as *pb.ActionSummary, err error) {
	if len(params) < 1 {
		logger.Error("Invalid Argument Error")
		return &pb.ActionSummary{}, status.Error(codes.InvalidArgument, fmt.Sprintf("Expect %d but receive %d parameters for registration", START_PG_PARAM_NUM, len(params)))
	}
	s.pollgroups.Lock()
	delete(s.pollgroups.rooms, params[0])
	return &pb.ActionSummary{}, error(nil)
}

// first "empty" command from host is needed for initializing the poll room
func (s *server) EstablishClientStream(srv pb.DDPoll_EstablishClientStreamServer) error {
	nextCommand, err := srv.Recv()
	roomKey := nextCommand.GetRoomKey()
	pg := s.pollgroups.rooms[roomKey]
	pg.members = make(map[string]bool)
	uid := pg.creator
	var pids []string
	for _, gid := range pg.gids {
		tempPIDs, err := s.usersDB.GetUserPollsByGroup(uid, gid)
		if err != nil {
			logger.Errorf("%s, uid: %s, gid: %s", err, uid, string(gid))
		}
		pids = append(pids, tempPIDs...)
	}
	if len(pids) == 0 {
		logger.Errorf("PID not found")
		return srv.SendAndClose(&pb.ActionSummary{})
	}
	// Index in poll array to determine next poll
	idx := 0

	// Helper method to convert DB poll to RPC poll
	var DB2RPC = func(nextPoll *poll.Poll, currentPoll *pb.Poll) {
		currentPoll.Body = nextPoll.Content
		currentPoll.Category = nextPoll.Category
		if nextPoll.DisplayType == 0 {
			currentPoll.DisplayType = pb.Poll_OnVote
		} else {
			currentPoll.DisplayType = pb.Poll_OnReveal
		}
		currentPoll.Id = nextPoll.PID
		currentPoll.Options = nextPoll.Choices
		currentPoll.Stars = nextPoll.Stars
		currentPoll.Tags = nextPoll.Tags
	}

	// Get DB poll
	nextPoll, err := s.pollsDB.GetPollByPID(pids[idx])
	if err != nil {
		logger.Infof("%s, pid: %s", err, pids[idx])
	}

	// Convert DB poll to RPC poll
	pg.currentPoll = new(pb.Poll)
	DB2RPC(nextPoll, pg.currentPoll)
	logger.Infof("current pid: %s", pg.currentPoll)
	for {
		nextCommand, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Errorf("Receive Error: %s", err)
			return err
		}
		// Block and wait for next command
		switch nextCommand.GetSignal() {
		case pb.Next_foward:
			if idx < len(pids)-1 {
				pg.Lock()
				idx++
				nextPoll, err := s.pollsDB.GetPollByPID(pids[idx])
				if err != nil {
					pg.Unlock()
					logger.Errorf("Get PID Error: %s", err)
					return err
				}
				// Convert DB poll to RPC poll
				DB2RPC(nextPoll, pg.currentPoll)
				pg.Unlock()
			}
		case pb.Next_backward:
			if idx > 0 {
				pg.Lock()
				idx--
				nextPoll, err := s.pollsDB.GetPollByPID(pids[idx])
				if err != nil {
					pg.Unlock()
					logger.Errorf("Get PID Error: %s", err)
					return err
				}
				// Convert DB poll to RPC poll
				DB2RPC(nextPoll, pg.currentPoll)
				pg.Unlock()
			}
		case pb.Next_start:
			pg.Lock()
			pg.canVote = true
			pg.Unlock()
		case pb.Next_stop:
			pg.Lock()
			pg.canVote = false
			pg.Unlock()
		case pb.Next_terminateGroup:
			return srv.SendAndClose(&pb.ActionSummary{})
		}
		logger.Infof("current pid: %s", pg.currentPoll)
	}
	return err
}

// RPC serivce for a client joining a poll group
func (s *server) JoinPollGroup(req *pb.JoinPollQuery, stream pb.DDPoll_JoinPollGroupServer) error {
	roomKey := req.GetPhrase()
	logger.Infof("%s has joined the room", req.GetDisplayName())
	// Initialize first poll
	pollroom := s.pollgroups.rooms[roomKey]
	pollroom.Lock()
	clientP := pollroom.currentPoll
	if clientP == nil {
		logger.Error("Poll is empty")
	}

	if _, ok := pollroom.members[req.DisplayName]; ok {
		logger.Infof("DisplayName Occupied")
		return os.ErrClosed
	}
	pollroom.members[req.DisplayName] = false
	logger.Infof("Participants count: %d", len(pollroom.members))
	logger.Infof("Current poll %s", pollroom.currentPoll)

	for {
		// Checking for poll update every 500 miliseconds
		time.Sleep(500 * time.Millisecond)
		pollroom.Lock()
		clientP = pollroom.currentPoll

		if errSend := stream.Send(clientP); errSend != nil {
			logger.Errorf("[SendPolls] %s", errSend)
			delete(pollroom.members, req.GetDisplayName())
			logger.Infof("Participants count: %d", len(pollroom.members))
			pollroom.Unlock()
			return errSend
		}

		pollroom.Unlock()
	}
}

package poll

import (
	"context"
	"fmt"
	"net"

	pb "github.com/miska12345/DDPoll/ddpoll"
	goLogger "github.com/phachon/go-logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type poll struct {
	id           int64
	host         string
	members      []string
	title        string
	content      string
	accessbility int8     // Private - 1 (members is effective) | Public - 0 (members is effective)
	choices      []string // Description of each choices in the poill
	counts       []int64  // Vote counts, connected to choices by indices
	total int64
}



func (s *server) doCreatePoll(ctx context.Context, params []string) (as *pb.ActionSummary, id int64, err error) {
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

func createPoll(host string, members string[], title, content string, accessbility int8, choices []string) *server {
	p := new(Poll)

	// Initialize server struct
	p.host = host
	p.members = members
	p.title = title
	p.content = content
	p.accessbility = accessbility
	p.choices = choices
	p.counts = make([len(choices)] int)
	return s
}

func (p *poll) VoteUp(int choices) bool {
	if choices >= len(p.choices) {
		return false
	}
	p.choices[choices]++
	total++
}

func (p *poll) VoteDown(int choices) bool {
	if choices >= len(p.choices) {
		return false
	}
	p.choices[choices]--
	total--
}



// Package server provides an interface for DDPoll server
package server

import (
	"context"
	"errors"
	"fmt"
	"net"

	pb "github.com/miska12345/DDPoll/ddpoll"
	goLogger "github.com/phachon/go-logger"
	"google.golang.org/grpc"
)

var logger *goLogger.Logger

// server is a single instance of a server node - a serving entity
type server struct {
	pb.UnimplementedDDPollServer
	maxConnection int
}

func Run(port string) error {
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
	pb.RegisterDDPollServer(grpcServer, newServer())
	err = grpcServer.Serve(ls)
	return err
}

func newServer() *server {
	s := new(server)

	// Initialize server struct

	return s
}

func (s *server) Authenticate(ctx context.Context, query *pb.AuthQuery) (*pb.AuthResp, error) {
	if query.GetName() == "admin" && query.GetPassword() == "666" {
		/*
			md, ok := metadata.FromIncomingContext(ctx)
			if ok {
				md["sessionKey"] = make([]string, 1)
				md["sessionKey"][0] = "0xdeadbeef"
			}
		*/
		return &pb.AuthResp{
			Status:     1,
			SessionKey: "0xdeadbeef",
		}, nil
	}
	return nil, errors.New("Failed to verify")
}

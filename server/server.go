// Package server provides an interface for DDPoll server
package server

import (
	"context"
	"errors"
	"fmt"
	"net"

	pb "github.com/miska12345/DDPoll/ddpoll"
	"google.golang.org/grpc"
)

// server is a single instance of a server node - a serving entity
type server struct {
	pb.UnimplementedDDPollServer
	maxConnection int
}

func Run() error {
	fmt.Println("Server run called")
	ls, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		fmt.Println(err)
		return err
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	fmt.Println("here")
	pb.RegisterDDPollServer(grpcServer, newServer())
	err = grpcServer.Serve(ls)
	fmt.Println("Server closing")
	return err
}

func newServer() *server {
	fmt.Println("New server called")
	s := new(server)
	return s
}

func (s *server) Authenticate(ctx context.Context, query *pb.AuthQuery) (*pb.AuthResp, error) {
	if query.GetName() == "admin" && query.GetPassword() == "666" {
		return &pb.AuthResp{
			Status:     1,
			SessionKey: "0xdeadbeef",
		}, nil
	}
	return nil, errors.New("Failed to verify")
}

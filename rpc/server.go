package rpc

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"

	"github.com/mohamedshaaban/raft/config"
	"github.com/mohamedshaaban/raft/pb/raft"
	r "github.com/mohamedshaaban/raft/raft"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	config *config.Config
	node   *r.RaftNode
	raft.UnimplementedRaftServer
}

func NewGRPCServer(cfg *config.Config, node *r.RaftNode) *GRPCServer {
	return &GRPCServer{
		config: cfg,
		node:   node,
	}
}

func (s *GRPCServer) Start(port string, server *GRPCServer) {
	slog.Info("Starting grpc server on port " + port)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()

	raft.RegisterRaftServer(grpcServer, server)

	grpcServer.Serve(lis)
}

func (s *GRPCServer) Stop() {
}

func (s *GRPCServer) AppendEntries(ctx context.Context, pb *raft.AppendEntriesRequest) (*raft.AppendEntriesResponse, error) {
	// s.node.HandleAppendEntries()

	return &raft.AppendEntriesResponse{}, nil
}

func (s *GRPCServer) RequestVote(ctx context.Context, pb *raft.RequestVoteRequest) (*raft.RequestVoteResponse, error) {
	fmt.Println(s.node)
	voteGranted, term := s.node.HandleRequestVote(pb.GetCandidateID(), pb.GetTerm(), pb.GetLastLogIndex(), pb.GetLastLogTerm())

	return &raft.RequestVoteResponse{
		VoteGranted: voteGranted,
		Term:        term,
	}, nil
}

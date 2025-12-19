package rpc

import (
	"context"
	"log"
	"log/slog"
	"net"

	"github.com/mohamedshaaban/raft/config"
	"github.com/mohamedshaaban/raft/pb/raft"
	r "github.com/mohamedshaaban/raft/raft"
	"github.com/mohamedshaaban/raft/raft/election"
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

func (s *GRPCServer) AppendEntries(ctx context.Context, req *raft.AppendEntriesRequest) (*raft.AppendEntriesResponse, error) {
	return nil, nil
}

func (s *GRPCServer) RequestVote(ctx context.Context, pb *raft.RequestVoteRequest) (*raft.RequestVoteResponse, error) {
	req := &election.RequestVoteReq{
		CandidateID:   pb.CandidateID,
		CandidateTerm: pb.Term,
		LastLogTerm:   pb.LastLogTerm,
		LastLogIndex:  pb.LastLogIndex,
	}

	resp := s.node.HandleRequestVote(req)

	return &raft.RequestVoteResponse{
		VoteGranted: resp.VoteGranted,
		Term:        resp.Term,
	}, nil
}

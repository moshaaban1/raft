package client

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mohamedshaaban/raft/pb/raft"
	"github.com/mohamedshaaban/raft/raft/election"
	"github.com/mohamedshaaban/raft/raft/replication"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	conns map[string]raft.RaftClient
}

func NewClient(peers map[string]string) *GRPCClient {
	slog.Info("Initialize grpc client")

	clients := make(map[string]raft.RaftClient, len(peers))

	for ID, Address := range peers {
		conn, err := grpc.NewClient(Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			slog.Error(err.Error())
			return nil
		}

		clients[ID] = raft.NewRaftClient(conn)

		slog.Info(fmt.Sprintf("new connection to peerID: %s - address: %s has been established", ID, Address))
	}

	return &GRPCClient{
		conns: clients,
	}
}

func (c *GRPCClient) Close() {}

func (c *GRPCClient) SendRequestVote(ctx context.Context, peerID string, req *election.RequestVoteReq) (*raft.RequestVoteResponse, error) {
	conn := c.conns[peerID]

	if conn == nil {
		slog.Error(fmt.Sprintf("failed to find a connection to peerID: %s", peerID))
	}

	protoReq := &raft.RequestVoteRequest{
		CandidateID:  req.CandidateID,
		Term:         req.CandidateTerm,
		LastLogIndex: req.LastLogIndex,
		LastLogTerm:  req.CandidateTerm,
	}

	res, err := conn.RequestVote(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *GRPCClient) SendAppendEntries(ctx context.Context, peerID string, req *replication.AppendEntriesReq) (*raft.AppendEntriesResponse, error) {
	conn := c.conns[peerID]

	if conn == nil {
		slog.Error(fmt.Sprintf("failed to find a connection to peerID: %s", peerID))
	}

	protoReq := &raft.AppendEntriesRequest{
		Term:         req.Term,
		LeaderID:     req.LeaderID,
		PrevLogIndex: req.PrevLogIndex,
		PrevLogTerm:  req.PrevLogTerm,
		LeaderCommit: req.LeaderCommit,
	}

	res, err := conn.AppendEntries(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	return res, nil
}

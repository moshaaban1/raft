// Package raft
package raft

import (
	"github.com/mohamedshaaban/raft/config"
	"github.com/mohamedshaaban/raft/raft/election"
	"github.com/mohamedshaaban/raft/raft/state"
	"github.com/mohamedshaaban/raft/rpc/client"
)

type RaftNode struct {
	client *client.GRPCClient
	cfg    *config.Config
	state  *state.State
	le     *election.LeaderElection
}

func NewRaftNode(cfg *config.Config) *RaftNode {
	grpcClient := client.NewClient(cfg.GetOtherPeers())

	state := state.NewState()

	le := election.NewLeaderElection(grpcClient, cfg, state)

	return &RaftNode{
		client: grpcClient,
		le:     le,
		state:  state,
	}
}

func (n *RaftNode) Bootstrap() {
	n.le.StartElectionTimeout()
}

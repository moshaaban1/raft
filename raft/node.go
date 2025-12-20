// Package raft
package raft

import (
	"github.com/mohamedshaaban/raft/config"
	"github.com/mohamedshaaban/raft/raft/election"
	"github.com/mohamedshaaban/raft/raft/log"
	"github.com/mohamedshaaban/raft/raft/replication"
	"github.com/mohamedshaaban/raft/raft/state"
	"github.com/mohamedshaaban/raft/rpc/client"
)

type RaftNode struct {
	client *client.GRPCClient
	cfg    *config.Config
	state  *state.State
	le     *election.LeaderElection
	lr     *replication.LogReplication
	log    *log.Log
}

func NewRaftNode(cfg *config.Config) *RaftNode {
	grpcClient := client.NewClient(cfg.GetOtherPeers())

	state := state.NewState()
	log := log.NewLog()

	le := election.NewLeaderElection(grpcClient, cfg, state, log)
	lr := replication.NewLogReplication(grpcClient, cfg, state, log)

	le.SetHeartbeatManager(lr)
	lr.SetTimeoutResetter(le)

	return &RaftNode{
		client: grpcClient,
		le:     le,
		lr:     lr,
		state:  state,
		log:    log,
	}
}

func (n *RaftNode) Bootstrap() {
	n.le.StartElectionTimeout()
}

func (n *RaftNode) HandleRequestVote(req *election.RequestVoteReq) *election.RequestVoteResp {
	return n.le.HandleRequestVote(req)
}

func (n *RaftNode) HandleAppendEntries(req *replication.AppendEntriesReq) (*replication.AppendEntriesResp, error) {
	return n.lr.HandleAppendEntries(req)
}

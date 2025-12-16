// Package raft
package raft

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mohamedshaaban/raft/config"
	"github.com/mohamedshaaban/raft/rpc/client"
)

type RaftNode struct {
	client *client.GRPCClient
	cfg    *config.Config

	timer              *time.Timer
	heartbeatCtxCancel context.CancelFunc
	mu                 sync.Mutex

	state NodeState

	// Persistent state on all servers
	// It should be updated on a stable storage before responsing to RPCs
	currentTerm int32
	votedFor    string
	log         []LogEntry

	lastApplied uint
	commitIndex uint

	// Leader state
	nextIndex  any
	matchIndex any
	logger     *slog.Logger
}

func NewRaftNode(cfg *config.Config) *RaftNode {
	logger := slog.With("node_id", cfg.NodeID)

	logger.Info("Initialize a new Raft Node")

	grpcClient := client.NewClient(cfg.GetOtherPeers())

	return &RaftNode{
		client:      grpcClient,
		logger:      logger,
		cfg:         cfg,
		state:       FollowerState,
		currentTerm: 0,
		lastApplied: 0,
		log:         make([]LogEntry, 0),
		commitIndex: 0,
		nextIndex:   0,
		matchIndex:  0,
	}
}

func (n *RaftNode) Bootstrap() {
	n.startElectionTimeout()
}

func (n *RaftNode) startElectionTimeout() {
	if n.getCurrentState() == LeaderState {
		n.logger.Error("can't start a new election while currenct node is leader")
		return
	}

	n.timer = time.AfterFunc(n.randamDurationTimeout(), n.startNewElection)
}

func (n *RaftNode) resetElectionTimeout() {
	// TODO: Only if node is follower
	// Wrap timer with nil check
	if n.getCurrentState() == LeaderState {
		n.logger.Error("can't start a new election while currenct node is leader")
		return
	}

	n.timer.Reset(n.randamDurationTimeout())
}

func (n *RaftNode) startHeartbeatInterval() {
	if n.getCurrentState() != LeaderState {
		n.logger.Error("only leader can send heart beat requests to peers")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	n.mu.Lock()
	n.heartbeatCtxCancel = cancel
	n.mu.Unlock()

	startTicker(ctx, n.cfg.HeartbeatInterval, func() {})
}

func (n *RaftNode) getCurrentState() NodeState {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.state
}

func (n *RaftNode) becomeLeader() {
	n.mu.Lock()

	// Only Candidates may transition to Leader.
	if n.state != CandidateState {
		n.logger.Info("only candidate can become a leader")
		return
	}

	n.state = LeaderState

	n.mu.Unlock()

	n.logger.Info(fmt.Sprintf("candidate became leader for term %d", n.currentTerm))

	n.startHeartbeatInterval()
}

func (n *RaftNode) stepDown(term int32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if term > n.currentTerm {
		n.currentTerm = term
		n.state = FollowerState
		n.votedFor = ""

		if n.heartbeatCtxCancel != nil {
			n.heartbeatCtxCancel()
			n.heartbeatCtxCancel = nil
		}

		n.logger.Info(fmt.Sprintf("stepped down to follower, updated term to %d", term))
	}
}

func (n *RaftNode) isCandidateAtTerm(term int32) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.state == CandidateState && n.currentTerm == term
}

func (n *RaftNode) HandleRequestVote(candidateID string, candidateTerm int32, logIndex int32, logTerm int32) (bool, int32) {
	n.logger.Info(fmt.Sprintf("recieved a request vote from candidateID: %s for term %d", candidateID, candidateTerm))

	n.mu.Lock()
	defer n.mu.Unlock()

	if candidateTerm < n.currentTerm {
		return false, n.currentTerm
	}

	if n.votedFor != "" && n.votedFor != candidateID {
		return false, n.currentTerm
	}

	if candidateTerm > n.currentTerm {
		n.mu.Unlock()
		n.stepDown(candidateTerm)
		n.mu.Lock()
	}
	// TODO: is log updated

	n.votedFor = candidateID
	// n.resetElectionTimeout()

	return true, n.currentTerm
}

// AppendEntries appends log entries and also serves as a heartbeat to indicate that the leader is alive.
func (n *RaftNode) HandleAppendEntries() {
}

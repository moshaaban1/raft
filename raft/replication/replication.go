package replication

import (
	"context"
	"log/slog"
	"sync"

	"github.com/mohamedshaaban/raft/config"
	"github.com/mohamedshaaban/raft/raft/interfaces"
	"github.com/mohamedshaaban/raft/raft/log"
	"github.com/mohamedshaaban/raft/raft/state"
	"github.com/mohamedshaaban/raft/rpc/client"
)

type LogReplication struct {
	timeoutResetter interfaces.TimeoutResetter
	cfg             *config.Config
	client          *client.GRPCClient
	ctx             context.Context
	cancel          context.CancelFunc
	state           *state.State
	log             *log.Log
	mu              sync.Mutex
	logger          *slog.Logger

	// For each server, index of the next log entry
	// to send to that server (initialized to leader last log index + 1)
	nextIndex map[string]int32

	//	For each server, index of highest log entry
	// known to be replicated on server (initialized to 0, increases monotonically)
	matchIndex map[string]int32
}

func NewLogReplication(grpClient *client.GRPCClient, cfg *config.Config, state *state.State, log *log.Log) *LogReplication {
	logger := slog.With("node_id", cfg.NodeID)

	return &LogReplication{
		cfg:    cfg,
		client: grpClient,
		state:  state,
		log:    log,
		logger: logger,
	}
}

func (le *LogReplication) SetTimeoutResetter(i interfaces.TimeoutResetter) {
	le.timeoutResetter = i
}

func (lr *LogReplication) StartHeartbeat() {
	if !lr.state.IsLeader() {
		lr.logger.Debug("only leader can send heart beat requests to peers")
		return
	}

	lr.initializeLogState()

	// TODO: add context timeout duration
	ctx, cancel := context.WithCancel(context.Background())

	lr.mu.Lock()
	lr.ctx = ctx
	lr.cancel = cancel
	lr.mu.Unlock()

	startTicker(ctx, lr.cfg.HeartbeatInterval, func() {
		lr.appendEntriesToPeers(ctx)
	})
}

func (lr *LogReplication) StopHeartbeat() {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	if lr.cancel != nil {
		lr.cancel()
		lr.cancel = nil
	}
	lr.nextIndex = nil
	lr.matchIndex = nil
}

// Reinitialize nextIndex and matchIndex with each election
func (lr *LogReplication) initializeLogState() {
	peers := lr.cfg.GetOtherPeers()
	lastLogIndex := lr.log.GetLastLogIndex()

	if lr.nextIndex == nil || lr.matchIndex == nil {
		lr.nextIndex = make(map[string]int32)
		lr.matchIndex = make(map[string]int32)

		for peer := range peers {
			lr.nextIndex[peer] = lastLogIndex + 1
			lr.matchIndex[peer] = 0
		}
	}
}

func (lr *LogReplication) appendEntriesToPeers(ctx context.Context) {
	peers := lr.cfg.GetOtherPeers()
	currentTerm := lr.state.CurrentTerm()

	for ID := range peers {
		prevLogIndex := lr.nextIndex[ID] - 1
		entries := lr.log.GetEntries(lr.nextIndex[ID])

		logEntry, err := lr.log.GetEntry(int(prevLogIndex))
		if err != nil {
			lr.logger.Error(err.Error())
			return
		}

		go lr.sendAppendEntries(ctx, ID, AppendEntriesReq{
			Term:         currentTerm,
			LeaderID:     lr.cfg.NodeID,
			LeaderCommit: lr.state.CommitIndex(),
			PrevLogTerm:  logEntry.Term,
			Entries:      entries,
		},
		)
	}
}

func (lr *LogReplication) sendAppendEntries(ctx context.Context, peerID string, req AppendEntriesReq) {
	resp, err := lr.client.SendAppendEntries(ctx, peerID, &req)
	if err != nil {
		lr.logger.Error(err.Error())
		return
	}

	// Higher term was discoverd and node stepped down already
	if req.Term != lr.state.CurrentTerm() {
		return
	}

	ok := lr.state.StepDown(resp.Term) // only step down if term is greater than currentTerm

	if ok {
		// stop heartbeat and start electiontimeout
	}
}

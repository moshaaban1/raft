package election

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mohamedshaaban/raft/config"
	"github.com/mohamedshaaban/raft/raft/state"
	"github.com/mohamedshaaban/raft/rpc/client"
)

type electionContext struct {
	term         int32
	candidateID  string
	lastLogIndex int32
	lastLogTerm  int32
}

type LeaderElection struct {
	cfg    *config.Config
	client *client.GRPCClient
	state  *state.State
	mu     sync.Mutex
	logger *slog.Logger
	timer  *time.Timer
}

func NewLeaderElection(grpClient *client.GRPCClient, cfg *config.Config, state *state.State) *LeaderElection {
	logger := slog.With("node_id", cfg.NodeID)

	return &LeaderElection{
		cfg:    cfg,
		client: grpClient,
		state:  state,
		logger: logger,
	}
}

func (le *LeaderElection) StartElectionTimeout() {
	le.mu.Lock()
	defer le.mu.Unlock()

	if le.state.IsLeader() {
		le.logger.Error("can't start a new election while current node is leader")
		return
	}

	if le.timer != nil {
		le.timer.Stop()
	}

	le.timer = time.AfterFunc(le.randamDurationTimeout(), le.startNewElection)
}

func (le *LeaderElection) ResetElectionTimeout() {
	le.mu.Lock()
	timer := le.timer
	le.mu.Unlock()

	if le.state.IsLeader() {
		le.logger.Debug("can't start a new election while current node is leader")
		return
	}

	if timer == nil {
		le.StartElectionTimeout()
		return
	}

	le.timer.Reset(le.randamDurationTimeout())
}

func (le *LeaderElection) startNewElection() {
	electionTimeout := le.randamDurationTimeout()

	ctx, cancel := context.WithTimeout(context.Background(), electionTimeout)
	defer cancel()

	ec := le.prepareElection()
	le.logger.Info("start a new election", "term", le.state.CurrentTerm())

	peers := le.cfg.GetOtherPeers()
	voteChan := make(chan bool, len(peers))

	le.sendElectionRequestVoteToPeers(ctx, ec, peers, voteChan)

	votes := le.collectElectionVotes(ctx, ec, voteChan)

	le.evaluateElectionResult(ec, votes)
}

func (le *LeaderElection) prepareElection() *electionContext {
	term := le.state.PrepareElection(le.cfg.NodeID)

	return &electionContext{
		term:         term,
		lastLogIndex: 0,
		lastLogTerm:  0,
		candidateID:  le.cfg.NodeID,
	}
}

func (le *LeaderElection) requestVoteFromPeer(ctx context.Context, peerID string, ec *electionContext) (bool, error) {
	resp, err := le.client.SendRequestVote(ctx, peerID, ec.candidateID, ec.term, ec.lastLogIndex, ec.lastLogTerm)
	if err != nil {
		return false, err
	}

	// Higher term discovered
	if resp.Term > ec.term {
		if le.state.StepDown(resp.Term) {
			le.ResetElectionTimeout() // stepped down, reset
		}
		// If false, someone else already stepped down and reset
		return false, nil
	}

	if resp.VoteGranted {
		le.logger.Info(fmt.Sprintf("vote granted from peer: %s", peerID))
	}

	return resp.VoteGranted, nil
}

func (le *LeaderElection) sendElectionRequestVoteToPeers(electionCtx context.Context, ec *electionContext, peers map[string]string, voteChan chan<- bool) {
	for ID := range peers {
		go func(peerId string) {
			rpcCtx, cancel := context.WithTimeout(electionCtx, 150*time.Millisecond)

			defer cancel()

			granted, err := le.requestVoteFromPeer(rpcCtx, peerId, ec)
			if err != nil {
				le.logger.Error(err.Error(), "peer_id", peerId)
				voteChan <- false
				return
			}

			voteChan <- granted
		}(ID)
	}
}

func (le *LeaderElection) collectElectionVotes(electionCtx context.Context, ec *electionContext, voteChan <-chan bool) int {
	votes := 1 // Each candidate vote for itself
	peersLen := len(le.cfg.Peers) - 1

	for i := 0; i < peersLen; i++ {
		// Early termination as node is no longer a candidate - since state change or new term is discovered
		if !le.state.IsCandidateAtTerm(ec.term) {
			return votes
		}

		select {
		case voteGranted := <-voteChan:
			if voteGranted {
				votes++
				if le.isRecivedMajorityVotes(votes) {
					return votes // early exit on winning
				}
			}
		case <-electionCtx.Done():
			return votes
		}
	}

	return votes
}

func (le *LeaderElection) isRecivedMajorityVotes(votes int) bool {
	majority := (len(le.cfg.Peers) / 2) + 1

	return votes >= majority
}

func (le *LeaderElection) evaluateElectionResult(ec *electionContext, votes int) {
	if le.isRecivedMajorityVotes(votes) && le.state.IsCandidateAtTerm(ec.term) {

		le.logger.Info("candidate won the election", "term", ec.term)

		le.state.BecomeLeader()

		return
	}

	le.logger.Info("election either end with no winner or candidate los the election", "term", ec.term)

	// Start a new election with a randamized timer to prevent split votes
	le.ResetElectionTimeout()
}

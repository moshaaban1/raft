package raft

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type electionContext struct {
	term         int32
	candidateID  string
	lastLogIndex int32
	lastLogTerm  int32
}

func (n *RaftNode) startNewElection() {
	electionTimeout := n.randamDurationTimeout()

	ctx, cancel := context.WithTimeout(context.Background(), electionTimeout)
	defer cancel()

	ec := n.prepareElection()

	slog.Info("Start a new election", "Term", n.currentTerm)

	peers := n.cfg.GetOtherPeers()
	voteChan := make(chan bool, len(peers))

	n.sendElectionRequestVoteToPeers(ctx, ec, peers, voteChan)

	votes := n.collectElectionVotes(ctx, ec, voteChan)

	n.evaluateElectionResult(ec, votes)
}

func (n *RaftNode) prepareElection() *electionContext {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.currentTerm++
	n.state = CandidateState
	n.votedFor = n.cfg.NodeID

	return &electionContext{
		term:         n.currentTerm,
		lastLogIndex: 0,
		lastLogTerm:  0,
		candidateID:  n.cfg.NodeID,
	}
}

func (n *RaftNode) requestVoteFromPeer(ctx context.Context, peerID string, ec *electionContext) (bool, error) {
	resp, err := n.client.SendRequestVote(ctx, peerID, ec.candidateID, ec.term, ec.lastLogIndex, ec.lastLogTerm)
	if err != nil {
		return false, err
	}

	// Higher term discovered
	if resp.Term > ec.term {
		n.stepDown(resp.Term)
		return false, nil
	}

	if resp.VoteGranted {
		slog.Info(fmt.Sprintf("Vote granted from peer: %s", peerID))
	}

	return resp.VoteGranted, nil
}

func (n *RaftNode) sendElectionRequestVoteToPeers(electionCtx context.Context, ec *electionContext, peers map[string]string, voteChan chan<- bool) {
	for ID := range peers {
		go func(peerId string) {
			rpcCtx, cancel := context.WithTimeout(electionCtx, 150*time.Millisecond)

			defer cancel()

			granted, err := n.requestVoteFromPeer(rpcCtx, peerId, ec)
			if err != nil {
				slog.Error(err.Error())
				voteChan <- false
				return
			}

			voteChan <- granted
		}(ID)
	}
}

func (n *RaftNode) collectElectionVotes(electionCtx context.Context, ec *electionContext, voteChan <-chan bool) int {
	votes := 1 // Each candidate vote for itself
	peersLen := len(n.cfg.Peers)

	for i := 0; i < peersLen; i++ {
		// Early termination as node is no longer a candidate - since state change or new term is discovered
		if !n.isCandidateAtTerm(ec.term) {
			return votes
		}

		select {
		case voteGranted := <-voteChan:
			if voteGranted {
				votes++
				if n.isRecivedMajorityVotes(votes) {
					return votes // early exit on winning
				}
			}
		case <-electionCtx.Done():
			return votes
		}
	}

	return votes
}

func (n *RaftNode) isRecivedMajorityVotes(votes int) bool {
	majority := (len(n.cfg.Peers) / 2) + 1

	return votes >= majority
}

// EvaluateElectionResult evaluates the election result based on the collected votes.
// The election can end in one of three states:
//  1. The node wins the election by reaching a majority and should become leader.
//  2. The node loses the election because another term or leader is discovered.
//  3. The election results in a split vote, requiring a new election round.
func (n *RaftNode) evaluateElectionResult(ec *electionContext, votes int) {
	if n.isRecivedMajorityVotes(votes) && n.isCandidateAtTerm(ec.term) {

		slog.Info(fmt.Sprintf("Candidate won the election - term: %d", ec.term))

		n.becomeLeader()

		return
	}

	slog.Info(fmt.Sprintf("Election either end with no winner or candidate los the election for term: %d", ec.term))

	// Start a new election with a randamized timer to prevent split votes
	n.startElectionTimeout()
}

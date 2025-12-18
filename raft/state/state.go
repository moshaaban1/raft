package state

import (
	"sync"
)

type State struct {
	mu sync.Mutex

	state NodeState

	// Persistent state on all servers
	// It should be updated on a stable storage before responsing to RPCs
	term     int32
	votedFor string
	log      []LogEntry

	lastApplied uint
	commitIndex uint

	// Leader state
	nextIndex  any
	matchIndex any
}

func NewState() *State {
	return &State{
		state:       FollowerState,
		term:        0,
		lastApplied: 0,
		log:         make([]LogEntry, 0),
		commitIndex: 0,
		nextIndex:   0,
		matchIndex:  0,
	}
}

func (n *State) CurrentState() NodeState {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.state
}

func (n *State) CurrentTerm() int32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.term
}

func (n *State) BecomeLeader() {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Only Candidates may transition to Leader.
	// if n.state != CandidateState {
	// 	n.logger.Info("only candidate can become a leader")
	//
	// 	return
	// }

	n.state = LeaderState

	// n.logger.Info(fmt.Sprintf("candidate became leader for term %d", n.currentTerm))

	// n.startHeartbeatInterval()
}

func (n *State) StepDown(term int32) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if term > n.term {
		n.term = term
		n.state = FollowerState
		n.votedFor = ""

		return true
	}

	return false
}

func (n *State) IsCandidateAtTerm(term int32) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.state == CandidateState && n.term == term
}

func (s *State) IsLeader() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.state == LeaderState
}

func (s *State) PrepareElection(candidateID string) int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.term++
	s.state = CandidateState
	s.votedFor = candidateID
	return s.term
}

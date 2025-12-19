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

func (s *State) CurrentState() NodeState {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.state
}

func (s *State) CurrentTerm() int32 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.term
}

func (s *State) BecomeLeader(candidateTerm int32) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Only Candidates may transition to Leader.
	if s.state != CandidateState || s.term != candidateTerm {
		return false
	}

	s.state = LeaderState

	return true
}

func (s *State) StepDown(term int32) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if term > s.term {
		s.term = term
		s.state = FollowerState
		s.votedFor = ""

		return true
	}

	return false
}

func (s *State) IsCandidateAtTerm(term int32) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.state == CandidateState && s.term == term
}

func (s *State) IsLeader() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.state == LeaderState
}

func (s *State) IsFollower() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.state == FollowerState
}

func (s *State) PrepareElection(candidateID string) int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.term++
	s.state = CandidateState
	s.votedFor = candidateID
	return s.term
}

func (s *State) TryGrantVote(candidateID string, candidateTerm int32, logOk bool) (bool, int32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Reject if candidate's term is outdated
	if candidateTerm < s.term {
		return false, s.term
	}

	// Step down if candidate has higher term
	if candidateTerm > s.term {
		s.term = candidateTerm
		s.state = FollowerState
		s.votedFor = ""
	}

	// Reject if already voted for someone else this term
	if s.votedFor != "" && s.votedFor != candidateID {
		return false, s.term
	}

	// Reject if candidate's log is not up-to-date
	if !logOk {
		return false, s.term
	}

	s.votedFor = candidateID
	return true, s.term
}

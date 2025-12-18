package state

type NodeState int

const (
	FollowerState NodeState = iota
	CandidateState
	LeaderState
)

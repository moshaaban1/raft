package raft

type NodeState int

const (
	FollowerState NodeState = iota
	CandidateState
	LeaderState
)

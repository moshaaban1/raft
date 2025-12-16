package raft

type LogEntry struct {
	Term    uint
	Index   uint
	Command any
}

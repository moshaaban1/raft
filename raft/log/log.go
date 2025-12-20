package log

import (
	"errors"
	"sync"
)

type Log struct {
	entries []LogEntry
	mu      sync.RWMutex
}

func NewLog() *Log {
	return &Log{
		entries: make([]LogEntry, 0),
	}
}

func (l *Log) GetEntry(index int) (LogEntry, error) {
	if index < 1 || index > len(l.entries) {
		return LogEntry{}, errors.New("index out of range")
	}
	return l.entries[index-1], nil // Convert to 0-based
}

func (l *Log) Append(entries []LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
}

func (l *Log) GetLastLogIndex() int32 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.entries) == 0 {
		return 0
	}

	return l.entries[len(l.entries)-1].Index
}

func (l *Log) GetLatestLogInfo() LogInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.entries) == 0 {
		return LogInfo{Index: 0, Term: 0}
	}

	last := l.entries[len(l.entries)-1]
	return LogInfo{Index: int32(len(l.entries)), Term: last.Term}
}

func (l *Log) GetEntries(start int32) []*LogEntry {
}

func (l *Log) TruncateFrom(index int32) {
}

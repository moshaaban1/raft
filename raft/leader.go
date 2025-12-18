package raft

import (
	"context"

	"github.com/mohamedshaaban/raft/pb/raft"
)

func (n *RaftNode) appendEntriesToPeers(ctx context.Context) {
	peers := n.cfg.GetOtherPeers()

	n.mu.Lock()
	req := &raft.AppendEntriesRequest{
		Term:         n.currentTerm,
		LeaderID:     n.cfg.NodeID,
		PrevLogTerm:  0,
		PrevLogIndex: 0,
		LeaderCommit: 0,
	}
	n.mu.Unlock()

	for ID := range peers {
		go n.appendEntries(ctx, ID, req)
	}
}

func (n *RaftNode) appendEntries(ctx context.Context, peerID string, req *raft.AppendEntriesRequest) {
	resp, err := n.client.SendAppendEntries(ctx, peerID, req)
	if err != nil {
		n.logger.Error(err.Error())
		return
	}

	// Higher term was discoverd and node stepped down already
	if req.Term != n.getCurrentTerm() {
		return
	}

	n.stepDown(resp.Term) // only step down if term is greater than currentTerm
}

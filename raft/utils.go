package raft

import (
	"context"
	"math/rand"
	"time"
)

func randamNumber(min int, max int) int {
	return rand.Intn(max-min+1) + min
}

func (n *RaftNode) randamDurationTimeout() time.Duration {
	randamDurationTimeout := randamNumber(n.cfg.ElectionTimeoutMin, n.cfg.ElectionTimeoutMax)

	return time.Duration(randamDurationTimeout) * time.Millisecond
}

func startTicker(ctx context.Context, interval time.Duration, f func()) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// do work

			case <-ctx.Done():

				return
			}
		}
	}()
}

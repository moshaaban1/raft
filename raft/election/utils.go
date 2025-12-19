package election

import (
	"math/rand"
	"time"
)

func randamNumber(min int, max int) int {
	return rand.Intn(max-min+1) + min
}

func (le *LeaderElection) randamDurationTimeout() time.Duration {
	randamDurationTimeout := randamNumber(le.cfg.ElectionTimeoutMin, le.cfg.ElectionTimeoutMax)

	return time.Duration(randamDurationTimeout) * time.Millisecond
}

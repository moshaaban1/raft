package election

// func (le *LeaderElection) HandleRequestVote(candidateID string, candidateTerm int32, logIndex int32, logTerm int32) (bool, int32) {
// 	n.logger.Info(fmt.Sprintf("recieved a request vote from candidateID: %s for term %d", candidateID, candidateTerm))
//
// 	n.mu.Lock()
// 	defer n.mu.Unlock()
//
// 	if candidateTerm < n.currentTerm || n.votedFor != "" && n.votedFor != candidateID {
// 		return false, n.currentTerm
// 	}
//
// 	if candidateTerm > n.currentTerm {
// 		n.mu.Unlock()
// 		n.stepDown(candidateTerm)
// 		n.mu.Lock()
// 	}
//
// 	// TODO: is log updated
// 	// if yes, grant vote to candidate
//
// 	n.votedFor = candidateID
// 	n.resetElectionTimeout()
//
// 	return true, n.currentTerm
// }

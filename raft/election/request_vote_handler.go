package election

import "fmt"

type RequestVoteReq struct {
	CandidateID   string
	CandidateTerm int32
	LastLogIndex  int32
	LastLogTerm   int32
}

type RequestVoteResp struct {
	VoteGranted bool
	Term        int32
}

func (le *LeaderElection) HandleRequestVote(req *RequestVoteReq) *RequestVoteResp {
	le.logger.Info(fmt.Sprintf("recieved a request vote from candidateID: %s for term %d", req.CandidateID, req.CandidateTerm))

	// TODO: is log updated
	// if yes, grant vote to candidate
	logOk := true

	// bypass the decision to state to handle it atomically
	granted, term := le.state.TryGrantVote(req.CandidateID, req.CandidateTerm, logOk)

	if !granted {
		return rejectVote(term)
	}

	le.ResetElectionTimeout()

	return grantVote(term)
}

func grantVote(term int32) *RequestVoteResp {
	return &RequestVoteResp{VoteGranted: true, Term: term}
}

func rejectVote(term int32) *RequestVoteResp {
	return &RequestVoteResp{VoteGranted: false, Term: term}
}

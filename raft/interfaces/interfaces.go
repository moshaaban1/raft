package interfaces

type HeartbeatManager interface {
	StartHeartbeat()
	StopHeartbeat()
}

type TimeoutResetter interface {
	ResetElectionTimeout()
}

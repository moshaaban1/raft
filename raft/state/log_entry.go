package state

type LogEntry struct {
	Term    int32
	Index   int32
	Command any
}

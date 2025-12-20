package log

type LogInfo struct {
	Index int32
	Term  int32
}

type LogEntry struct {
	LogInfo
	Command []byte
}

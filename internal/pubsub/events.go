package broker

type EventType int

const (
	EventUserMessage EventType = iota
	EventCustomTool
	EventAgentThinking
	EventAgentResponse
	EventAgentError
	EventSystemNotice
	EventSystemNoticeError
	EventStreamStarted
	EventToolFileReading
	EventToolFileWriting
	EventToolBash
	EventStreamDone
)

package broker

type EventType int

const (
	EventUserMessage EventType = iota
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

package broker

type EventType = string

const (
	EventUserMessage        = "new_message"
	EventAgentStartThinking = "agent_start_thinking"
	EventAgentStopThinking  = "agent_stop_thinking"
	EventAgentThinking      = "agent_thinking"
	EventAgentMessageChunk  = "agent_message"
	EventAgentError         = "agent_error"
)

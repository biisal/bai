package web

import (
	"sync"
	"time"

	"charm.land/fantasy"
)

func decorate(run *Run) func(*fantasy.AgentStreamCall) {
	started := time.Now()
	var mu sync.Mutex
	toolStarted := make(map[string]bool)

	return func(call *fantasy.AgentStreamCall) {
		baseTextStart := call.OnTextStart
		call.OnTextStart = func(id string) error {
			run.Emit(fTextStart(id))
			if baseTextStart != nil {
				return baseTextStart(id)
			}
			return nil
		}

		baseTextDelta := call.OnTextDelta
		call.OnTextDelta = func(id, text string) error {
			run.Emit(fTextDelta(id, text))
			if baseTextDelta != nil {
				return baseTextDelta(id, text)
			}
			return nil
		}

		baseTextEnd := call.OnTextEnd
		call.OnTextEnd = func(id string) error {
			run.Emit(fTextEnd(id))
			if baseTextEnd != nil {
				return baseTextEnd(id)
			}
			return nil
		}

		baseReasoningStart := call.OnReasoningStart
		call.OnReasoningStart = func(id string, reasoning fantasy.ReasoningContent) error {
			run.Emit(fReasoningStart(id))
			if baseReasoningStart != nil {
				return baseReasoningStart(id, reasoning)
			}
			return nil
		}

		baseReasoningDelta := call.OnReasoningDelta
		call.OnReasoningDelta = func(id, text string) error {
			run.Emit(fReasoningDelta(id, text))
			if baseReasoningDelta != nil {
				return baseReasoningDelta(id, text)
			}
			return nil
		}

		baseReasoningEnd := call.OnReasoningEnd
		call.OnReasoningEnd = func(id string, reasoning fantasy.ReasoningContent) error {
			run.Emit(fReasoningEnd(id))
			if baseReasoningEnd != nil {
				return baseReasoningEnd(id, reasoning)
			}
			return nil
		}

		baseToolInputStart := call.OnToolInputStart
		call.OnToolInputStart = func(id, toolName string) error {
			mu.Lock()
			toolStarted[id] = true
			mu.Unlock()
			run.Emit(fToolInputStart(id, toolName))
			if baseToolInputStart != nil {
				return baseToolInputStart(id, toolName)
			}
			return nil
		}

		baseToolInputDelta := call.OnToolInputDelta
		call.OnToolInputDelta = func(id, delta string) error {
			run.Emit(fToolInputDelta(id, delta))
			if baseToolInputDelta != nil {
				return baseToolInputDelta(id, delta)
			}
			return nil
		}

		baseToolInputEnd := call.OnToolInputEnd
		call.OnToolInputEnd = func(id string) error {
			if baseToolInputEnd != nil {
				return baseToolInputEnd(id)
			}
			return nil
		}

		baseToolCall := call.OnToolCall
		call.OnToolCall = func(toolCall fantasy.ToolCallContent) error {
			mu.Lock()
			started := toolStarted[toolCall.ToolCallID]
			toolStarted[toolCall.ToolCallID] = true
			mu.Unlock()
			if !started {
				run.Emit(fToolInputStart(toolCall.ToolCallID, toolCall.ToolName))
			}
			run.Emit(fToolInputAvailable(toolCall.ToolCallID, toolCall.ToolName, toolCall.Input))
			if baseToolCall != nil {
				return baseToolCall(toolCall)
			}
			return nil
		}

		baseToolResult := call.OnToolResult
		call.OnToolResult = func(result fantasy.ToolResultContent) error {
			run.Emit(toolResultFrame(result))
			if baseToolResult != nil {
				return baseToolResult(result)
			}
			return nil
		}

		baseStepStart := call.OnStepStart
		call.OnStepStart = func(stepNumber int) error {
			run.Emit(fStartStep())
			if baseStepStart != nil {
				return baseStepStart(stepNumber)
			}
			return nil
		}

		baseStepFinish := call.OnStepFinish
		call.OnStepFinish = func(step fantasy.StepResult) error {
			err := error(nil)
			if baseStepFinish != nil {
				err = baseStepFinish(step)
			}
			run.Emit(fFinishStep())
			return err
		}

		baseFinish := call.OnFinish
		call.OnFinish = func(result *fantasy.AgentResult) {
			run.Emit(fUsage(result.TotalUsage, time.Since(started)))
			if baseFinish != nil {
				baseFinish(result)
			}
		}

		baseRetry := call.OnRetry
		call.OnRetry = func(err *fantasy.ProviderError, delay time.Duration) {
			mu.Lock()
			clear(toolStarted)
			mu.Unlock()
			run.Emit(fResetStep())
			if baseRetry != nil {
				baseRetry(err, delay)
			}
		}
	}
}

package tools

import (
	"context"
	"fmt"

	fantasy "charm.land/fantasy"
	broker "github.com/biisal/bai/internal/pubsub"
)

// Tool name constants – referenced by instruction and chat-builder packages.
const (
	ReadFileName  = "read_file"
	WriteFileName = "write_file"
	EditFileName  = "edit_file"
	BashName      = "bash"
)

// input types – fantasy generates JSON schemas from these automatically.

type readFileInput struct {
	Path   string `json:"path"`
	Offset *int64 `json:"offset,omitempty"`
	Limit  *int64 `json:"limit,omitempty"`
}

type writeFileInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type editEntry struct {
	OldText string `json:"old_text"`
	NewText string `json:"new_text"`
}

type editFileInput struct {
	Path  string      `json:"path"`
	Edits []editEntry `json:"edits"`
}

type bashInput struct {
	Command string `json:"command"`
	Timeout *int   `json:"timeout,omitempty"`
}

// toolSet holds the broker so tool methods can publish events without closures.
type toolSet struct {
	broker broker.Service
}

func (t *toolSet) readFile(ctx context.Context, input readFileInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
	var offset, limit int64
	if input.Offset != nil {
		offset = *input.Offset
	}
	if input.Limit != nil {
		limit = *input.Limit
	}
	t.broker.Publish(ctx, broker.Message{
		Type:       broker.EventToolFileReading,
		Text:       fmt.Sprintf("read:%s:%d:%d", input.Path, offset, limit),
		IsComplete: true,
	})
	content, err := ReadFile(input.Path, offset, limit)
	if err != nil {
		return fantasy.NewTextErrorResponse(err.Error()), nil
	}
	return fantasy.NewTextResponse(content), nil
}

func (t *toolSet) writeFile(ctx context.Context, input writeFileInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
	t.broker.Publish(ctx, broker.Message{
		Type:       broker.EventToolFileWriting,
		Text:       fmt.Sprintf("write:%s", input.Path),
		IsComplete: true,
	})
	if err := WriteFile(ctx, input.Path, input.Content); err != nil {
		return fantasy.NewTextErrorResponse(err.Error()), nil
	}
	return fantasy.NewTextResponse(fmt.Sprintf("Successfully wrote %d bytes to %s", len(input.Content), input.Path)), nil
}

func (t *toolSet) editFile(ctx context.Context, input editFileInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
	t.broker.Publish(ctx, broker.Message{
		Type:       broker.EventToolFileWriting,
		Text:       fmt.Sprintf("edit:%s", input.Path),
		IsComplete: true,
	})
	edits := make([]Edit, len(input.Edits))
	for i, e := range input.Edits {
		edits[i] = Edit{OldText: e.OldText, NewText: e.NewText}
	}
	if err := EditFile(ctx, input.Path, edits); err != nil {
		return fantasy.NewTextErrorResponse(err.Error()), nil
	}
	return fantasy.NewTextResponse(fmt.Sprintf("Successfully replaced %d block(s) in %s", len(edits), input.Path)), nil
}

func (t *toolSet) bash(ctx context.Context, input bashInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
	t.broker.Publish(ctx, broker.Message{
		Type:       broker.EventToolBash,
		Text:       input.Command,
		IsComplete: true,
	})
	out, err := executeBash(ctx, input.Command, input.Timeout)
	if err != nil {
		return fantasy.NewTextErrorResponse(err.Error()), nil
	}
	return fantasy.NewTextResponse(out), nil
}

// NewTools returns all agent tools wired up with broker publishing.
func NewTools(b broker.Service) []fantasy.AgentTool {
	ts := &toolSet{broker: b}
	return []fantasy.AgentTool{
		fantasy.NewAgentTool(ReadFileName, "Read a file from the filesystem using path, offset, and limit", ts.readFile),
		fantasy.NewAgentTool(WriteFileName, "Write content to a file. Creates the file if it doesn't exist, overwrites if it does. Automatically creates parent directories.", ts.writeFile),
		fantasy.NewAgentTool(EditFileName, "Edit a file using exact text replacement. Each edit's old_text must be a unique, non-overlapping match in the original file and is replaced with new_text. Provide multiple edits in one call to change several locations at once.", ts.editFile),
		fantasy.NewAgentTool(BashName, "Execute a bash command in the current working directory. Returns combined stdout and stderr, and the exit code when non-zero. Output is truncated to the last 2000 lines or 256KB. Optionally provide a timeout in seconds, after which the command is killed.", ts.bash),
	}
}

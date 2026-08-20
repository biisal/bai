package tools

import (
	"context"
	"encoding/json"
	"fmt"

	broker "github.com/biisal/bai/internal/pubsub"
)

type ToolType string

const (
	ReadFileTool  ToolType = "read_file"
	WriteFileTool ToolType = "write_file"
	EditFileTool  ToolType = "edit_file"
	BashTool      ToolType = "bash"
)

type Call struct {
	ID   string
	Name ToolType
	Args json.RawMessage
}

type Definition struct {
	Type        ToolType
	Description string
	Parameters  map[string]any
}

var Definitions = []Definition{
	{
		Type:        ReadFileTool,
		Description: "Read a file from the filesystem using path, offset, and limit",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":   map[string]string{"type": "string"},
				"offset": map[string]string{"type": "integer"},
				"limit":  map[string]string{"type": "integer"},
			},
			"required": []string{"path"},
		},
	},
	{
		Type:        WriteFileTool,
		Description: "Write content to a file. Creates the file if it doesn't exist, overwrites if it does. Automatically creates parent directories.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":    map[string]string{"type": "string"},
				"content": map[string]string{"type": "string"},
			},
			"required": []string{"path", "content"},
		},
	},
	{
		Type:        EditFileTool,
		Description: "Edit a file using exact text replacement. Each edit's old_text must be a unique, non-overlapping match in the original file (not matched incrementally against prior edits) and is replaced with new_text. Provide multiple edits in one call to change several locations at once instead of calling this repeatedly.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]string{"type": "string"},
				"edits": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"old_text": map[string]any{
								"type":        "string",
								"description": "Exact text to replace. Must be unique in the file and must not overlap with any other edit's old_text.",
							},
							"new_text": map[string]string{"type": "string"},
						},
						"required": []string{"old_text", "new_text"},
					},
				},
			},
			"required": []string{"path", "edits"},
		},
	},
	{
		Type:        BashTool,
		Description: "Execute a bash command in the current working directory. Returns combined stdout and stderr, and the exit code when non-zero. Output is truncated to the last 2000 lines or 256KB. Optionally provide a timeout in seconds, after which the command is killed.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]string{"type": "string"},
				"timeout": map[string]any{
					"type":        "integer",
					"description": "Timeout in seconds (optional)",
				},
			},
			"required": []string{"command"},
		},
	},
}

func Execute(ctx context.Context, call Call, b broker.Service) (content string, isError bool) {
	switch call.Name {
	case ReadFileTool:
		var args struct {
			Path   string `json:"path"`
			Offset *int64 `json:"offset"`
			Limit  *int64 `json:"limit"`
		}
		if err := json.Unmarshal(call.Args, &args); err != nil {
			return fmt.Sprintf("error parsing arguments for %s: %v", call.Name, err), true
		}
		if args.Path == "" {
			return fmt.Sprintf("error: missing required argument \"path\" for %s", call.Name), true
		}
		var offset, limit int64
		if args.Offset != nil {
			offset = *args.Offset
		}
		if args.Limit != nil {
			limit = *args.Limit
		}
		b.Publish(ctx, broker.Message{Type: broker.EventToolFileReading, Text: fmt.Sprintf("read:%s:%d:%d", args.Path, offset, limit)})
		content, err := ReadFile(args.Path, offset, limit)
		if err != nil {
			return fmt.Sprintf("error reading file: %v", err), true
		}
		return content, false
	case WriteFileTool:
		var args struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(call.Args, &args); err != nil {
			return fmt.Sprintf("error parsing arguments for %s: %v", call.Name, err), true
		}
		if args.Path == "" {
			return fmt.Sprintf("error: missing required argument \"path\" for %s", call.Name), true
		}
		b.Publish(ctx, broker.Message{Type: broker.EventToolFileWriting, Text: fmt.Sprintf("write:%s", args.Path)})
		if err := WriteFile(ctx, args.Path, args.Content); err != nil {
			return fmt.Sprintf("error writing file: %v", err), true
		}
		return fmt.Sprintf("Successfully wrote %d bytes to %s", len(args.Content), args.Path), false
	case EditFileTool:
		var args struct {
			Path  string `json:"path"`
			Edits []struct {
				OldText string `json:"old_text"`
				NewText string `json:"new_text"`
			} `json:"edits"`
		}
		if err := json.Unmarshal(call.Args, &args); err != nil {
			return fmt.Sprintf("error parsing arguments for %s: %v", call.Name, err), true
		}
		if args.Path == "" {
			return fmt.Sprintf("error: missing required argument \"path\" for %s", call.Name), true
		}
		if len(args.Edits) == 0 {
			return fmt.Sprintf("error: missing required argument \"edits\" for %s", call.Name), true
		}
		edits := make([]Edit, len(args.Edits))
		for i, e := range args.Edits {
			edits[i] = Edit{OldText: e.OldText, NewText: e.NewText}
		}
		b.Publish(ctx, broker.Message{Type: broker.EventToolFileWriting, Text: fmt.Sprintf("edit:%s", args.Path)})
		if err := EditFile(ctx, args.Path, edits); err != nil {
			return fmt.Sprintf("error editing file: %v", err), true
		}
		return fmt.Sprintf("Successfully replaced %d block(s) in %s", len(edits), args.Path), false
	case BashTool:
		var args struct {
			Command string `json:"command"`
			Timeout *int   `json:"timeout"`
		}
		if err := json.Unmarshal(call.Args, &args); err != nil {
			return fmt.Sprintf("error parsing arguments for %s: %v", call.Name, err), true
		}
		if args.Command == "" {
			return fmt.Sprintf("error: missing required argument \"command\" for %s", call.Name), true
		}
		b.Publish(ctx, broker.Message{Type: broker.EventToolBash, Text: fmt.Sprintf("$%s", args.Command)})
		return executeBash(ctx, args.Command, args.Timeout)
	default:
		return fmt.Sprintf("unknown tool: %s", call.Name), true
	}
}

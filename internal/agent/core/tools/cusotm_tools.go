package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/fantasy"
	broker "github.com/biisal/bai/internal/pubsub"
)

type inputSchema map[string]string

type ReadCustomTool struct {
	FileName    string      `json:"file_name"`
	Description string      `json:"description"`
	InputSchema inputSchema `json:"input_schema"`
}

type CustomToolDef struct {
	FileName       string      `json:"file_name"`
	ExecutablePath string      `json:"executable_path"`
	Description    string      `json:"description"`
	InputSchema    inputSchema `json:"input_schema"`
}

func makeNameFromPath(fileName string) string {
	name := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	name = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
			return r
		}
		return '_'
	}, name)
	return name
}

func parseTooolsByPath(jsonPath string) ([]CustomToolDef, error) {
	file, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	var tools []ReadCustomTool
	if err = json.Unmarshal(file, &tools); err != nil {
		return nil, err
	}
	describe := func(desc string, args map[string]string) string {
		desc += "args: "
		for k, v := range args {
			desc += k + ": " + v + " "
		}
		return desc
	}
	var customTools []CustomToolDef
	for _, tool := range tools {
		customTools = append(customTools, CustomToolDef{
			FileName:       tool.FileName,
			ExecutablePath: filepath.Join(filepath.Dir(jsonPath), tool.FileName),
			Description:    describe(tool.Description, tool.InputSchema),
			InputSchema:    tool.InputSchema,
		})
	}
	return customTools, nil
}

func (ts *toolSet) makeToolFromCustomTools(tool CustomToolDef) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		makeNameFromPath(tool.FileName),
		tool.Description,
		func(ctx context.Context, input inputSchema, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			args := make([]string, 0, len(input))
			for k, v := range input {
				args = append(args, k+"="+v)
			}
			ts.broker.Publish(ctx, broker.Message{
				Type:       broker.EventToolBash,
				Text:       strings.Join(args, " "),
				IsComplete: false,
			})
			script := fmt.Sprintf("'%s' \"$@\"", strings.ReplaceAll(tool.ExecutablePath, "'", `'\''`))
			out, err := executeBash(ctx, script, nil, args...)
			return fantasy.NewTextResponse(out), err
		},
	)
}

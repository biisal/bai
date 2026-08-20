package providers

import (
	"github.com/biisal/bai/internal/agent/core/tools"
	"github.com/openai/openai-go/v3"
)

func buildOpenAIToolParams(defs []tools.Definition) []openai.ChatCompletionToolUnionParam {
	params := make([]openai.ChatCompletionToolUnionParam, 0, len(defs))
	for _, def := range defs {
		params = append(params, openai.ChatCompletionFunctionTool(
			openai.FunctionDefinitionParam{
				Name:        string(def.Type),
				Description: openai.String(def.Description),
				Parameters:  openai.FunctionParameters(def.Parameters),
			},
		))
	}
	return params
}

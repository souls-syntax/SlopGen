package model

import (
	"encoding/json"

	"github.com/openai/openai-go/v3"
)

func ToRawJSONSchema(p Parameters) (map[string]any, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (t Tool) ConvertToOpenAITool() (openai.ChatCompletionToolUnionParam, error) {
	params, err := ToRawJSONSchema(t.Function.Parameters)
	if err != nil {
		return openai.ChatCompletionToolUnionParam{}, err
	}

	return openai.ChatCompletionFunctionTool(
		openai.FunctionDefinitionParam{
			Name:        t.Function.Name,
			Description: openai.String(t.Function.Description),
			Parameters:  openai.FunctionParameters(params),
		},
	), nil
}

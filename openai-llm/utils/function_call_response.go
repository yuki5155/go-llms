package utils

import (
	"encoding/json"
	"fmt"
)

type ChatCompletion struct {
	Choices           []Choice `json:"choices"`
	Created           int64    `json:"created"`
	ID                string   `json:"id"`
	Model             string   `json:"model"`
	Object            string   `json:"object"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Usage             Usage    `json:"usage"`
}

type Choice struct {
	FinishReason string      `json:"finish_reason"`
	Message      ChatMessage `json:"message"` // Message を ChatMessage に変更
}

type ChatMessage struct { // Message を ChatMessage に変更
	Content   json.RawMessage `json:"content"`
	Refusal   *string         `json:"refusal,omitempty"`
	Role      string          `json:"role"`
	ToolCalls []ToolCall      `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	Function Function `json:"function"`
	ID       string   `json:"id"`
	Type     string   `json:"type"`
}

type Function struct {
	Arguments string `json:"arguments"`
	Name      string `json:"name"`
}

type Usage struct {
	CompletionTokens        int                    `json:"completion_tokens"`
	CompletionTokensDetails CompletionTokenDetails `json:"completion_tokens_details"`
	PromptTokens            int                    `json:"prompt_tokens"`
	PromptTokensDetails     PromptTokenDetails     `json:"prompt_tokens_details"`
	TotalTokens             int                    `json:"total_tokens"`
}

type CompletionTokenDetails struct {
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	AudioTokens              int `json:"audio_tokens"`
	ReasoningTokens          int `json:"reasoning_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
}

type PromptTokenDetails struct {
	AudioTokens  int `json:"audio_tokens"`
	CachedTokens int `json:"cached_tokens"`
}

// GetFunctionCall はfunction名を指定してToolCallを取得します
func (c *ChatCompletion) GetFunctionCall(functionName string) (*ToolCall, error) {
	if len(c.Choices) == 0 {
		return nil, fmt.Errorf("no choices available")
	}

	for _, choice := range c.Choices {
		for _, toolCall := range choice.Message.ToolCalls {
			if toolCall.Function.Name == functionName {
				return &toolCall, nil
			}
		}
	}

	return nil, fmt.Errorf("function %s not found", functionName)
}

// GetAllFunctionCalls は指定されたfunction名の全てのToolCallを取得します
func (c *ChatCompletion) GetAllFunctionCalls(functionName string) ([]*ToolCall, error) {
	if len(c.Choices) == 0 {
		return nil, fmt.Errorf("no choices available")
	}

	var calls []*ToolCall
	for _, choice := range c.Choices {
		for i, toolCall := range choice.Message.ToolCalls {
			if toolCall.Function.Name == functionName {
				calls = append(calls, &choice.Message.ToolCalls[i])
			}
		}
	}

	if len(calls) == 0 {
		return nil, fmt.Errorf("function %s not found", functionName)
	}

	return calls, nil
}

func (c *ChatCompletion) GetMessages() []ChatMessage {
	if len(c.Choices) == 0 {
		return nil
	}

	var messages []ChatMessage
	for _, choice := range c.Choices {
		messages = append(messages, choice.Message)
	}

	return messages
}

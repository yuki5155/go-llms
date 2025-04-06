package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type ContentBlock struct {
	Type string          `json:"type"`
	Text string          `json:"text,omitempty"`
	JSON json.RawMessage `json:"json,omitempty"`
}

type ToolUse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolResultBlock struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
}

type MessageResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	StopReason   string         `json:"stop_reason"`
	StopSequence *string        `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	ToolUses []*ToolUse `json:"tool_uses,omitempty"`
}

type APIResponse struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Model   string          `json:"model"`
	Message MessageResponse `json:"message"`
}

// Parse structured JSON from Anthropic response
func ParseStructuredResponse[T any](content []ContentBlock) (*T, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("empty content blocks")
	}

	var jsonText string

	// For Claude API, the JSON might be returned in a text block with markdown code formatting
	for _, block := range content {
		if block.Type == "text" {
			// Check if text contains JSON wrapped in markdown code blocks
			text := block.Text

			// Use more robust regex to extract JSON from markdown
			jsonRegex := `(?s)\x60\x60\x60(?:json)?\s*({.*?})\s*\x60\x60\x60|\x60\x60\x60\s*({.*?})\s*\x60\x60\x60|^({.*})$`
			reg := regexp.MustCompile(jsonRegex)
			matches := reg.FindStringSubmatch(text)

			if len(matches) > 1 {
				// Find the first non-empty capture group
				for i := 1; i < len(matches); i++ {
					if matches[i] != "" {
						jsonText = strings.TrimSpace(matches[i])
						break
					}
				}
			} else if strings.HasPrefix(strings.TrimSpace(text), "{") && strings.HasSuffix(strings.TrimSpace(text), "}") {
				// Directly use the text if it appears to be JSON already
				jsonText = strings.TrimSpace(text)
			}

			// If we found valid JSON, break out of the loop
			if jsonText != "" {
				break
			}
		} else if block.Type == "json" {
			// If the API returns JSON directly in a JSON block
			jsonText = string(block.JSON)
			break
		}
	}

	if jsonText == "" {
		return nil, fmt.Errorf("no JSON content found in response")
	}

	// Print debug info about the extracted JSON
	fmt.Printf("DEBUG - Extracted JSON: %s\n", jsonText)

	var result T
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON content: %v", err)
	}

	return &result, nil
}

// Custom error type for response handling
type ResponseError struct {
	Type    string
	Message string
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func NewResponseError(errorType, message string) *ResponseError {
	return &ResponseError{
		Type:    errorType,
		Message: message,
	}
}

// Handle the response and convert to the specified type
func HandleResponse[T any](resp *APIResponse) (*T, error) {
	if resp == nil {
		return nil, NewResponseError("NullResponse", "response is nil")
	}

	// Debug the response
	fmt.Printf("DEBUG - Response: %+v\n", resp)

	// Parse structured JSON response
	result, err := ParseStructuredResponse[T](resp.Message.Content)
	if err != nil {
		return nil, NewResponseError("ParseError", fmt.Sprintf("error parsing response: %v", err))
	}

	return result, nil
}

// Check if an error is of a specific type
func ResponseErrorIs(err error, errorType string) bool {
	if respErr, ok := err.(*ResponseError); ok {
		return respErr.Type == errorType
	}
	return false
}

// Get a tool use by name
func (resp *APIResponse) GetToolUse(name string) (*ToolUse, error) {
	if resp == nil {
		return nil, NewResponseError("NullResponse", "response is nil")
	}

	if len(resp.Message.ToolUses) == 0 {
		return nil, NewResponseError("NoToolUses", "no tool uses in the response")
	}

	for _, toolUse := range resp.Message.ToolUses {
		if toolUse.Name == name {
			return toolUse, nil
		}
	}

	return nil, NewResponseError("ToolUseNotFound", fmt.Sprintf("tool use with name %s not found", name))
}

// Get all tool uses with a specific name
func (resp *APIResponse) GetAllToolUses(name string) ([]*ToolUse, error) {
	if resp == nil {
		return nil, NewResponseError("NullResponse", "response is nil")
	}

	if len(resp.Message.ToolUses) == 0 {
		return nil, NewResponseError("NoToolUses", "no tool uses in the response")
	}

	var toolUses []*ToolUse
	for _, toolUse := range resp.Message.ToolUses {
		if toolUse.Name == name {
			toolUses = append(toolUses, toolUse)
		}
	}

	if len(toolUses) == 0 {
		return nil, NewResponseError("ToolUseNotFound", fmt.Sprintf("tool use with name %s not found", name))
	}

	return toolUses, nil
}

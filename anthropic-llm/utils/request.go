package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	DefaultAPIEndpoint = "https://api.anthropic.com/v1/messages"
	DefaultModel       = "claude-3-opus-20240229"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type MediaType struct {
	Type   string `json:"type"`
	Data   string `json:"data,omitempty"`
	Source string `json:"source,omitempty"`
}

type Content struct {
	Type  string     `json:"type"`
	Text  string     `json:"text,omitempty"`
	Media *MediaType `json:"media,omitempty"`
}

type Message struct {
	Role    Role      `json:"role"`
	Content []Content `json:"content"`
}

type Tool struct {
	Type     string          `json:"type"`
	Function json.RawMessage `json:"function"`
}

// For Anthropic's special requirements, using a map for the request body
type RequestBody struct {
	Model          string          `json:"model"`
	Messages       []Message       `json:"messages"`
	System         string          `json:"system,omitempty"`
	Temperature    *float64        `json:"temperature,omitempty"`
	MaxTokens      *int            `json:"max_tokens,omitempty"`
	Tools          json.RawMessage `json:"tools,omitempty"`
	ToolChoice     string          `json:"tool_choice,omitempty"`
	ResponseFormat interface{}     `json:"response_format,omitempty"`
}

// ResponseFormat defines the format for API responses
type ResponseFormat struct {
	Type   string          `json:"type"`
	Schema json.RawMessage `json:"schema,omitempty"`
}

type ClientConfig struct {
	APIKey           string
	Endpoint         string
	Model            string
	MaxTokens        int
	Client           *http.Client
	AnthropicVersion string
}

func NewClientConfig(apiKey string) *ClientConfig {
	return &ClientConfig{
		APIKey:           apiKey,
		Endpoint:         DefaultAPIEndpoint,
		Model:            DefaultModel,
		MaxTokens:        4096,
		Client:           &http.Client{},
		AnthropicVersion: "2023-06-01",
	}
}

type Client struct {
	config *ClientConfig
}

func NewClient(config *ClientConfig) *Client {
	return &Client{config: config}
}

type RequestOptions struct {
	Messages    []Message
	Schema      json.RawMessage
	System      string
	MaxTokens   *int
	Temperature *float64
}

func NewTextMessage(role Role, text string) Message {
	return Message{
		Role: role,
		Content: []Content{
			{
				Type: "text",
				Text: text,
			},
		},
	}
}

func NewMessageWithImage(imageUrl string, text string) Message {
	return Message{
		Role: RoleUser,
		Content: []Content{
			{
				Type: "text",
				Text: text,
			},
			{
				Type: "media",
				Media: &MediaType{
					Type:   "image",
					Source: imageUrl,
				},
			},
		},
	}
}

func NewMessageWithImageBase64(imageBytes []byte, text string) Message {
	base64Str := base64.StdEncoding.EncodeToString(imageBytes)

	return Message{
		Role: RoleUser,
		Content: []Content{
			{
				Type: "text",
				Text: text,
			},
			{
				Type: "media",
				Media: &MediaType{
					Type: "image",
					Data: fmt.Sprintf("data:image/jpeg;base64,%s", base64Str),
				},
			},
		},
	}
}

func (c *Client) SendRequestWithTool(opts RequestOptions) (*APIResponse, error) {
	if len(opts.Messages) == 0 {
		return nil, fmt.Errorf("at least one message is required")
	}

	maxTokens := c.config.MaxTokens
	if opts.MaxTokens != nil {
		maxTokens = *opts.MaxTokens
	}

	reqBody := RequestBody{
		Model:      c.config.Model,
		Messages:   opts.Messages,
		System:     opts.System,
		MaxTokens:  &maxTokens,
		Tools:      opts.Schema,
		ToolChoice: "auto",
	}

	if opts.Temperature != nil {
		reqBody.Temperature = opts.Temperature
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %v", err)
	}

	// Debug: Print the actual JSON being sent
	fmt.Printf("DEBUG - Request JSON: %s\n", string(jsonData))

	req, err := http.NewRequest("POST", c.config.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", c.config.AnthropicVersion)

	resp, err := c.config.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &apiResp, nil
}

func (c *Client) SendRequestWithStructuredOutput(opts RequestOptions) (*APIResponse, error) {
	if len(opts.Messages) == 0 {
		return nil, fmt.Errorf("at least one message is required")
	}

	maxTokens := c.config.MaxTokens
	if opts.MaxTokens != nil {
		maxTokens = *opts.MaxTokens
	}

	// Add JSON instruction to the system prompt
	systemPrompt := opts.System
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant."
	}
	systemPrompt += " Return your response as a valid JSON object. Do not include any explanations or text outside of the JSON object."

	// Create the basic request body without response_format
	reqBody := map[string]interface{}{
		"model":      c.config.Model,
		"messages":   opts.Messages,
		"system":     systemPrompt,
		"max_tokens": maxTokens,
		// No response_format field since it's causing errors
	}

	if opts.Temperature != nil {
		reqBody["temperature"] = *opts.Temperature
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %v", err)
	}

	// Debug: Print the actual JSON being sent
	fmt.Printf("DEBUG - Request JSON: %s\n", string(jsonData))

	req, err := http.NewRequest("POST", c.config.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", c.config.AnthropicVersion)

	resp, err := c.config.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &apiResp, nil
}

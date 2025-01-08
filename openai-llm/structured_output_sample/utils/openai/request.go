package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	DefaultAPIEndpoint = "https://api.openai.com/v1/chat/completions"
	DefaultModel       = "gpt-4o-2024-08-06"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type RequestFormat struct {
	Type       string          `json:"type"`
	JSONSchema json.RawMessage `json:"json_schema"`
}

type RequestBody struct {
	Model          string          `json:"model"`
	Messages       []Message       `json:"messages"`
	ResponseFormat RequestFormat   `json:"response_format,omitempty"`
	Tools          json.RawMessage `json:"tools,omitempty"`
}

type ClientConfig struct {
	APIKey   string
	Endpoint string
	Model    string
	Client   *http.Client
}

func NewClientConfig(apiKey string) *ClientConfig {
	return &ClientConfig{
		APIKey:   apiKey,
		Endpoint: DefaultAPIEndpoint,
		Model:    DefaultModel,
		Client:   &http.Client{},
	}
}

type Client struct {
	config *ClientConfig
}

func NewClient(config *ClientConfig) *Client {
	return &Client{config: config}
}

type RequestOptions struct {
	Messages []Message
	Schema   json.RawMessage
}

func NewMessage(role Role, content string) Message {
	return Message{
		Role:    role,
		Content: content,
	}
}

type functionCallRequestBody struct {
	Model    string          `json:"model"`
	Messages []Message       `json:"messages"`
	Tools    json.RawMessage `json:"tools"`
}

func (c *Client) SendRequestWithFunctionCall(opts RequestOptions) (*ChatCompletion, error) {
	if len(opts.Messages) == 0 {
		return nil, fmt.Errorf("at least one message is required")
	}

	reqBody := functionCallRequestBody{
		Model:    c.config.Model,
		Messages: opts.Messages,
		Tools:    opts.Schema,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %v", err)
	}
	req, err := http.NewRequest("POST", c.config.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

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
	var parsedJson map[string]interface{}
	if err := json.Unmarshal(body, &parsedJson); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	var completion *ChatCompletion
	if err := json.Unmarshal(body, &completion); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return completion, nil

}

func (c *Client) SendRequestWithStructuredOutput(opts RequestOptions) (*APIResponse, error) {
	if len(opts.Messages) == 0 {
		return nil, fmt.Errorf("at least one message is required")
	}

	reqBody := RequestBody{
		Model:    c.config.Model,
		Messages: opts.Messages,
		ResponseFormat: RequestFormat{
			Type:       "json_schema",
			JSONSchema: opts.Schema,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %v", err)
	}

	req, err := http.NewRequest("POST", c.config.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

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

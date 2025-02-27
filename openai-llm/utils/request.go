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
	DefaultAPIEndpoint = "https://api.openai.com/v1/chat/completions"
	DefaultModel       = "gpt-4o-2024-08-06"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type ImageUrl struct {
	Url string `json:"url"`
}

type Content struct {
	Text     string    `json:"text,omitempty"`
	Type     string    `json:"type,omitempty"`
	ImageUrl *ImageUrl `json:"image_url,omitempty"`
}

type Message struct {
	Role    Role            `json:"role"`
	Content json.RawMessage `json:"content"`
}

type RequestFormat struct {
	Type       string          `json:"type"`
	JSONSchema json.RawMessage `json:"json_schema"`
}

type RequestBody struct {
	Model          string          `json:"model"`
	Messages       []Message       `json:"messages"`
	ResponseFormat *RequestFormat  `json:"response_format,omitempty"`
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
	// コンテンツを文字列としてJSON形式にエンコード
	contentBytes, err := json.Marshal(content)
	if err != nil {
		// エラーハンドリングが必要な場合は、
		// 関数のシグネチャを (Message, error) に変更することを検討
		return Message{}
	}

	return Message{
		Role:    role,
		Content: contentBytes,
	}
}

func NewMessageWithImage(imageUrl string, text string) Message {
	imageContent := Content{
		Type: "image_url",
		ImageUrl: &ImageUrl{
			Url: imageUrl,
		},
	}

	messageContent := Content{
		Text: text,
		Type: "text",
	}

	contentBytes, _ := json.Marshal([]Content{imageContent, messageContent})

	return Message{
		Role:    "user",
		Content: contentBytes,
	}
}

func NewMessageWithImageBase64(imageBytes []byte, text string) Message {
	base64Str := base64.StdEncoding.EncodeToString(imageBytes)
	dataURL := fmt.Sprintf("data:image/jpeg;base64,%s", base64Str)
	imageContent := Content{
		Type: "image_url",
		ImageUrl: &ImageUrl{
			Url: dataURL,
		},
	}
	messageContent := Content{
		Text: text,
		Type: "text",
	}
	contentBytes, _ := json.Marshal([]Content{imageContent, messageContent})
	return Message{
		Role:    "user",
		Content: contentBytes,
	}
}

func (c *Client) SendRequestWithFunctionCall(opts RequestOptions) (*ChatCompletion, error) {
	if len(opts.Messages) == 0 {
		return nil, fmt.Errorf("at least one message is required")
	}

	reqBody := RequestBody{
		Model:          c.config.Model,
		Messages:       opts.Messages,
		Tools:          opts.Schema,
		ResponseFormat: nil,
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
		ResponseFormat: &RequestFormat{
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

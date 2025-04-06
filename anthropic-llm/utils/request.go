package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/yuki5155/go-llms/anthropic-llm/schema"
)

const (
	DefaultAPIEndpoint = "https://api.anthropic.com/v1/messages"
	DefaultModel       = "claude-3-7-sonnet-20250219"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type MediaType struct {
	Source *ImageSource `json:"source"`
}

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	Url       string `json:"url,omitempty"`
}

type Content struct {
	Type   string       `json:"type"`
	Text   string       `json:"text,omitempty"`
	Source *ImageSource `json:"source,omitempty"`
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
	Model       string
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
				Type: "image",
				Source: &ImageSource{
					Type: "url",
					Url:  imageUrl,
				},
			},
		},
	}
}

func NewMessageWithImageBase64(imageBytes []byte, text string) Message {
	// Encode the image to base64
	base64Image := base64.StdEncoding.EncodeToString(imageBytes)

	// Detect mime type - for now assuming jpeg
	mediaType := "image/jpeg"

	// Create the content array with text first, followed by image (may be more compatible with Anthropic's expected format)
	content := []Content{
		{
			Type: "text",
			Text: text,
		},
		{
			Type: "image",
			Source: &ImageSource{
				Type:      "base64",
				MediaType: mediaType,
				Data:      base64Image,
			},
		},
	}

	// Return the message with the content
	return Message{
		Role:    RoleUser,
		Content: content,
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
		return nil, fmt.Errorf("error parsing response: %v, body: %s", err, string(body))
	}

	// Fix content if it's not correctly parsed
	if len(apiResp.Message.Content) == 0 {
		// Try to extract content from the raw response
		var rawResp map[string]interface{}
		if err := json.Unmarshal(body, &rawResp); err == nil {
			if content, ok := rawResp["content"].([]interface{}); ok && len(content) > 0 {
				// Convert the raw content to ContentBlock
				for _, c := range content {
					if contentMap, ok := c.(map[string]interface{}); ok {
						contentType, _ := contentMap["type"].(string)
						contentText, _ := contentMap["text"].(string)
						apiResp.Message.Content = append(apiResp.Message.Content, ContentBlock{
							Type: contentType,
							Text: contentText,
						})
					}
				}
			}
		}
	}

	return &apiResp, nil
}

// generateJSONTemplate generates a JSON template from a schema
func generateJSONTemplate(schemaData json.RawMessage) (string, error) {
	var baseSchema schema.BaseSchema
	if err := json.Unmarshal(schemaData, &baseSchema); err == nil {
		return generateJSONTemplateFromSchema(baseSchema)
	}

	// If we couldn't parse the schema, provide a generic instruction
	return "You MUST respond with valid JSON only. Your entire response should be valid parseable JSON without any text before or after.", nil
}

// formatJSONValue formats a value for inclusion in a JSON template
func formatJSONValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", val)
	case int, int64, float64, float32:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%v", val)
	case []string:
		return fmt.Sprintf("[%s]", strings.Join(val, ", "))
	case map[string]interface{}:
		return "{...}" // Simplified for nested objects
	default:
		return "\"<value>\""
	}
}

// generateRules generates rules based on schema properties
func generateRules(schema schema.BaseSchema) string {
	rules := []string{
		"1. Do not add any fields not shown above",
		"2. Do not nest any fields or create additional objects",
		"3. Do not include notes about data being samples",
		"4. Your entire response must be valid JSON only, with no text before or after",
	}

	// Add specific rules for required fields if any
	if len(schema.Required) > 0 {
		requiredFields := strings.Join(schema.Required, ", ")
		rules = append(rules, fmt.Sprintf("5. The following fields are REQUIRED: %s", requiredFields))
	}

	// Add type-specific rules
	stringFields := []string{}
	numberFields := []string{}
	integerFields := []string{}
	booleanFields := []string{}

	for name, prop := range schema.Properties {
		switch prop.Type {
		case "string":
			stringFields = append(stringFields, name)
		case "number":
			numberFields = append(numberFields, name)
		case "integer":
			integerFields = append(integerFields, name)
		case "boolean":
			booleanFields = append(booleanFields, name)
		}
	}

	ruleNum := 6

	if len(stringFields) > 0 {
		rules = append(rules, fmt.Sprintf("%d. All string fields (%s) MUST be simple STRING values, not objects or arrays",
			ruleNum, strings.Join(stringFields, ", ")))
		ruleNum++
	}

	if len(numberFields) > 0 {
		rules = append(rules, fmt.Sprintf("%d. All number fields (%s) MUST be NUMERIC values without quotes or units (e.g., 23.5 NOT \"23.5\" or \"23.5°C\")",
			ruleNum, strings.Join(numberFields, ", ")))
		ruleNum++
	}

	if len(integerFields) > 0 {
		rules = append(rules, fmt.Sprintf("%d. All integer fields (%s) MUST be INTEGER values without quotes or units (e.g., 42 NOT \"42\")",
			ruleNum, strings.Join(integerFields, ", ")))
		ruleNum++
	}

	if len(booleanFields) > 0 {
		rules = append(rules, fmt.Sprintf("%d. All boolean fields (%s) MUST be true or false without quotes (e.g., true NOT \"true\")",
			ruleNum, strings.Join(booleanFields, ", ")))
		ruleNum++
	}

	return strings.Join(rules, "\n")
}

// SendRequestWithStructuredOutput sends a request to Anthropic API and returns the structured output
func (c *Client) SendRequestWithStructuredOutput(opts RequestOptions) (*APIResponse, error) {
	if len(opts.Messages) == 0 {
		return nil, fmt.Errorf("at least one message is required")
	}

	// Check if schema is provided
	if opts.Schema == nil {
		return nil, fmt.Errorf("schema is required for structured output")
	}

	model := DefaultModel
	if opts.Model != "" {
		model = opts.Model
	}

	// Initialize request body - using RequestBody struct to ensure correct format
	reqBody := RequestBody{
		Model:    model,
		Messages: opts.Messages,
	}

	// Add system prompt if provided
	systemMsg := opts.System
	if systemMsg == "" {
		systemMsg = "You are a helpful assistant."
	}

	// Generate template from schema
	var baseSchema schema.BaseSchema
	if err := json.Unmarshal(opts.Schema, &baseSchema); err == nil {
		template, err := generateJSONTemplateFromSchema(baseSchema)
		if err != nil {
			reqBody.System = systemMsg + " You MUST respond with valid JSON only. Your entire response should be valid parseable JSON without any text before or after."
		} else {
			reqBody.System = systemMsg + " " + template
		}
	} else {
		// Fallback to a generic instruction
		reqBody.System = systemMsg + " You MUST respond with valid JSON only. If you don't have access to real-time data, provide realistic sample data in valid JSON format. DO NOT add any notes or explanations outside the JSON structure."
	}

	// Add max tokens - required for Anthropic API, especially for image input
	maxTokens := c.config.MaxTokens
	if opts.MaxTokens != nil {
		maxTokens = *opts.MaxTokens
	}
	reqBody.MaxTokens = &maxTokens

	// Add temperature if provided
	if opts.Temperature != nil {
		reqBody.Temperature = opts.Temperature
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
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.config.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v, body: %s", err, string(body))
	}

	// Fix content if it's not correctly parsed
	if len(apiResp.Message.Content) == 0 {
		// Try to extract content from the raw response
		var rawResp map[string]interface{}
		if err := json.Unmarshal(body, &rawResp); err == nil {
			if content, ok := rawResp["content"].([]interface{}); ok && len(content) > 0 {
				// Convert the raw content to ContentBlock
				for _, c := range content {
					if contentMap, ok := c.(map[string]interface{}); ok {
						contentType, _ := contentMap["type"].(string)
						contentText, _ := contentMap["text"].(string)
						apiResp.Message.Content = append(apiResp.Message.Content, ContentBlock{
							Type: contentType,
							Text: contentText,
						})
					}
				}
			}
		}
	}

	return &apiResp, nil
}

// generateJSONTemplateFromSchema generates a JSON template from a BaseSchema
func generateJSONTemplateFromSchema(baseSchema schema.BaseSchema) (string, error) {
	// Extract properties and generate template
	exampleValues := map[string]interface{}{
		"string":  "example string",
		"integer": 42,
		"number":  23.5,
		"boolean": true,
		"array":   []string{"item1", "item2"},
		"object":  map[string]interface{}{"key": "value"},
	}

	// Build template JSON
	templateJSON := "{\n"
	var lines []string

	// Process properties in a deterministic order (alphabetical)
	propNames := make([]string, 0, len(baseSchema.Properties))
	for name := range baseSchema.Properties {
		propNames = append(propNames, name)
	}
	sort.Strings(propNames)

	for _, name := range propNames {
		prop := baseSchema.Properties[name]
		var exampleValue interface{}

		// Use generic example based on type
		if prop.Type != "" {
			if val, ok := exampleValues[prop.Type]; ok {
				exampleValue = val
			} else {
				exampleValue = fmt.Sprintf("<%s>", prop.Type)
			}

			// Use property description for example if available (only for strings)
			if prop.Description != "" && prop.Type == "string" {
				// Truncate if too long
				if len(prop.Description) > 30 {
					exampleValue = prop.Description[:27] + "..."
				} else {
					exampleValue = prop.Description
				}
			}
		} else {
			exampleValue = "<any value>"
		}

		// Format the property
		line := fmt.Sprintf("  \"%s\": %s", name, formatJSONValue(exampleValue))
		lines = append(lines, line)
	}

	templateJSON += strings.Join(lines, ",\n")
	templateJSON += "\n}"

	// Generate rules based on the properties
	rules := generateRules(baseSchema)

	template := fmt.Sprintf("You MUST respond with EXACTLY this JSON format:\n%s\n\nIMPORTANT RULES:\n%s",
		templateJSON, rules)

	return template, nil
}

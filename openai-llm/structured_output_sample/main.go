package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const openaiAPIEndpoint = "https://api.openai.com/v1/chat/completions"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model          string    `json:"model"`
	Messages       []Message `json:"messages"`
	ResponseFormat struct {
		Type       string          `json:"type"`
		JSONSchema json.RawMessage `json:"json_schema"`
	} `json:"response_format"`
}

type ResponseChoice struct {
	Message struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
		Refusal *string         `json:"refusal,omitempty"`
	} `json:"message"`
	FinishReason string `json:"finish_reason"`
}

type APIResponse struct {
	Choices []ResponseChoice `json:"choices"`
}

type JSONSchemaWrapper struct {
	Name   string        `json:"name"`
	Schema WeatherSchema `json:"schema"`
}

type WeatherSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]SchemaProperty `json:"properties"`
	Required   []string                  `json:"required"`
}

type SchemaProperty struct {
	Type string   `json:"type"`
	Enum []string `json:"enum,omitempty"`
}

func createWeatherSchema() JSONSchemaWrapper {
	return JSONSchemaWrapper{
		Name: "weather_response",
		Schema: WeatherSchema{
			Type: "object",
			Properties: map[string]SchemaProperty{
				"location":    {Type: "string"},
				"temperature": {Type: "number"},
				"unit": {
					Type: "string",
					Enum: []string{"C", "F"},
				},
				"conditions": {Type: "string"},
			},
			Required: []string{"location", "temperature", "unit", "conditions"},
		},
	}
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the OPENAI_API_KEY environment variable.")
		return
	}

	weatherSchema := createWeatherSchema()
	weatherSchemaJSON, err := json.Marshal(weatherSchema)
	if err != nil {
		fmt.Printf("Error marshalling weather schema: %v\n", err)
		return
	}

	reqBody := RequestBody{
		Model: "gpt-4o-2024-08-06",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant designed to output JSON about weather conditions.",
			},
			{
				Role:    "user",
				Content: "What's the weather like in Tokyo today?",
			},
		},
		ResponseFormat: struct {
			Type       string          `json:"type"`
			JSONSchema json.RawMessage `json:"json_schema"`
		}{
			Type:       "json_schema",
			JSONSchema: weatherSchemaJSON,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Error marshalling request: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", openaiAPIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API returned non-200 status code: %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Response body: %s\n", string(body))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	// Print the raw response for debugging
	fmt.Printf("Raw API response:\n%s\n", string(body))

	var apiResp APIResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if len(apiResp.Choices) > 0 {
		choice := apiResp.Choices[0]

		// Handle different finish reasons
		switch choice.FinishReason {
		case "stop":
			fmt.Println("Received structured output. Writing to file...")

			// The content is a JSON string, so we need to unmarshal it twice
			var jsonString string
			err := json.Unmarshal(choice.Message.Content, &jsonString)
			if err != nil {
				fmt.Printf("Error parsing JSON string: %v\n", err)
				return
			}

			var structuredOutput map[string]interface{}
			err = json.Unmarshal([]byte(jsonString), &structuredOutput)
			if err != nil {
				fmt.Printf("Error parsing structured output: %v\n", err)
				return
			}

			// Then, marshal it back to JSON with indentation
			prettyJSON, err := json.MarshalIndent(structuredOutput, "", "  ")
			if err != nil {
				fmt.Printf("Error formatting JSON: %v\n", err)
				return
			}

			// Write to file
			err = os.WriteFile("structured_output.json", prettyJSON, 0644)
			if err != nil {
				fmt.Printf("Error writing to file: %v\n", err)
				return
			}
			fmt.Println("Structured output written to structured_output.json")
		case "length":
			fmt.Println("Warning: The response was truncated due to token limit.")
		case "content_filter":
			fmt.Println("Warning: The response was filtered due to content restrictions.")
		default:
			fmt.Printf("Unexpected finish reason: %s\n", choice.FinishReason)
		}

		// Handle refusals
		if choice.Message.Refusal != nil {
			fmt.Printf("The model refused to generate a response: %s\n", *choice.Message.Refusal)
		}
	} else {
		fmt.Println("No choices in the API response.")
	}
}

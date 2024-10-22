package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sample/schema"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model          string               `json:"model"`
	Messages       []Message            `json:"messages"`
	ResponseFormat schema.RequestFormat `json:"response_format"`
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

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the OPENAI_API_KEY environment variable.")
		return
	}

	// WeatherSchemaの作成
	weatherSchema := schema.NewWeatherSchema()
	schemaJSON, err := json.Marshal(weatherSchema)
	if err != nil {
		fmt.Printf("Error marshalling weather schema: %v\n", err)
		return
	}

	// リクエストボディの作成
	reqBody := RequestBody{
		Model: "gpt-4o-2024-08-06",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant designed to output weather information in JSON format.",
			},
			{
				Role:    "user",
				Content: "What's the weather like in Tokyo today?",
			},
		},
		ResponseFormat: schema.RequestFormat{
			Type:       "json_schema",
			JSONSchema: schemaJSON,
		},
	}

	// OpenAI APIにリクエストを送信
	resp, err := sendRequest(reqBody, apiKey)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}

	// レスポンスの処理
	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]

		switch choice.FinishReason {
		case "stop":
			if choice.Message.Refusal != nil {
				fmt.Printf("The model refused to generate a response: %s\n", *choice.Message.Refusal)
				return
			}

			// デバッグ出力
			fmt.Printf("Raw content: %s\n", string(choice.Message.Content))

			// 最初のJSONアンマーシャル：文字列として取得
			var jsonString string
			if err := json.Unmarshal(choice.Message.Content, &jsonString); err != nil {
				fmt.Printf("Error unmarshaling outer JSON: %v\n", err)
				return
			}

			// 文字列からWeatherResponseへのアンマーシャル
			var weather schema.WeatherResponse
			if err := json.Unmarshal([]byte(jsonString), &weather); err != nil {
				fmt.Printf("Error unmarshaling inner JSON: %v\n", err)
				return
			}

			// レスポンスの表示
			prettyJSON, err := json.MarshalIndent(weather, "", "  ")
			if err != nil {
				fmt.Printf("Error formatting response: %v\n", err)
				return
			}
			fmt.Printf("Weather Information:\n%s\n", string(prettyJSON))

		case "length":
			fmt.Println("Warning: The response was truncated due to token limit.")
		case "content_filter":
			fmt.Println("Warning: The response was filtered due to content restrictions.")
		default:
			fmt.Printf("Unexpected finish reason: %s\n", choice.FinishReason)
		}
	}
}

func sendRequest(reqBody RequestBody, apiKey string) (*APIResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
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

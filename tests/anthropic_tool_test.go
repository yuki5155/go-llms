package tests

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	anthropicSchema "github.com/yuki5155/go-llms/anthropic-llm/schema"
	anthropicUtils "github.com/yuki5155/go-llms/anthropic-llm/utils"
)

func loadEnvForTool(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

func TestAnthropicToolUsage(t *testing.T) {
	// Load environment variables from .env file
	err := loadEnvForTool("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Skip test if env var is not set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping test")
	}

	// Create client
	config := anthropicUtils.NewClientConfig(apiKey)
	client := anthropicUtils.NewClient(config)

	// Create the weather tool schema
	weatherTool := anthropicSchema.NewWeatherToolSchema()
	tools := []anthropicSchema.Tool{*weatherTool}
	toolsJSON, err := json.Marshal(tools)
	if err != nil {
		t.Fatalf("Error marshalling weather tool schema: %v", err)
	}

	// Create messages
	systemMessage := "You are a helpful assistant that provides accurate weather information."
	messages := []anthropicUtils.Message{
		anthropicUtils.NewTextMessage(anthropicUtils.RoleUser, "What's the current weather in Kyoto?"),
	}

	// Create request options
	opts := anthropicUtils.RequestOptions{
		Messages: messages,
		System:   systemMessage,
		Schema:   toolsJSON,
	}

	// Send request to Anthropic API
	resp, err := client.SendRequestWithTool(opts)
	if err != nil {
		t.Fatalf("Error sending request: %v", err)
	}

	// Get tool call response
	toolUse, err := resp.GetToolUse("get_weather")
	if err != nil {
		t.Fatalf("Error getting tool use: %v", err)
	}

	// Parse arguments
	var args map[string]interface{}
	err = json.Unmarshal([]byte(toolUse.Arguments), &args)
	if err != nil {
		t.Fatalf("Error parsing tool arguments: %v", err)
	}

	// Validate arguments
	location, ok := args["location"].(string)
	if !ok {
		t.Fatalf("Expected location to be a string")
	}

	if location != "Kyoto" {
		t.Errorf("Expected location to be Kyoto, got %s", location)
	}

	// Display results
	fmt.Printf("Tool Name: %s\n", toolUse.Name)
	fmt.Printf("Tool Arguments: %s\n", toolUse.Arguments)
}

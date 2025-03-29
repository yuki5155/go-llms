package tests

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	anthropicUtils "github.com/yuki5155/go-llms/anthropic-llm/utils"
)

func loadEnvForImage(filename string) error {
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

func TestAnthropicImageInput(t *testing.T) {
	// Load environment variables from .env file
	err := loadEnvForImage("../.env")
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

	// Load test image
	imagePath := filepath.Join("images", "test_image.jpg")
	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		t.Fatalf("Error reading image file: %v", err)
	}

	// Create message with image
	message := anthropicUtils.NewMessageWithImageBase64(
		imageBytes,
		"Please describe what you see in this image.",
	)

	// Create system message
	systemMessage := "You are a helpful assistant that describes images accurately."

	// Create request options
	opts := anthropicUtils.RequestOptions{
		Messages: []anthropicUtils.Message{message},
		System:   systemMessage,
	}

	// Send request to Anthropic API
	resp, err := client.SendRequestWithStructuredOutput(opts)
	if err != nil {
		t.Fatalf("Error sending request: %v", err)
	}

	// Check that we got a response
	if len(resp.Message.Content) == 0 {
		t.Fatalf("No content in response")
	}

	// Display text from the first content block
	var responseText string
	for _, block := range resp.Message.Content {
		if block.Type == "text" {
			responseText = block.Text
			break
		}
	}

	fmt.Println("Image description:")
	fmt.Println(responseText)
}

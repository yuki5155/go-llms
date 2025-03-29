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

func loadEnvForStructured(filename string) error {
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

func TestAnthropicStructuredOutput(t *testing.T) {
	// Load environment variables from .env file
	err := loadEnvForStructured("../.env")
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

	// Create the weather schema
	weatherSchema := anthropicSchema.NewWeatherSchema()
	schemaJSON, err := json.Marshal(weatherSchema)
	if err != nil {
		t.Fatalf("Error marshalling weather schema: %v", err)
	}

	// Create messages
	systemMessage := "You are a helpful assistant that provides accurate weather information."
	messages := []anthropicUtils.Message{
		anthropicUtils.NewTextMessage(anthropicUtils.RoleUser, "What's the weather like in Tokyo today?"),
	}

	// Create request options
	opts := anthropicUtils.RequestOptions{
		Messages: messages,
		System:   systemMessage,
		Schema:   schemaJSON,
	}

	// Send request to Anthropic API
	resp, err := client.SendRequestWithStructuredOutput(opts)
	if err != nil {
		t.Fatalf("Error sending request: %v", err)
	}

	// Process the response
	weather, err := anthropicUtils.HandleResponse[anthropicSchema.WeatherResponse](resp)
	if err != nil {
		t.Fatalf("Error handling response: %v", err)
	}

	// Validate response
	if weather.Location != "Tokyo" {
		t.Errorf("Expected location to be Tokyo, got %s", weather.Location)
	}

	if weather.Unit != "C" && weather.Unit != "F" {
		t.Errorf("Expected unit to be either C or F, got %s", weather.Unit)
	}

	fmt.Printf("Weather in %s:\n", weather.Location)
	fmt.Printf("Temperature: %.1f %s\n", weather.Temperature, weather.Unit)
	fmt.Printf("Forecast: %s\n", weather.Forecast)
	fmt.Printf("Humidity: %d%%\n", weather.Humidity)
	fmt.Printf("Wind: %.1f, Direction: %s\n", weather.WindSpeed, weather.WindDirection)
	fmt.Printf("Precipitation: %d%%\n", weather.PrecipitationPct)
}

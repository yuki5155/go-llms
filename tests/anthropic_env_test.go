package tests

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
)

func loadEnvTest(filename string) error {
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

func TestEnvironmentLoading(t *testing.T) {
	// Clear any existing value
	os.Unsetenv("ANTHROPIC_API_KEY")

	// Load environment variables from .env file
	err := loadEnvTest("../.env")
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Check if ANTHROPIC_API_KEY was loaded
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	fmt.Printf("ANTHROPIC_API_KEY value: %s\n", maskAPIKey(apiKey))

	// Check if OPENAI_API_KEY was loaded
	openaiKey := os.Getenv("OPENAI_API_KEY")
	fmt.Printf("OPENAI_API_KEY value: %s\n", maskAPIKey(openaiKey))
}

// maskAPIKey returns a masked version of the API key for security
func maskAPIKey(key string) string {
	if key == "" {
		return "<not set>"
	}

	if len(key) <= 8 {
		return "****" + key[len(key)-4:]
	}

	return key[:4] + "****" + key[len(key)-4:]
}

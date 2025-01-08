package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sample/schema"
	"sample/utils/openai"
	"strings"
	"testing"
)

func TestSample(t *testing.T) {
	fmt.Println("This is a sample.")
}

func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// コメントや空行をスキップ
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

func TestFunctionCall(t *testing.T) {
	// load .env file
	err := loadEnv(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the OPENAI_API_KEY environment variable.")
		return
	}
	// OpenAIクライアントの設定と作成
	config := openai.NewClientConfig(apiKey)
	client := openai.NewClient(config)
	weatherSchema := schema.NewWeatherFunctionCallSchema()
	tools := []schema.Tool{*weatherSchema}
	toolsJSON, err := json.Marshal(tools)
	if err != nil {
		fmt.Printf("Error marshalling weather schema: %v\n", err)
		return
	}
	messages := []openai.Message{
		openai.NewMessage(openai.RoleSystem, "You are a helpful assistant designed to output weather information in JSON format."),
		openai.NewMessage(openai.RoleUser, "What's the weather like in Tokyo today?"),
	}
	opts := openai.RequestOptions{
		Messages: messages,
		Schema:   toolsJSON,
	}
	res, err := client.SendRequestWithFunctionCall(opts)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}

	item, err := res.GetAllFunctionCalls("weather")
	if err != nil {
		fmt.Printf("Error handling response: %v\n", err)
		return
	}
	fmt.Println(item[0].Function.Arguments)

}

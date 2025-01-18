package tests

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/yuki5155/go-llms/openai-llm/schema"
	"github.com/yuki5155/go-llms/openai-llm/utils"
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

func TestImageAnalyze(t *testing.T) {
	err := loadEnv("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the OPENAI_API_KEY environment variable.")
		return
	}
	// utilsクライアントの設定と作成
	config := utils.NewClientConfig(apiKey)
	client := utils.NewClient(config)
	weatherSchema := schema.NewWeatherFunctionCallSchema()
	tools := []schema.Tool{*weatherSchema}
	toolsJSON, err := json.Marshal(tools)
	if err != nil {
		fmt.Printf("Error marshalling weather schema: %v\n", err)
		return
	}
	messages := []utils.Message{
		utils.NewMessageWithImage("https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg", "tell me the iamge"),
	}
	opts := utils.RequestOptions{
		Messages: messages,
		Schema:   toolsJSON,
	}
	res, err := client.SendRequestWithFunctionCall(opts)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}

	fmt.Println(res.GetMessages()[0].Content)

}

// load a image from dir and send it to openai

// image analyze with structured_output
func TestImageAnalyzeWithStructuredOutput(t *testing.T) {
	err := loadEnv("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set the OPENAI_API_KEY environment variable.")
		return
	}
	// utilsクライアントの設定と作成
	config := utils.NewClientConfig(apiKey)
	client := utils.NewClient(config)
	imageSchema := schema.NewImageAnalysisSchema()
	schemaJSON, err := json.Marshal(imageSchema)
	if err != nil {
		fmt.Printf("Error marshalling weather schema: %v\n", err)
		return
	}
	messages := []utils.Message{
		utils.NewMessageWithImage("https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg", "tell me the iamge"),
	}
	opts := utils.RequestOptions{
		Messages: messages,
		Schema:   schemaJSON,
	}
	res, err := client.SendRequestWithStructuredOutput(opts)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	imageAnalyze, err := utils.HandleResponse[schema.ImageAnalysisResponse](res)
	if err != nil {
		fmt.Printf("Error handling response: %v\n", err)
		return
	}
	fmt.Println(imageAnalyze.Category)
	fmt.Println(imageAnalyze.Description)
	fmt.Println(imageAnalyze.Objects)

}

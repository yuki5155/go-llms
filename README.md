## How to Setup

```
go get github.com/yuki5155/go-llms@v1.0.0
```

## How to Use

### Structured Output

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yuki5155/go-llms/openai-llm/schema"
	"github.com/yuki5155/go-llms/openai-llm/utils"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the OPENAI_API_KEY environment variable.")
		return
	}

	// utilsクライアントの設定と作成
	config := utils.NewClientConfig(apiKey)
	client := utils.NewClient(config)

	// WeatherSchemaの作成
	weatherSchema := schema.NewWeatherSchema()
	schemaJSON, err := json.Marshal(weatherSchema)
	if err != nil {
		fmt.Printf("Error marshalling weather schema: %v\n", err)
		return
	}

	// メッセージの準備
	messages := []utils.Message{
		utils.NewMessage(utils.RoleSystem, "You are a helpful assistant designed to output weather information in JSON format."),
		utils.NewMessage(utils.RoleUser, "What's the weather like in Tokyo today?"),
	}

	// リクエストオプションの作成
	opts := utils.RequestOptions{
		Messages: messages,
		Schema:   schemaJSON,
	}

	// utils APIにリクエストを送信
	resp, err := client.SendRequestWithStructuredOutput(opts)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}

	// レスポンスの処理
	weather, err := utils.HandleResponse[schema.WeatherResponse](resp)
	if err != nil {
		fmt.Printf("Error handling response: %v\n", err)
		return
	}

	// キーを指定して情報を取得
	fmt.Printf("Temperature: %v\n", weather.Temperature)

	// 結果の表示
	prettyJSON, err := json.MarshalIndent(weather, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting response: %v\n", err)
		return
	}
	fmt.Printf("Weather Information:\n%s\n", string(prettyJSON))

}

```

## function call

```go
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
		utils.NewMessage(utils.RoleSystem, "You are a helpful assistant designed to output weather information in JSON format."),
		utils.NewMessage(utils.RoleUser, "What's the weather like in Tokyo today?"),
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

	item, err := res.GetAllFunctionCalls("weather")
	if err != nil {
		fmt.Printf("Error handling response: %v\n", err)
		return
	}
	fmt.Println(item[0].Function.Arguments)

}

```

## For Another Version

```
go list -m -versions github.com/yuki5155/go-llms
```
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

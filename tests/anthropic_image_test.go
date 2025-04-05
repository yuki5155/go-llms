package tests

import (
	"bufio"
	"encoding/json"
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
	fmt.Println("========== テスト開始：画像入力のテスト ==========")

	// Load environment variables from .env file
	fmt.Println("環境変数を.envファイルから読み込みます")
	err := loadEnvForImage("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Skip test if env var is not set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping test")
	}
	fmt.Println("APIキーの取得に成功しました")

	// Create client
	fmt.Println("Anthropicクライアントを作成します")
	config := anthropicUtils.NewClientConfig(apiKey)
	// Set custom endpoint for debugging if needed
	endpoint := os.Getenv("ANTHROPIC_API_ENDPOINT")
	if endpoint != "" {
		fmt.Printf("カスタムエンドポイントを使用します: %s\n", endpoint)
		config.Endpoint = endpoint
	}
	client := anthropicUtils.NewClient(config)
	fmt.Printf("クライアント設定: エンドポイント=%s, バージョン=%s\n",
		config.Endpoint, config.AnthropicVersion)

	// Load test image
	fmt.Println("========== 画像ファイルの読み込み ==========")
	imagePath := filepath.Join("images", "31353427_s.jpg")
	absPath, _ := filepath.Abs(imagePath)
	fmt.Printf("画像パス: %s\n", imagePath)
	fmt.Printf("絶対パス: %s\n", absPath)

	// Check if file exists and get file info
	fileInfo, err := os.Stat(imagePath)
	if os.IsNotExist(err) {
		t.Fatalf("画像ファイルが存在しません: %s", imagePath)
	}
	if err != nil {
		t.Fatalf("ファイル情報の取得に失敗: %v", err)
	}
	fmt.Printf("ファイル情報: サイズ=%d バイト, 更新日時=%s\n",
		fileInfo.Size(), fileInfo.ModTime().String())

	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		t.Fatalf("画像ファイルの読み込みエラー: %v", err)
	}

	// Verify image data and display detailed debug info
	fmt.Printf("画像ファイルから %d バイトを読み込みました\n", len(imageBytes))
	if len(imageBytes) == 0 {
		t.Fatalf("画像ファイルが空です")
	}

	// Check image header to confirm file type
	fmt.Print("画像ヘッダーの確認: ")
	if len(imageBytes) > 2 {
		// JPEG signature check
		if imageBytes[0] == 0xFF && imageBytes[1] == 0xD8 && imageBytes[2] == 0xFF {
			fmt.Println("JPEGフォーマットが確認できました")
		} else {
			fmt.Printf("警告: JPEGではない可能性があります。最初の3バイト: %X %X %X\n",
				imageBytes[0], imageBytes[1], imageBytes[2])
		}
	}

	// Display first few bytes for debugging
	bytesToShow := 16
	if len(imageBytes) < bytesToShow {
		bytesToShow = len(imageBytes)
	}
	fmt.Printf("画像データの最初の%dバイト: ", bytesToShow)
	for i := 0; i < bytesToShow; i++ {
		fmt.Printf("%02X ", imageBytes[i])
	}
	fmt.Println()

	fmt.Println("========== メッセージの作成 ==========")
	// Create message with image
	fmt.Println("画像データをBase64に変換してメッセージを作成します")
	message := anthropicUtils.NewMessageWithImageBase64(
		imageBytes,
		"Please describe what you see in this image.",
	)

	// Debug: Print the message structure
	messageJSON, _ := json.MarshalIndent(message, "", "  ")
	truncateAt := 500
	messageJSONStr := string(messageJSON)
	if len(messageJSONStr) > truncateAt {
		// Truncate message but show structure
		fmt.Printf("メッセージ構造 (先頭%d文字): %s...\n",
			truncateAt, messageJSONStr[:truncateAt])
	} else {
		fmt.Printf("メッセージ構造: %s\n", messageJSONStr)
	}

	// Count messages and content items
	fmt.Printf("メッセージのコンテンツアイテム数: %d\n", len(message.Content))
	for i, content := range message.Content {
		fmt.Printf("コンテンツ #%d: タイプ=%s\n", i+1, content.Type)
		if content.Type == "image" && content.Source != nil {
			// For image content, show details but truncate the data
			fmt.Printf("  画像ソースタイプ: %s\n", content.Source.Type)
			fmt.Printf("  メディアタイプ: %s\n", content.Source.MediaType)
			dataLen := len(content.Source.Data)
			fmt.Printf("  データ長: %d バイト\n", dataLen)
			if dataLen > 0 {
				prefix := content.Source.Data
				if len(prefix) > 30 {
					prefix = prefix[:30]
				}
				fmt.Printf("  データプレフィックス: %s...\n", prefix)
			}
		} else if content.Type == "text" {
			fmt.Printf("  テキスト: %s\n", content.Text)
		}
	}

	// Create system message
	systemMessage := "You are a helpful assistant that describes images accurately."
	fmt.Printf("システムメッセージ: %s\n", systemMessage)

	// Set max tokens (required parameter for Anthropic API)
	maxTokens := 1024
	fmt.Printf("最大トークン数: %d\n", maxTokens)

	// Create request options
	fmt.Println("========== リクエストの送信 ==========")
	fmt.Println("リクエストオプションを作成します")
	opts := anthropicUtils.RequestOptions{
		Messages:  []anthropicUtils.Message{message},
		System:    systemMessage,
		MaxTokens: &maxTokens,
	}

	// Display model that will be used
	model := os.Getenv("ANTHROPIC_MODEL")
	if model != "" {
		opts.Model = model
		fmt.Printf("使用するモデル: %s (環境変数から)\n", model)
	} else {
		fmt.Println("使用するモデル: デフォルト (Claude 3 Opus)")
	}

	// Send request to Anthropic API
	fmt.Println("Anthropic APIにリクエストを送信します")
	resp, err := client.SendRequestWithStructuredOutput(opts)
	if err != nil {
		t.Fatalf("リクエスト送信エラー: %v", err)
	}

	// Check that we got a response
	if resp == nil {
		t.Fatalf("レスポンスを受信できませんでした")
	}

	fmt.Println("========== レスポンスの解析 ==========")
	fmt.Printf("レスポンスID: %s\n", resp.ID)
	fmt.Printf("レスポンスモデル: %s\n", resp.Model)
	fmt.Printf("レスポンスタイプ: %s\n", resp.Type)

	// Debug additional response details
	fmt.Println("--- レスポンス詳細デバッグ情報 ---")
	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Printf("完全なレスポンス構造体: %s\n", string(respJSON))

	// Check message structure if available
	fmt.Printf("Message構造体: ID=%s, Type=%s, Role=%s\n",
		resp.Message.ID, resp.Message.Type, resp.Message.Role)

	// Check if usage statistics are available
	fmt.Printf("使用トークン: 入力=%d, 出力=%d\n",
		resp.Message.Usage.InputTokens, resp.Message.Usage.OutputTokens)

	// Check stop reason and sequence if available
	fmt.Printf("停止理由: %s\n", resp.Message.StopReason)
	if resp.Message.StopSequence != nil {
		fmt.Printf("停止シーケンス: %s\n", *resp.Message.StopSequence)
	}

	// Check message content
	contentCount := len(resp.Message.Content)
	fmt.Printf("レスポンスのコンテンツブロック数: %d\n", contentCount)

	// 問題の詳細診断: Content配列が空の場合
	if contentCount == 0 {
		fmt.Println("警告: Content配列が空です - APIの応答構造を確認します")

		// Check if we can access the raw response
		rawRespBytes, err := json.Marshal(resp)
		if err == nil {
			var rawResp map[string]interface{}
			if err := json.Unmarshal(rawRespBytes, &rawResp); err == nil {
				fmt.Println("APIレスポンスの最上位キー:")
				for k := range rawResp {
					fmt.Printf("- %s\n", k)
				}

				// Inspect message field specifically
				if msgRaw, ok := rawResp["message"].(map[string]interface{}); ok {
					fmt.Println("'message'フィールドのキー:")
					for k, v := range msgRaw {
						fmt.Printf("- %s (型: %T)\n", k, v)
					}

					// Specifically check for content field structure
					if contentRaw, ok := msgRaw["content"]; ok {
						fmt.Printf("'content'フィールドの型: %T\n", contentRaw)

						// If it's an array, check its elements
						if contentArr, ok := contentRaw.([]interface{}); ok {
							fmt.Printf("'content'配列の長さ: %d\n", len(contentArr))

							// Inspect first item if available
							if len(contentArr) > 0 {
								fmt.Printf("最初のコンテンツアイテム: %+v\n", contentArr[0])

								// If it's a map, check its keys
								if firstItem, ok := contentArr[0].(map[string]interface{}); ok {
									fmt.Println("最初のコンテンツアイテムのキー:")
									for k, v := range firstItem {
										fmt.Printf("- %s: %+v (型: %T)\n", k, v, v)
									}
								}
							}
						}
					} else {
						fmt.Println("警告: 'content'フィールドが見つかりません")
					}
				}
			}
		}

		// Move to the fatal error after diagnosis
		t.Fatalf("レスポンスにコンテンツがありません")
	}

	// Display all content blocks
	fmt.Println("すべてのコンテンツブロック:")
	for i, block := range resp.Message.Content {
		fmt.Printf("ブロック #%d: タイプ=%s\n", i+1, block.Type)
		if block.Type == "text" {
			fmt.Printf("テキスト内容:\n%s\n", block.Text)
		} else {
			fmt.Printf("テキスト以外のブロック: %s\n", block.Type)
		}
	}

	// Display text from the first content block
	var responseText string
	for _, block := range resp.Message.Content {
		if block.Type == "text" {
			responseText = block.Text
			break
		}
	}

	fmt.Println("========== 画像の説明 ==========")
	fmt.Println(responseText)
	fmt.Println("========== テスト終了 ==========")
}

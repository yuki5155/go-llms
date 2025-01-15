package utils_test

import (
	"encoding/json"
	"testing"

	"github.com/yuki5155/go-llms/openai-llm/utils"
)

func TestContext(t *testing.T) {
	content := utils.Content{
		Text: "",
		Type: "text",
	}

	// 通常のJSON出力
	jsonBytes, err := json.Marshal(content)
	if err != nil {
		t.Errorf("Failed to marshal JSON: %v", err)
	}

	// jsonの出力をアサート
	if string(jsonBytes) != `{"type":"text"}` {
		t.Errorf("JSON output is incorrect")
	}
}

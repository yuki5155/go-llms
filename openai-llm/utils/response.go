package utils

import (
	"encoding/json"
	"fmt"
)

type APIResponse struct {
	Choices []Choice `json:"choices"`
}

func ParseStructuredResponse[T any](content json.RawMessage) (*T, error) {
	// 最初のJSONアンマーシャル：文字列として取得
	var jsonString string
	if err := json.Unmarshal(content, &jsonString); err != nil {
		return nil, fmt.Errorf("error unmarshaling outer JSON: %v", err)
	}

	// デバッグ出力（必要に応じて有効化）
	// fmt.Printf("Parsed JSON string: %s\n", jsonString)

	// 文字列から目的の型へのアンマーシャル
	var result T
	if err := json.Unmarshal([]byte(jsonString), &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling inner JSON: %v", err)
	}

	return &result, nil
}

// カスタムエラーにディレクトリに移行する
type ResponseError struct {
	Type    string
	Message string
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// [TODO]NewResponseError は新しいResponseErrorを作成します
func NewResponseError(errorType, message string) *ResponseError {
	return &ResponseError{
		Type:    errorType,
		Message: message,
	}
}

func HandleResponse[T any](resp *APIResponse) (*T, error) {
	if resp == nil {
		return nil, NewResponseError("NullResponse", "response is nil")
	}

	if len(resp.Choices) == 0 {
		return nil, NewResponseError("NoChoices", "no choices in the API response")
	}

	choice := resp.Choices[0]
	switch choice.FinishReason {
	case "stop":
		if choice.Message.Refusal != nil {
			return nil, NewResponseError("ModelRefusal", *choice.Message.Refusal)
		}

		// レスポンスが空でないことを確認
		if len(choice.Message.Content) == 0 {
			return nil, NewResponseError("EmptyContent", "response content is empty")
		}

		// 構造化レスポンスのパース
		result, err := ParseStructuredResponse[T](choice.Message.Content)
		if err != nil {
			return nil, NewResponseError("ParseError", fmt.Sprintf("error parsing response: %v", err))
		}

		return result, nil

	case "length":
		return nil, NewResponseError("TokenLimit", "the response was truncated due to token limit")
	case "content_filter":
		return nil, NewResponseError("ContentFilter", "the response was filtered due to content restrictions")
	default:
		return nil, NewResponseError("UnexpectedFinishReason", fmt.Sprintf("unexpected finish reason: %s", choice.FinishReason))
	}
}

func ResponseErrorIs(err error, errorType string) bool {
	if respErr, ok := err.(*ResponseError); ok {
		return respErr.Type == errorType
	}
	return false
}

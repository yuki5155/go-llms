package schema

import (
	"encoding/json"
)

// BaseSchema は全てのスキーマに共通する基本構造を定義します
type BaseSchema struct {
	Type                 string                    `json:"type"`
	Properties           map[string]SchemaProperty `json:"properties"`
	Required             []string                  `json:"required"`
	AdditionalProperties *bool                     `json:"additionalProperties"`
}

// SchemaProperty はJSONスキーマのプロパティを表現します
type SchemaProperty struct {
	Type        string                    `json:"type"`
	Description string                    `json:"description,omitempty"`
	Enum        []string                  `json:"enum,omitempty"`
	Items       *SchemaProperty           `json:"items,omitempty"`
	Required    []string                  `json:"required,omitempty"`
	Properties  map[string]SchemaProperty `json:"properties,omitempty"`
}

// RequestFormat はOpenAIへのリクエストのフォーマットを定義します
type RequestFormat struct {
	Type       string          `json:"type"`
	JSONSchema json.RawMessage `json:"json_schema"`
}

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
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// RequestFormat はOpenAIへのリクエストのフォーマットを定義します
type RequestFormat struct {
	Type       string          `json:"type"`
	JSONSchema json.RawMessage `json:"json_schema"`
}

// WeatherSchema は天気情報のスキーマを定義します
type WeatherSchema struct {
	Name   string     `json:"name"`
	Schema BaseSchema `json:"schema"`
}

// WeatherResponse は天気情報のレスポンスを定義します
type WeatherResponse struct {
	Location    string  `json:"location"`
	Temperature float64 `json:"temperature"`
	Unit        string  `json:"unit"`
	Conditions  string  `json:"conditions"`
}

// NewWeatherSchema は新しいWeatherSchemaを作成します
func NewWeatherSchema() *WeatherSchema {
	falseValue := false
	return &WeatherSchema{
		Name: "weather_response",
		Schema: BaseSchema{
			Type: "object",
			Properties: map[string]SchemaProperty{
				"location": {
					Type:        "string",
					Description: "Location for weather information",
				},
				"temperature": {
					Type:        "number",
					Description: "Current temperature",
				},
				"unit": {
					Type:        "string",
					Description: "Temperature unit",
					Enum:        []string{"C", "F"},
				},
				"conditions": {
					Type:        "string",
					Description: "Current weather conditions",
				},
			},
			Required:             []string{"location", "temperature", "unit", "conditions"},
			AdditionalProperties: &falseValue,
		},
	}
}

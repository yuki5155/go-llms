package schema

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

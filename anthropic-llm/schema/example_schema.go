package schema

// WeatherResponse represents a structured weather information response
type WeatherResponse struct {
	Location         string  `json:"location"`
	Temperature      float64 `json:"temperature"`
	Unit             string  `json:"unit"`
	Forecast         string  `json:"forecast"`
	Humidity         int     `json:"humidity"`
	WindSpeed        float64 `json:"wind_speed"`
	WindDirection    string  `json:"wind_direction"`
	PrecipitationPct int     `json:"precipitation_pct"`
}

// NewWeatherSchema creates a schema for weather information
func NewWeatherSchema() *RequestFormat {
	return &RequestFormat{
		Type: "json_object",
	}
}

// NewWeatherToolSchema creates a tool schema for weather function calling
func NewWeatherToolSchema() *Tool {
	falseValue := false

	return &Tool{
		Type: "function",
		Function: Function{
			Name:        "get_weather",
			Description: "Get the current weather for a location",
			Parameters: BaseSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"location": {
						Type:        "string",
						Description: "The location to get weather for (e.g., city name, postal code)",
					},
					"unit": {
						Type:        "string",
						Description: "The unit of measurement for temperature",
						Enum:        []string{"celsius", "fahrenheit"},
					},
				},
				Required:             []string{"location"},
				AdditionalProperties: &falseValue,
			},
		},
	}
}

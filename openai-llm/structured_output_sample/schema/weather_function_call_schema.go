package schema

func NewWeatherFunctionCallSchema() *Tool {
	falseValue := false
	return &Tool{
		Type: "function",
		Function: Function{
			Name:        "weather",
			Description: "Get weather information",
			Parameters: BaseSchema{
				Type: "object",
				Properties: map[string]SchemaProperty{
					"location": {
						Type:        "string",
						Description: "Location for weather information",
					},
				},
				Required:             []string{"location"},
				AdditionalProperties: &falseValue,
			},
			Strict: false,
		},
	}
}

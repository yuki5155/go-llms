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
func NewWeatherSchema() *BaseSchema {
	falseValue := false

	return &BaseSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"location": {
				Type:        "string",
				Description: "The location name (city)",
			},
			"temperature": {
				Type:        "number",
				Description: "The current temperature value",
			},
			"unit": {
				Type:        "string",
				Description: "The temperature unit (C or F)",
				Enum:        []string{"C", "F"},
			},
			"forecast": {
				Type:        "string",
				Description: "Weather forecast description",
			},
			"humidity": {
				Type:        "integer",
				Description: "Humidity percentage (0-100)",
			},
			"wind_speed": {
				Type:        "number",
				Description: "Wind speed value",
			},
			"wind_direction": {
				Type:        "string",
				Description: "Wind direction (N, NE, E, SE, S, SW, W, NW)",
			},
			"precipitation_pct": {
				Type:        "integer",
				Description: "Precipitation percentage (0-100)",
			},
		},
		Required: []string{
			"location",
			"temperature",
			"unit",
			"forecast",
			"humidity",
			"wind_speed",
			"wind_direction",
			"precipitation_pct",
		},
		AdditionalProperties: &falseValue,
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

// ProductInfo represents a structured product information response
type ProductInfo struct {
	ProductID   string  `json:"product_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Category    string  `json:"category"`
	Brand       string  `json:"brand"`
	InStock     bool    `json:"in_stock"`
	Quantity    int     `json:"quantity"`
	Rating      float64 `json:"rating"`
	ReviewCount int     `json:"review_count"`
	ReleaseDate string  `json:"release_date"`
}

// NewProductSchema creates a schema for product information
func NewProductSchema() *BaseSchema {
	falseValue := false

	return &BaseSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"product_id": {
				Type:        "string",
				Description: "Unique identifier for the product",
			},
			"name": {
				Type:        "string",
				Description: "Product name",
			},
			"description": {
				Type:        "string",
				Description: "Detailed product description",
			},
			"price": {
				Type:        "number",
				Description: "Current price of the product",
			},
			"currency": {
				Type:        "string",
				Description: "Currency code (USD, EUR, JPY, etc.)",
			},
			"category": {
				Type:        "string",
				Description: "Product category",
			},
			"brand": {
				Type:        "string",
				Description: "Brand name",
			},
			"in_stock": {
				Type:        "boolean",
				Description: "Whether the product is in stock",
			},
			"quantity": {
				Type:        "integer",
				Description: "Available quantity",
			},
			"rating": {
				Type:        "number",
				Description: "Average customer rating (0-5)",
			},
			"review_count": {
				Type:        "integer",
				Description: "Number of customer reviews",
			},
			"release_date": {
				Type:        "string",
				Description: "Product release date in ISO format",
			},
		},
		Required: []string{
			"product_id",
			"name",
			"price",
			"currency",
			"category",
			"in_stock",
		},
		AdditionalProperties: &falseValue,
	}
}

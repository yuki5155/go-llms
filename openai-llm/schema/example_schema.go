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

// ImageAnalysisSchema は画像分析のスキーマを定義します
type ImageAnalysisSchema struct {
	Name   string     `json:"name"`
	Schema BaseSchema `json:"schema"`
}

// ImageAnalysisResponse は画像分析のレスポンスを定義します
type ImageAnalysisResponse struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Objects     string `json:"objects"`
}

const ImageAnalysisPrompt = `Please analyze the image and provide information in the following format:
- Identify the overall category of the scene (landscape, cityscape, indoor, etc.)
- Provide a detailed description of what you see in the image
- List all significant objects present in the image in a comma-separated format

Rules and guidelines:
1. For category: Use a single, concise term that best describes the scene type
2. For description: Write a clear and detailed paragraph about the image content, including notable features and composition
3. For objects: List only distinct, clearly visible objects, separated by commas without spaces

Example response format:
{
    "category": "landscape",
    "description": "A serene mountain landscape captured during sunset, featuring snow-capped peaks reflected in a calm alpine lake. The foreground shows scattered pine trees and rocky terrain.",
    "objects": "mountain,lake,snow,trees,rocks,sky"
}

Please ensure all observations are objective and focus on visible elements in the image.`

func NewImageAnalysisSchema() *ImageAnalysisSchema {
	falseValue := false
	return &ImageAnalysisSchema{
		Name: "image_analysis_response",
		Schema: BaseSchema{
			Type: "object",
			Properties: map[string]SchemaProperty{
				"category": {
					Type:        "string",
					Description: "Category of the image scene (e.g., landscape, cityscape, indoor)",
				},
				"description": {
					Type:        "string",
					Description: "Detailed description of the image content",
				},
				"objects": {
					Type:        "string",
					Description: "Comma-separated list of detected objects in the image (e.g., tree,mountain,sky)",
				},
			},
			Required:             []string{"category", "description", "objects"},
			AdditionalProperties: &falseValue,
		},
	}
}

// ----------------------------------------------------------------------------

// ObjectAnalysisSchema はオブジェクト分析のスキーマを定義します
type ObjectAnalysisSchema struct {
	Name   string     `json:"name"`
	Schema BaseSchema `json:"schema"`
}

// ObjectAnalysisResponse はオブジェクト分析のレスポンスを定義します
type ObjectAnalysisResponse struct {
	Objects []ObjectItem `json:"objects"`
}

// ObjectItem は分析された個別のオブジェクトを表す構造体です
type ObjectItem struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

const ObjectAnalysisPrompt = `Please analyze the objects in the image and provide information in the following format:
{
    "objects": [
        {
            "name": "",
            "category": ""
        }
    ]
}

Rules and guidelines:
1. For name: Provide a descriptive name for the identified object
2. For category: Use a single, concise term that best categorizes the object

Example response format:
{
    "objects": [
        {
            "name": "Red Chair",
            "category": "furniture"
        },
        {
            "name": "Coffee Cup",
            "category": "kitchenware"
        }
    ]
}

Please identify and list all distinct objects visible in the image.`

func NewObjectAnalysisSchema() *ObjectAnalysisSchema {
	falseValue := false
	return &ObjectAnalysisSchema{
		Name: "object_analysis_response",
		Schema: BaseSchema{
			Type: "object",
			Properties: map[string]SchemaProperty{
				"objects": {
					Type:        "array",
					Description: "List of objects detected in the image",
					Items: &SchemaProperty{
						Type: "object",
						Properties: map[string]SchemaProperty{
							"name": {
								Type:        "string",
								Description: "Name of the identified object",
							},
							"category": {
								Type:        "string",
								Description: "Category of the identified object",
							},
						},
						Required: []string{"name", "category"},
					},
				},
			},
			Required:             []string{"objects"},
			AdditionalProperties: &falseValue,
		},
	}
}

# Structured Output with Anthropic Claude

This document explains how to use the structured output feature of the `anthropic-llm` package to get JSON responses from Claude models.

## Overview

The structured output feature allows you to get responses from Claude in a specific JSON format defined by a JSON schema. This is useful when you need to parse and process the LLM's responses programmatically.

## Basic Usage

### Simple Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/yuki5155/go-llms/anthropic-llm/utils"
)

func main() {
	// Initialize the client
	config := utils.NewClientConfig("your-api-key")
	client := utils.NewClient(config)

	// Define a schema for the response
	schema := []byte(`{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"description": "The name of the person"
			},
			"age": {
				"type": "integer",
				"description": "The age of the person"
			},
			"hobbies": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"description": "List of the person's hobbies"
			}
		},
		"required": ["name", "age"]
	}`)

	// Create a message to send to Claude
	message := utils.NewTextMessage(utils.RoleUser, "Tell me about a fictional person named Alex.")

	// Set up the request options
	options := utils.RequestOptions{
		Messages: []utils.Message{message},
		Schema:   schema,
		System:   "You are a helpful assistant that provides information about fictional people.",
	}

	// Send the request
	response, err := client.SendRequestWithStructuredOutput(options)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Process the response into a typed struct (recommended approach)
	type Person struct {
		Name    string   `json:"name"`
		Age     int      `json:"age"`
		Hobbies []string `json:"hobbies,omitempty"`
	}

	person, err := utils.HandleResponse[Person](response)
	if err != nil {
		log.Fatalf("Error processing response: %v", err)
	}

	// Use the structured data
	fmt.Printf("Name: %s\n", person.Name)
	fmt.Printf("Age: %d\n", person.Age)
	fmt.Printf("Hobbies: %v\n", person.Hobbies)
}
```

### Image Input Example

Claude can also analyze images and provide structured output based on their content:

```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/yuki5155/go-llms/anthropic-llm/utils"
)

func main() {
	// Initialize the client
	config := utils.NewClientConfig("your-api-key")
	client := utils.NewClient(config)

	// Define a schema for the response
	schema := []byte(`{
		"type": "object",
		"properties": {
			"objects": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"description": "List of objects identified in the image"
			},
			"scene_description": {
				"type": "string",
				"description": "A brief description of the scene in the image"
			},
			"main_colors": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"description": "The main colors present in the image"
			}
		},
		"required": ["objects", "scene_description"]
	}`)

	// Read image file
	imageBytes, err := ioutil.ReadFile("path/to/your/image.jpg")
	if err != nil {
		log.Fatalf("Error reading image: %v", err)
	}

	// Create a message with the image
	message := utils.NewMessageWithImageBase64(imageBytes, "What can you see in this image? Provide structured information.")

	// Set up the request options
	options := utils.RequestOptions{
		Messages:  []utils.Message{message},
		Schema:    schema,
		System:    "You are a helpful assistant that analyzes images.",
		MaxTokens: intPtr(1000), // Use helper function to create a pointer to int
	}

	// Send the request
	response, err := client.SendRequestWithStructuredOutput(options)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Process the response into a typed struct
	type ImageAnalysis struct {
		Objects         []string `json:"objects"`
		SceneDescription string   `json:"scene_description"`
		MainColors      []string `json:"main_colors,omitempty"`
	}
	
	analysis, err := utils.HandleResponse[ImageAnalysis](response)
	if err != nil {
		log.Fatalf("Error processing response: %v", err)
	}
	
	// Use the structured data
	fmt.Printf("Scene description: %s\n", analysis.SceneDescription)
	fmt.Printf("Objects identified: %v\n", analysis.Objects)
	fmt.Printf("Main colors: %v\n", analysis.MainColors)
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
```

You can also use URLs for images:

```go
// Create a message with an image URL
imageURL := "https://example.com/image.jpg"
message := utils.NewMessageWithImage(imageURL, "What can you see in this image? Provide structured information.")
```

## Response Handling

### HandleResponse - Type-Safe Response Processing

The `HandleResponse` function is a generic function for safely converting API responses to any specified type:

```go
// Get typed data directly from API response
person, err := utils.HandleResponse[Person](response)
if err != nil {
	log.Fatalf("Error processing response: %v", err)
}

// You can also handle errors more specifically
person, err := utils.HandleResponse[Person](response)
if err != nil {
	if utils.ResponseErrorIs(err, "ParseError") {
		// Special handling for JSON parsing errors
		log.Printf("JSON parsing error: %v", err)
	} else if utils.ResponseErrorIs(err, "NullResponse") {
		// Handle case when response is null
		log.Printf("Response is null")
	} else {
		// Other errors
		log.Fatalf("Error processing response: %v", err)
	}
}
```

### ParseStructuredResponse - Manual Response Processing

For more fine-grained control in specific needs, you can also use the lower-level `ParseStructuredResponse` function:

```go
// Manually extract JSON from ContentBlock
if len(response.Message.Content) > 0 {
	person, err := utils.ParseStructuredResponse[Person](response.Message.Content)
	if err != nil {
		log.Fatalf("Parsing error: %v", err)
	}
	fmt.Printf("Person: %+v\n", person)
}
```

### Error Handling

You can use the `ResponseError` type and `ResponseErrorIs` function for more detailed error handling:

```go
person, err := utils.HandleResponse[Person](response)
if err != nil {
	switch {
	case utils.ResponseErrorIs(err, "ParseError"):
		log.Printf("JSON parsing error: %v", err)
		// Fix JSON or handle differently
	case utils.ResponseErrorIs(err, "NullResponse"):
		log.Printf("Empty response")
		// Set default values, etc.
	default:
		log.Fatalf("Unexpected error: %v", err)
	}
}
```

## Advanced Usage

### Custom Response Types

You can define your own struct types to unmarshal JSON directly:

```go
// Define custom struct
type WeatherForecast struct {
	Location       string  `json:"location"`
	Temperature    float64 `json:"temperature"`
	Unit           string  `json:"unit"`
	Conditions     string  `json:"conditions"`
	Humidity       int     `json:"humidity"`
	WindSpeed      float64 `json:"wind_speed"`
	WindDirection  string  `json:"wind_direction"`
	PrecipitationPct int   `json:"precipitation_pct"`
	Forecast       string  `json:"forecast"`
}

// Process response
weather, err := utils.HandleResponse[WeatherForecast](response)
if err != nil {
	log.Fatalf("Error: %v", err)
}

fmt.Printf("Weather in %s:\n", weather.Location)
fmt.Printf("Temperature: %.1f %s\n", weather.Temperature, weather.Unit)
fmt.Printf("Conditions: %s\n", weather.Conditions)
fmt.Printf("Humidity: %d%%\n", weather.Humidity)
```

### Setting Temperature

To control the randomness of the response, set the temperature:

```go
temp := 0.7
options := utils.RequestOptions{
	Messages:    []utils.Message{message},
	Schema:      schema,
	System:      "You are a helpful assistant.",
	Temperature: &temp,
}
```

### Setting Max Tokens

To limit the response size, set the maximum number of tokens:

```go
maxTokens := 1000
options := utils.RequestOptions{
	Messages:  []utils.Message{message},
	Schema:    schema,
	System:    "You are a helpful assistant.",
	MaxTokens: &maxTokens,
}
```

## Schema Requirements

When defining your JSON schema:

1. It must be a valid JSON schema object
2. Properties should have appropriate types (string, number, integer, boolean, array, object)
3. You can specify required fields with the `required` array
4. Include descriptions for better results

## How It Works

The `SendRequestWithStructuredOutput` function works as follows:

1. Takes a JSON schema and message(s) to send to Claude
2. Generates instructions for Claude based on the schema
3. Sets system instructions to enforce JSON format
4. Handles recovery in case of parsing issues

The `HandleResponse` function handles:

1. API response validation
2. JSON text extraction from ContentBlock (including handling markdown code blocks)
3. Converting extracted JSON to the specified type
4. Providing detailed error information in case of failures

## Error Handling

This library handles the following types of errors:

- Missing messages or schema
- API request failures
- Non-200 responses from the API
- Response parsing errors
- Invalid JSON formats

Always check the returned error in your application and handle it appropriately.

## Best Practices

1. Define specific and detailed schemas
2. Include clear descriptions for each property
3. Use appropriate types for each property
4. Make required fields explicit
5. Keep schemas as simple as possible
6. Test your schemas with various prompts
7. Implement proper error handling
8. For complex schemas, try incrementally and adjust

## Limitations

1. Claude may occasionally produce invalid JSON, especially with complex schemas
2. Very large or deeply nested schemas may not work reliably
3. Schema validation happens on the client side, not in the API
4. JSON may sometimes be returned in special formats like markdown 
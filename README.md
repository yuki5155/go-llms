# go-llms

A Go library for interacting with Large Language Models (LLMs), primarily focused on OpenAI's GPT models and Anthropic's Claude models with support for structured output and function calling.

## Features

- Structured JSON output from LLM responses
- Function calling capabilities 
- Support for images and multimodal inputs
- Customizable schema definitions
- Type-safe response handling
- Support for both OpenAI and Anthropic APIs

## Installation

```bash
go get github.com/yuki5155/go-llms@v1.0.0
```

## Usage Examples

### OpenAI Structured Output

Retrieve structured JSON data from OpenAI's GPT models:

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yuki5155/go-llms/openai-llm/schema"
	"github.com/yuki5155/go-llms/openai-llm/utils"
)

func main() {
	// Set up API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the OPENAI_API_KEY environment variable.")
		return
	}

	// Create client
	config := utils.NewClientConfig(apiKey)
	client := utils.NewClient(config)

	// Create a schema for the response
	weatherSchema := schema.NewWeatherSchema()
	schemaJSON, err := json.Marshal(weatherSchema)
	if err != nil {
		fmt.Printf("Error marshalling weather schema: %v\n", err)
		return
	}

	// Prepare messages
	messages := []utils.Message{
		utils.NewMessage(utils.RoleSystem, "You are a helpful assistant designed to output weather information in JSON format."),
		utils.NewMessage(utils.RoleUser, "What's the weather like in Tokyo today?"),
	}

	// Create request options
	opts := utils.RequestOptions{
		Messages: messages,
		Schema:   schemaJSON,
	}

	// Send request to OpenAI API
	resp, err := client.SendRequestWithStructuredOutput(opts)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}

	// Process the response
	weather, err := utils.HandleResponse[schema.WeatherResponse](resp)
	if err != nil {
		fmt.Printf("Error handling response: %v\n", err)
		return
	}

	// Access specific data
	fmt.Printf("Temperature: %v\n", weather.Temperature)

	// Display all results
	prettyJSON, err := json.MarshalIndent(weather, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting response: %v\n", err)
		return
	}
	fmt.Printf("Weather Information:\n%s\n", string(prettyJSON))
}
```

### Anthropic Structured Output

Retrieve structured JSON data from Anthropic's Claude models:

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yuki5155/go-llms/anthropic-llm/schema"
	"github.com/yuki5155/go-llms/anthropic-llm/utils"
)

func main() {
	// Set up API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the ANTHROPIC_API_KEY environment variable.")
		return
	}

	// Create client
	config := utils.NewClientConfig(apiKey)
	client := utils.NewClient(config)

	// Create a schema for the response
	weatherSchema := schema.NewWeatherSchema()
	schemaJSON, err := json.Marshal(weatherSchema)
	if err != nil {
		fmt.Printf("Error marshalling weather schema: %v\n", err)
		return
	}

	// Create system message and user message
	systemMessage := "You are a helpful assistant designed to output weather information in JSON format."
	messages := []utils.Message{
		utils.NewTextMessage(utils.RoleUser, "What's the weather like in Tokyo today?"),
	}

	// Create request options
	opts := utils.RequestOptions{
		Messages: messages,
		System:   systemMessage,
		Schema:   schemaJSON,
	}

	// Send request to Anthropic API
	resp, err := client.SendRequestWithStructuredOutput(opts)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}

	// Process the response
	weather, err := utils.HandleResponse[schema.WeatherResponse](resp)
	if err != nil {
		fmt.Printf("Error handling response: %v\n", err)
		return
	}

	// Access specific data
	fmt.Printf("Temperature: %v\n", weather.Temperature)

	// Display all results
	prettyJSON, err := json.MarshalIndent(weather, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting response: %v\n", err)
		return
	}
	fmt.Printf("Weather Information:\n%s\n", string(prettyJSON))
}
```

### Anthropic Tool Usage

Implement Anthropic function calling for tool use:

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yuki5155/go-llms/anthropic-llm/schema"
	"github.com/yuki5155/go-llms/anthropic-llm/utils"
)

func main() {
	// Set up API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the ANTHROPIC_API_KEY environment variable.")
		return
	}

	// Create client
	config := utils.NewClientConfig(apiKey)
	client := utils.NewClient(config)

	// Create tool schema
	weatherTool := schema.NewWeatherToolSchema()
	tools := []schema.Tool{*weatherTool}
	toolsJSON, err := json.Marshal(tools)
	if err != nil {
		fmt.Printf("Error marshalling tools: %v\n", err)
		return
	}

	// Create system message and user message
	systemMessage := "You are a helpful assistant that can provide weather information."
	messages := []utils.Message{
		utils.NewTextMessage(utils.RoleUser, "What's the weather like in Kyoto today?"),
	}

	// Create request options
	opts := utils.RequestOptions{
		Messages: messages,
		System:   systemMessage,
		Schema:   toolsJSON,
	}

	// Send tool call request
	resp, err := client.SendRequestWithTool(opts)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}

	// Get tool use
	toolUse, err := resp.GetToolUse("get_weather")
	if err != nil {
		fmt.Printf("Error getting tool use: %v\n", err)
		return
	}

	// Parse arguments
	var args map[string]interface{}
	err = json.Unmarshal([]byte(toolUse.Arguments), &args)
	if err != nil {
		fmt.Printf("Error parsing arguments: %v\n", err)
		return
	}

	fmt.Printf("Tool Name: %s\n", toolUse.Name)
	fmt.Printf("Location: %s\n", args["location"])
}
```

## Project Structure

- `openai-llm/`
  - `schema/`: Data structures and JSON schemas for OpenAI
  - `utils/`: Client utilities and helper functions for OpenAI
- `anthropic-llm/`
  - `schema/`: Data structures and JSON schemas for Anthropic
  - `utils/`: Client utilities and helper functions for Anthropic

## Implementation Status

### OpenAI LLM
- ✅ Structured JSON output
- ✅ Function calling
- ✅ Image input support
- ✅ Testing

### Anthropic LLM
- ✅ Basic structured output (JSON)
- 🚧 Function/tool calling (not fully implemented)
- 🚧 Image input support (implementation ready, but tests require image files)
- 🔄 Testing (partial)

**Note**: The Anthropic API implementation is currently in progress. While basic structured output is working, some features like function calling and image input need additional work.

## Available Schemas

The library includes several pre-built schemas:

- `WeatherSchema`: For retrieving weather information
- `ImageAnalysisSchema`: For analyzing image content
- `ObjectAnalysisSchema`: For identifying objects in images

## Custom Schemas

You can create custom schemas by implementing the appropriate interfaces and structures:

```go
func NewCustomSchema() *CustomSchema {
    falseValue := false
    return &CustomSchema{
        Name: "custom_response",
        Schema: BaseSchema{
            Type: "object",
            Properties: map[string]SchemaProperty{
                "field1": {
                    Type:        "string",
                    Description: "Description of field1",
                },
                "field2": {
                    Type:        "number",
                    Description: "Description of field2",
                },
            },
            Required:             []string{"field1", "field2"},
            AdditionalProperties: &falseValue,
        },
    }
}
```

## Version Information

Check available versions:

```bash
go list -m -versions github.com/yuki5155/go-llms
```

## License

See the LICENSE file for details.
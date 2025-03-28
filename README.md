# go-llms

A Go library for interacting with Large Language Models (LLMs), primarily focused on OpenAI's GPT models with support for structured output and function calling.

## Features

- Structured JSON output from LLM responses
- Function calling capabilities 
- Support for images and multimodal inputs
- Customizable schema definitions
- Type-safe response handling

## Installation

```bash
go get github.com/yuki5155/go-llms@v1.0.0
```

## Usage Examples

### Structured Output

Retrieve structured JSON data from an LLM:

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

### Function Calling

Implement OpenAI function calling for tool use:

```go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/yuki5155/go-llms/openai-llm/schema"
	"github.com/yuki5155/go-llms/openai-llm/utils"
)

func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

func main() {
	// Load environment variables
	err := loadEnv(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the OPENAI_API_KEY environment variable.")
		return
	}
	
	// Create client
	config := utils.NewClientConfig(apiKey)
	client := utils.NewClient(config)
	
	// Create function schema
	weatherSchema := schema.NewWeatherFunctionCallSchema()
	tools := []schema.Tool{*weatherSchema}
	toolsJSON, err := json.Marshal(tools)
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
		Schema:   toolsJSON,
	}
	
	// Send function call request
	res, err := client.SendRequestWithFunctionCall(opts)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}

	// Get function call results
	functionCalls, err := res.GetAllFunctionCalls("weather")
	if err != nil {
		fmt.Printf("Error handling response: %v\n", err)
		return
	}
	
	fmt.Println(functionCalls[0].Function.Arguments)
}
```

## Project Structure

- `openai-llm/`
  - `schema/`: Data structures and JSON schemas
  - `utils/`: Client utilities and helper functions

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
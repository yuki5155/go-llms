# Anthropic Claude Go Client Documentation

This directory contains documentation for the Go client for Anthropic's Claude API.

## Contents

- [Structured Output](structured_output.md) - How to use Claude to generate structured JSON responses
- [Tool Calling](tool_calling.md) - How to use Claude's function/tool calling capabilities

## Key Features

This client supports:

- Basic text conversations with Claude
- Multimodal inputs (text + images)
- Structured JSON outputs (based on JSON schema)
- Function/tool calling
- Custom system prompts
- Temperature control
- Token limit settings
- Type-safe response handling (HandleResponse)
- Advanced error handling (ResponseError)

## Getting Started

To use this client, you'll need an API key from Anthropic. Once you have that, you can initialize the client and start making requests.

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

	// Create a message
	message := utils.NewTextMessage(utils.RoleUser, "Hello, Claude! How are you today?")

	// Set up the request options
	options := utils.RequestOptions{
		Messages: []utils.Message{message},
		System:   "You are a helpful assistant.",
	}

	// Send the request
	response, err := client.SendRequestWithTool(options)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Process the response
	if len(response.Message.Content) > 0 {
		for _, content := range response.Message.Content {
			if content.Type == "text" {
				fmt.Println(content.Text)
			}
		}
	}
}
```

## Response Processing Functions

Helpful functions available for processing API responses:

### HandleResponse

Generic type-safe response processing function that converts API responses to a specified type:

```go
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

person, err := utils.HandleResponse[Person](response)
if err != nil {
    log.Fatalf("Error: %v", err)
}
fmt.Printf("Name: %s, Age: %d\n", person.Name, person.Age)
```

### ParseStructuredResponse

Lower-level parsing function for more fine-grained control:

```go
person, err := utils.ParseStructuredResponse[Person](response.Message.Content)
```

### ResponseErrorIs

Error type checking for more detailed error handling:

```go
if utils.ResponseErrorIs(err, "ParseError") {
    // Special handling for JSON parsing errors
}
```

### DebugResponse

Display response details for debugging purposes:

```go
utils.DebugResponse(response)
```

## Usage Examples

### Basic Text Conversation

```go
message := utils.NewTextMessage(utils.RoleUser, "What's the weather like in Tokyo?")
```

### Image-Based Requests

```go
// Using an image from URL
imageURL := "https://example.com/image.jpg" 
message := utils.NewMessageWithImage(imageURL, "Please describe this image")

// Or from a local file
imageBytes, _ := ioutil.ReadFile("path/to/image.jpg")
message := utils.NewMessageWithImageBase64(imageBytes, "Please describe this image")
```

### Structured Output

```go
schema := []byte(`{
    "type": "object",
    "properties": {
        "summary": {
            "type": "string",
            "description": "A summary of the text"
        },
        "sentiment": {
            "type": "string",
            "description": "The sentiment of the text (positive, negative, neutral)"
        }
    },
    "required": ["summary", "sentiment"]
}`)

// Set up request options
options := utils.RequestOptions{
    Messages: []utils.Message{message},
    Schema:   schema,
    System:   "You are an assistant that analyzes text.",
}

// Get structured response
response, err := client.SendRequestWithStructuredOutput(options)
```

## Additional Resources

- [Anthropic API Documentation](https://docs.anthropic.com/claude/reference/getting-started-with-the-api)
- [Claude 3 Models](https://www.anthropic.com/claude) 
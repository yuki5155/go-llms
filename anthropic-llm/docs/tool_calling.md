# Tool Calling with Anthropic Claude

This document explains how to use the function/tool calling feature of the `anthropic-llm` package, which allows Claude to call functions or tools defined in your code.

## Overview

Tool calling (also known as function calling) allows Claude to request specific functions to be executed during a conversation. This is useful for tasks that require:

- Retrieving real-time information
- Performing calculations
- Accessing external systems
- Taking actions on behalf of the user

## Basic Usage

### Define Your Tool Schema

First, define a JSON schema for the tools you want Claude to call:

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/yuki5155/go-llms/anthropic-llm/utils"
)

func main() {
	// Initialize the client
	config := utils.NewClientConfig("your-api-key")
	client := utils.NewClient(config)

	// Define the tool schema (following Anthropic's function calling format)
	toolsSchema := []byte(`{
		"tools": [
			{
				"name": "get_weather",
				"description": "Get the current weather in a location",
				"input_schema": {
					"type": "object",
					"properties": {
						"location": {
							"type": "string",
							"description": "The city and state, e.g. San Francisco, CA"
						},
						"unit": {
							"type": "string",
							"enum": ["celsius", "fahrenheit"],
							"description": "The unit of temperature"
						}
					},
					"required": ["location"]
				}
			}
		]
	}`)

	// Create a message
	message := utils.NewTextMessage(utils.RoleUser, "What's the weather like in Tokyo today?")

	// Set up the request options
	options := utils.RequestOptions{
		Messages: []utils.Message{message},
		Schema:   toolsSchema,
		System:   "You are a helpful assistant that can check the current weather.",
	}

	// Send the request
	response, err := client.SendRequestWithTool(options)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Process the tool use request
	if len(response.Message.ToolUses) > 0 {
		for _, toolUse := range response.Message.ToolUses {
			fmt.Printf("Tool called: %s\n", toolUse.Name)
			fmt.Printf("Tool arguments: %s\n", toolUse.Arguments)

			// Parse the arguments
			if toolUse.Name == "get_weather" {
				var args struct {
					Location string `json:"location"`
					Unit     string `json:"unit"`
				}
				if err := json.Unmarshal([]byte(toolUse.Arguments), &args); err != nil {
					log.Fatalf("Error parsing arguments: %v", err)
				}

				// Handle the function call (in a real application, you would call your actual function)
				fmt.Printf("Calling get_weather for location: %s, unit: %s\n", args.Location, args.Unit)
				
				// Here you would get the real weather data
				weatherData := map[string]interface{}{
					"temperature": 22.5,
					"unit": args.Unit,
					"condition": "sunny",
					"location": args.Location,
				}
				
				// Return the result to Claude (not shown in this example)
				// This would involve creating a new message with the tool results
				weatherDataJSON, _ := json.Marshal(weatherData)
				fmt.Printf("Weather data to return: %s\n", string(weatherDataJSON))
			}
		}
	} else {
		// Handle regular response without tool use
		for _, content := range response.Message.Content {
			if content.Type == "text" {
				fmt.Println(content.Text)
			}
		}
	}
}
```

## Complete Conversation Flow with Tool Results

A complete conversation with tool calling would look like this:

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/yuki5155/go-llms/anthropic-llm/utils"
)

func main() {
	// Initialize the client
	config := utils.NewClientConfig("your-api-key")
	client := utils.NewClient(config)

	// Define the tool schema
	toolsSchema := []byte(`{
		"tools": [
			{
				"name": "get_weather",
				"description": "Get the current weather in a location",
				"input_schema": {
					"type": "object",
					"properties": {
						"location": {
							"type": "string",
							"description": "The city and state, e.g. San Francisco, CA"
						},
						"unit": {
							"type": "string",
							"enum": ["celsius", "fahrenheit"],
							"description": "The unit of temperature"
						}
					},
					"required": ["location"]
				}
			}
		]
	}`)

	// Create a message for the initial request
	messages := []utils.Message{
		utils.NewTextMessage(utils.RoleUser, "What's the weather like in Tokyo today?"),
	}

	// Set up the request options
	options := utils.RequestOptions{
		Messages: messages,
		Schema:   toolsSchema,
		System:   "You are a helpful assistant that can check the current weather.",
	}

	// Step 1: Send the request and get the tool call
	response, err := client.SendRequestWithTool(options)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Process the tool use request
	if len(response.Message.ToolUses) > 0 {
		// Add the assistant's message (with the tool call) to the conversation
		assistantMessage := utils.Message{
			Role:    utils.RoleAssistant,
			Content: response.Message.Content,
		}
		messages = append(messages, assistantMessage)

		for _, toolUse := range response.Message.ToolUses {
			if toolUse.Name == "get_weather" {
				var args struct {
					Location string `json:"location"`
					Unit     string `json:"unit"`
				}
				if err := json.Unmarshal([]byte(toolUse.Arguments), &args); err != nil {
					log.Fatalf("Error parsing arguments: %v", err)
				}

				// In a real application, you would call your actual function to get weather data
				// This is simulated data for this example
				weatherData := map[string]interface{}{
					"temperature": 22.5,
					"unit": "celsius",
					"condition": "sunny",
					"location": "Tokyo",
					"humidity": 65,
					"wind": "5 km/h",
				}
				
				// Convert to JSON for the tool result
				weatherDataJSON, _ := json.Marshal(weatherData)
				
				// Step 2: Create a tool results message
				toolResultContent := fmt.Sprintf(`
{
  "type": "tool_result",
  "tool_use_id": "%s",
  "content": %s
}`, toolUse.ID, string(weatherDataJSON))

				// Add the tool result message to the conversation
				toolResultMessage := utils.Message{
					Role: utils.RoleUser,
					Content: []utils.Content{
						{
							Type: "tool_result",
							Text: toolResultContent,
						},
					},
				}
				messages = append(messages, toolResultMessage)
				
				// Step 3: Send the follow-up request with the tool result
				followUpOptions := utils.RequestOptions{
					Messages: messages,
					System:   "You are a helpful assistant that can check the current weather.",
				}
				
				finalResponse, err := client.SendRequestWithTool(followUpOptions)
				if err != nil {
					log.Fatalf("Error in follow-up request: %v", err)
				}
				
				// Display the final response with the weather information incorporated
				if len(finalResponse.Message.Content) > 0 {
					for _, content := range finalResponse.Message.Content {
						if content.Type == "text" {
							fmt.Println("Final response:")
							fmt.Println(content.Text)
						}
					}
				}
			}
		}
	} else {
		// Handle regular response without tool use
		for _, content := range response.Message.Content {
			if content.Type == "text" {
				fmt.Println(content.Text)
			}
		}
	}
}
```

## Helper Methods for Tool Handling

The `anthropic-llm` package includes several helper methods for working with tools:

### GetToolUse

Retrieve a specific tool use by name:

```go
// Get a specific tool by name
weatherTool, err := response.GetToolUse("get_weather")
if err != nil {
	log.Fatalf("Error getting tool: %v", err)
}
fmt.Printf("Tool ID: %s, Name: %s\n", weatherTool.ID, weatherTool.Name)
```

### GetAllToolUses

Retrieve all tool uses with a specific name:

```go
// Get all tools with a specific name (useful if Claude calls the same tool multiple times)
weatherTools, err := response.GetAllToolUses("get_weather")
if err != nil {
	log.Fatalf("Error getting tools: %v", err)
}
fmt.Printf("Found %d weather tool uses\n", len(weatherTools))
```

## Best Practices

1. **Clear Tool Descriptions**: Provide clear and detailed descriptions for each tool
2. **Specific Parameters**: Define specific parameters with clear descriptions for each one
3. **Required Parameters**: Mark essential parameters as required
4. **Handle Errors Gracefully**: Always check for errors and handle them appropriately
5. **Validate Input**: Validate the input parameters before executing the actual function
6. **Maintain Conversation Context**: Include previous messages and tool results in the conversation
7. **Security Considerations**: Validate and sanitize inputs before executing functions

## Limitations

1. Claude may not always use tools even when they would be appropriate
2. Tool descriptions and parameter definitions greatly influence when and how Claude uses the tools
3. Complex tool schemas might not be interpreted correctly
4. You need to implement the actual function logic in your application 
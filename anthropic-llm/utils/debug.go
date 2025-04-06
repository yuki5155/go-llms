package utils

import (
	"encoding/json"
	"fmt"
	"log"
)

// PrintJSON pretty prints any object as JSON for debugging purposes
func PrintJSON(v interface{}) {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Error marshaling to JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonBytes))
}

// DebugResponse prints details from an API response for debugging
func DebugResponse(resp *APIResponse) {
	if resp == nil {
		fmt.Println("Response is nil")
		return
	}
	
	fmt.Printf("Response ID: %s\n", resp.ID)
	fmt.Printf("Model: %s\n", resp.Model)
	fmt.Printf("Message Type: %s\n", resp.Message.Type)
	fmt.Printf("Stop Reason: %s\n", resp.Message.StopReason)
	fmt.Printf("Input Tokens: %d\n", resp.Message.Usage.InputTokens)
	fmt.Printf("Output Tokens: %d\n", resp.Message.Usage.OutputTokens)
	
	// Print content blocks
	fmt.Println("Content blocks:")
	for i, block := range resp.Message.Content {
		fmt.Printf("  Block %d: Type=%s\n", i, block.Type)
		if block.Type == "text" {
			fmt.Printf("    Text: %s\n", block.Text)
		} else if block.Type == "json" {
			fmt.Printf("    JSON: %s\n", string(block.JSON))
		}
	}
	
	// Print tool uses if any
	if len(resp.Message.ToolUses) > 0 {
		fmt.Println("Tool uses:")
		for i, use := range resp.Message.ToolUses {
			fmt.Printf("  Tool %d: ID=%s, Name=%s\n", i, use.ID, use.Name)
			fmt.Printf("    Arguments: %s\n", use.Arguments)
		}
	}
} 
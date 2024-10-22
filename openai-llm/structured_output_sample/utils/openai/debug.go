package openai

import (
	"encoding/json"
	"fmt"
)

func DebugPrintResponse(resp *APIResponse) {
	if resp == nil {
		fmt.Println("Response is nil")
		return
	}

	prettyJSON, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting response for debug: %v\n", err)
		return
	}

	fmt.Printf("Debug Response:\n%s\n", string(prettyJSON))
}

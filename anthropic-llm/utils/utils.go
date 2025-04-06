package utils

import (
	"encoding/json"
	"regexp"
	"strings"
)

// CleanJSONResponse removes markdown backticks and JSON language markers from the response
func CleanJSONResponse(response string) string {
	// Remove markdown code blocks (```json ... ```)
	re := regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)\\s*```")
	matches := re.FindStringSubmatch(response)
	if len(matches) > 1 {
		// Return just the content inside the code block
		return strings.TrimSpace(matches[1])
	}

	// If no code blocks found, just remove any extra whitespace
	return strings.TrimSpace(response)
}

// ParseJSONResponse parses the JSON string into the provided struct
func ParseJSONResponse[T any](jsonStr string) (*T, error) {
	// Clean the JSON response
	cleanJSON := CleanJSONResponse(jsonStr)

	// Parse the JSON
	var result T
	err := json.Unmarshal([]byte(cleanJSON), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

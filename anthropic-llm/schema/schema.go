package schema

import (
	"encoding/json"
)

// BaseSchema is the basic structure for all schemas
type BaseSchema struct {
	Type                 string                    `json:"type"`
	Properties           map[string]SchemaProperty `json:"properties"`
	Required             []string                  `json:"required"`
	AdditionalProperties *bool                     `json:"additionalProperties"`
}

// SchemaProperty represents a property in the JSON schema
type SchemaProperty struct {
	Type        string                    `json:"type"`
	Description string                    `json:"description,omitempty"`
	Enum        []string                  `json:"enum,omitempty"`
	Items       *SchemaProperty           `json:"items,omitempty"`
	Required    []string                  `json:"required,omitempty"`
	Properties  map[string]SchemaProperty `json:"properties,omitempty"`
}

// RequestFormat defines the format for Anthropic API requests
type RequestFormat struct {
	Type     string          `json:"type"`
	Schema   json.RawMessage `json:"schema"`
} 
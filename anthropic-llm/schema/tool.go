package schema

// Tool represents a tool that can be used by the Anthropic model
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function describes a function that can be called by the Anthropic model
type Function struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  BaseSchema `json:"parameters"`
} 
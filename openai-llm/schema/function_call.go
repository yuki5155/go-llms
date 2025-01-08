package schema

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  BaseSchema `json:"parameters"`
	Strict      bool       `json:"strict"`
}

package schema

type SchemaType string

const (
	SchemaTypeObject SchemaType = "object"
	SchemaTypeArray  SchemaType = "array"
	SchemaTypeString SchemaType = "string"
	SchemaTypeNumber SchemaType = "number"
	SchemaTypeBoolean SchemaType = "boolean"
) 
package gqlt

// Summary represents a schema summary
type Summary struct {
	TotalTypes       int    `json:"totalTypes"`
	QueryType        string `json:"queryType,omitempty"`
	MutationType     string `json:"mutationType,omitempty"`
	SubscriptionType string `json:"subscriptionType,omitempty"`
}

// TypeDescription represents a type description
type TypeDescription struct {
	Name        string         `json:"name"`
	Kind        string         `json:"kind"`
	Description string         `json:"description,omitempty"`
	Fields      []FieldSummary `json:"fields,omitempty"`
	InputFields []FieldSummary `json:"inputFields,omitempty"`
	EnumValues  []EnumValue    `json:"enumValues,omitempty"`
}

// FieldDescription represents a field description
type FieldDescription struct {
	RootType    string         `json:"rootType"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Type        string         `json:"type"`
	Arguments   []FieldSummary `json:"arguments,omitempty"`
}

// FieldSummary represents a field summary
type FieldSummary struct {
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	Type         string         `json:"type"`
	Signature    string         `json:"signature"`
	DefaultValue string         `json:"defaultValue,omitempty"`
	Arguments    []FieldSummary `json:"arguments,omitempty"`
}

// EnumValue represents an enum value
type EnumValue struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

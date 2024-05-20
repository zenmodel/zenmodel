package tools

type ToolCallDefinition struct {
	Type     ToolType            `json:"type"`
	Function *FunctionDefinition `json:"function,omitempty"`
	// call function with arguments in JSON format
	// and response in string maybe JSON, it depends on the function.
	// execute this function will really call the function.
	CallFunc CallFunction
}

type ToolType string

const (
	ToolTypeFunction ToolType = "function"
)

type CallFunction func(args string) (resp string, err error)

type FunctionDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Parameters is an object describing the function.
	// You can pass json.RawMessage to describe the schema,
	// or you can pass in a struct which serializes to the proper JSON schema.
	// The jsonschema package is provided for convenience, but you should
	// consider another specialized library if you require more complex schemas.
	Parameters any `json:"parameters"`
}

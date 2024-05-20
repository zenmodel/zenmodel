package openai_structured_output

import "github.com/sashabaranov/go-openai"

type RequestConfig struct {
	Model            string                               `json:"model"`
	MaxTokens        int                                  `json:"max_tokens,omitempty"`
	Temperature      float32                              `json:"temperature,omitempty"`
	TopP             float32                              `json:"top_p,omitempty"`
	N                int                                  `json:"n,omitempty"`
	Stream           bool                                 `json:"stream,omitempty"`
	Stop             []string                             `json:"stop,omitempty"`
	PresencePenalty  float32                              `json:"presence_penalty,omitempty"`
	ResponseFormat   *openai.ChatCompletionResponseFormat `json:"response_format,omitempty"`
	Seed             *int                                 `json:"seed,omitempty"`
	FrequencyPenalty float32                              `json:"frequency_penalty,omitempty"`
	// LogitBias is must be a token id string (specified by their token ID in the tokenizer), not a word string.
	// incorrect: `"logit_bias":{"You": 6}`, correct: `"logit_bias":{"1639": 6}`
	// refs: https://platform.openai.com/docs/api-reference/chat/create#chat/create-logit_bias
	LogitBias map[string]int `json:"logit_bias,omitempty"`
	// LogProbs indicates whether to return log probabilities of the output tokens or not.
	// If true, returns the log probabilities of each output token returned in the content of message.
	// This option is currently not available on the gpt-4-vision-preview model.
	LogProbs bool `json:"logprobs,omitempty"`
	// TopLogProbs is an integer between 0 and 5 specifying the number of most likely tokens to return at each
	// token position, each with an associated log probability.
	// logprobs must be set to true if this parameter is used.
	TopLogProbs int           `json:"top_logprobs,omitempty"`
	User        string        `json:"user,omitempty"`
	Tools       []openai.Tool `json:"tools,omitempty"`
	// This can be either a string or an ToolChoice object.
	ToolChoice any `json:"tool_choice,omitempty"`
}

type StructDefinition struct {
	Type     ToolType            `json:"type"`
	Function *FunctionDefinition `json:"function,omitempty"`

	// empty struct object, like Struct1{}
	StructObj interface{} `json:"-"`
}

type ToolType string

const (
	ToolTypeFunction ToolType = "function"
)

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

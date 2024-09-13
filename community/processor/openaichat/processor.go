package openaichat

import (
	"context"
	"fmt"
	"os"

	"github.com/zenmodel/zenmodel/community/tools"
	"github.com/zenmodel/zenmodel/processor"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

func NewProcessor() *OpenAIChatProcessor {
	processor := &OpenAIChatProcessor{
		memoryKeyMessages: "messages",
		requestConfig: RequestConfig{
			Model:       openai.GPT3Dot5Turbo0125,
			Temperature: 0.7,
			Stream:      false,
		},
		clientConfig: openai.DefaultConfig(os.Getenv("OPENAI_API_KEY")),
	}
	l, _ := zap.NewProductionConfig().Build()
	processor.logger = l

	return processor
}

type OpenAIChatProcessor struct { // nolint
	memoryKeyMessages string
	requestConfig     RequestConfig
	clientConfig      openai.ClientConfig
	client            *openai.Client

	logger *zap.Logger
}

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

func (p *OpenAIChatProcessor) Process(brain processor.BrainContext) error {
	p.logger.Info("openAI chat processor start processing")

	v := brain.GetMemory(p.memoryKeyMessages)
	messages, ok := v.([]openai.ChatCompletionMessage)
	if !ok {
		return fmt.Errorf("assert messages error, memory key: %s, value: %+v", p.memoryKeyMessages, v)
	}

	if p.client == nil {
		p.client = openai.NewClientWithConfig(p.clientConfig)
	}
	resp, err := p.client.CreateChatCompletion(context.Background(),
		openai.ChatCompletionRequest{
			Model:            p.requestConfig.Model,
			MaxTokens:        p.requestConfig.MaxTokens,
			Temperature:      p.requestConfig.Temperature,
			TopP:             p.requestConfig.TopP,
			N:                p.requestConfig.N,
			Stream:           p.requestConfig.Stream,
			Stop:             p.requestConfig.Stop,
			PresencePenalty:  p.requestConfig.PresencePenalty,
			ResponseFormat:   p.requestConfig.ResponseFormat,
			Seed:             p.requestConfig.Seed,
			FrequencyPenalty: p.requestConfig.FrequencyPenalty,
			LogitBias:        p.requestConfig.LogitBias,
			LogProbs:         p.requestConfig.LogProbs,
			TopLogProbs:      p.requestConfig.TopLogProbs,
			User:             p.requestConfig.User,
			Tools:            p.requestConfig.Tools,
			ToolChoice:       p.requestConfig.ToolChoice,

			Messages: messages,
		},
	)
	if err != nil || len(resp.Choices) != 1 {
		return fmt.Errorf("Completion error: err:%v len(choices):%v\n", err,
			len(resp.Choices))
	}

	msg := resp.Choices[0].Message
	p.logger.Debug("LLM respond", zap.Any("response", msg))

	messages = append(messages, msg)
	if err = brain.SetMemory(p.memoryKeyMessages, messages); err != nil {
		return fmt.Errorf("set memory error: %v", err)
	}

	return nil
}

func (p *OpenAIChatProcessor) Clone() processor.Processor {
	return &OpenAIChatProcessor{
		memoryKeyMessages: p.memoryKeyMessages,
		requestConfig:     p.requestConfig,
		clientConfig:      p.clientConfig,
		client:            nil,
		logger:            p.logger.WithOptions(), // with no option, only for clone
	}
}

func (p *OpenAIChatProcessor) WithLogger(logger *zap.Logger) processor.Processor {
	p.logger = logger
	return p
}

func (p *OpenAIChatProcessor) WithMemoryKeyMessages(memoryKeyMessages string) processor.Processor {
	p.memoryKeyMessages = memoryKeyMessages
	return p
}

func (p *OpenAIChatProcessor) WithRequestConfig(requestConfig RequestConfig) processor.Processor {
	p.requestConfig = requestConfig
	return p
}

func (p *OpenAIChatProcessor) WithClientConfig(clientConfig openai.ClientConfig) processor.Processor {
	p.clientConfig = clientConfig
	return p
}

func (p *OpenAIChatProcessor) WithClient(client *openai.Client) processor.Processor {
	p.client = client
	return p
}
func (p *OpenAIChatProcessor) WithToolCallDefinitions(toolCallDefinitions []tools.ToolCallDefinition) processor.Processor {
	toos := make([]openai.Tool, 0)
	for _, toolCallDefinition := range toolCallDefinitions {
		toos = append(toos, openai.Tool{
			Type: openai.ToolType(toolCallDefinition.Type),
			Function: &openai.FunctionDefinition{
				Name:        toolCallDefinition.Function.Name,
				Description: toolCallDefinition.Function.Description,
				Parameters:  toolCallDefinition.Function.Parameters,
			},
		})
		p.requestConfig.Tools = toos
	}

	return p
}

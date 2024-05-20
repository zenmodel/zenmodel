package openai_tool_agent

import (
	"github.com/sashabaranov/go-openai"
	"github.com/zenmodel/zenmodel/community/processor/calltools"
	"github.com/zenmodel/zenmodel/community/processor/openaichat"
	"github.com/zenmodel/zenmodel/community/tools"
	"go.uber.org/zap"
)

type Config struct {
	ToolCallDefinitions []tools.ToolCallDefinition
	ClientConfig        *openai.ClientConfig
	RequestConfig       *openaichat.RequestConfig
	Client              *openai.Client
	MemKeyMessages      *string
	Logger              *zap.Logger
}

func (c Config) newChatProcessor() (*openaichat.OpenAIChatProcessor, error) {
	p := openaichat.NewProcessor()

	if c.ClientConfig != nil {
		p.WithClientConfig(*c.ClientConfig)
	}
	if c.RequestConfig != nil {
		p.WithRequestConfig(*c.RequestConfig)
	}
	if c.Client != nil {
		p.WithClient(c.Client)
	}
	if c.MemKeyMessages != nil {
		p.WithMemoryKeyMessages(*c.MemKeyMessages)
	}
	if len(c.ToolCallDefinitions) != 0 {
		p.WithToolCallDefinitions(c.ToolCallDefinitions)
	}
	if c.Logger != nil {
		p.WithLogger(c.Logger)
	}

	return p, nil
}

func (c Config) newCallToolsProcessor() (*calltools.CallToolsProcessor, error) {
	p := calltools.NewProcessor()

	if c.MemKeyMessages != nil {
		p.WithMemoryKeyMessages(*c.MemKeyMessages)
	}
	if len(c.ToolCallDefinitions) != 0 {
		p.WithToolCallDefinitions(c.ToolCallDefinitions)
	}
	if c.Logger != nil {
		p.WithLogger(c.Logger)
	}

	return p, nil
}

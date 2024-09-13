package main

import (
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/zenmodel/zenmodel/community/processor/go_code_tester"
	"github.com/zenmodel/zenmodel/community/processor/openai_structured_output"
	"github.com/zenmodel/zenmodel/processor"
)

var (
	memKeyCodes = (&go_code_tester.Codes{}).FunctionName()
)

func NewRDProcessor() *CoderProcessor {
	return &CoderProcessor{
		clientConfig: openai.DefaultConfig(os.Getenv("OPENAI_API_KEY")),
		requestConfig: openai_structured_output.RequestConfig{
			Model:       openai.GPT3Dot5Turbo0125,
			Temperature: 0.7,
			Stream:      false,
		},
	}
}

type CoderProcessor struct {
	clientConfig  openai.ClientConfig
	client        *openai.Client
	requestConfig openai_structured_output.RequestConfig
}

func (p *CoderProcessor) Process(b processor.BrainContext) error {
	var prompt string
	if !b.ExistMemory(memKeyCodes) {
		// read task, write code
		prompt = fmt.Sprintf(`{{.%s}}`, memKeyTask)
	} else {
		// read task, old code and test result, write code
		prompt = fmt.Sprintf(`{{.%s}}

My code is as follows:

%s

test result is as follows:

%s

Help me correct my code.
`, memKeyTask, b.GetMemory(memKeyCodes).(*go_code_tester.Codes).String(), memKeyGoTestResult)
	}

	structuredOutput := p.newStructuredOutputProcessor(prompt)
	if err := structuredOutput.Process(b); err != nil {
		return err
	}

	if err := b.SetMemory(memKeyFeedback, b.GetCurrentNeuronID()); err != nil {
		return err
	}

	return nil
}

func (p *CoderProcessor) Clone() processor.Processor {
	return &CoderProcessor{
		requestConfig: p.requestConfig,
		clientConfig:  p.clientConfig,
		client:        nil,
	}
}

func (p *CoderProcessor) WithClientConfig(clientConfig openai.ClientConfig) *CoderProcessor {
	p.clientConfig = clientConfig
	return p
}

func (p *CoderProcessor) WithClient(client *openai.Client) *CoderProcessor {
	p.client = client
	return p
}

func (p *CoderProcessor) WithRequestConfig(requestConfig openai_structured_output.RequestConfig) *CoderProcessor {
	p.requestConfig = requestConfig
	return p
}

func (p *CoderProcessor) newStructuredOutputProcessor(prompt string) *openai_structured_output.OpenAIStructuredOutputProcessor {
	proc := openai_structured_output.NewProcessor()
	_ = proc.WithPromptTemplate(prompt)
	_ = proc.WithOutputStructDefinition(go_code_tester.Codes{}, (go_code_tester.Codes{}).FunctionName(), (go_code_tester.Codes{}).FunctionDescription())
	return proc.WithClientConfig(p.clientConfig).WithRequestConfig(p.requestConfig)
}

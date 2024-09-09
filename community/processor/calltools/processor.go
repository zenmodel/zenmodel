package calltools

import (
	"fmt"

	"github.com/zenmodel/zenmodel/community/tools"
	"github.com/zenmodel/zenmodel/processor"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

func NewProcessor() *CallToolsProcessor {
	processor := &CallToolsProcessor{
		memoryKeyMessages: "messages",
		callFuncMap:       make(map[string]tools.CallFunction),
	}
	l, _ := zap.NewProductionConfig().Build()
	processor.logger = l

	return processor
}

type CallToolsProcessor struct { // nolint
	memoryKeyMessages string
	callFuncMap       map[string]tools.CallFunction

	logger *zap.Logger
}

func (p *CallToolsProcessor) Process(brain processor.BrainContext) error {
	p.logger.Info("call tools from openAI chat processor start processing")

	v := brain.GetMemory(p.memoryKeyMessages)
	messages, ok := v.([]openai.ChatCompletionMessage)
	if !ok {
		return fmt.Errorf("assert messages error, memory key: %s, value: %+v", p.memoryKeyMessages, v)
	}
	lastMsg := messages[len(messages)-1]

	// call every tools
	for i, call := range lastMsg.ToolCalls {
		fn, ok := p.callFuncMap[call.Function.Name]
		if !ok {
			return fmt.Errorf("function %s not define", call.Function.Name)
		}
		resp, err := fn(call.Function.Arguments)
		if err != nil {
			return fmt.Errorf("call function %s with args %s error: %v",
				call.Function.Name, call.Function.Arguments, err)
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    resp,
			Name:       lastMsg.ToolCalls[i].Function.Name,
			ToolCallID: lastMsg.ToolCalls[i].ID,
		})
	}

	if err := brain.SetMemory(p.memoryKeyMessages, messages); err != nil {
		return fmt.Errorf("set memory error: %v", err)
	}

	return nil
}

func (p *CallToolsProcessor) Clone() processor.Processor {
	callFuncMap := make(map[string]tools.CallFunction)
	for name, callFunc := range p.callFuncMap {
		callFuncMap[name] = callFunc
	}

	return &CallToolsProcessor{
		memoryKeyMessages: p.memoryKeyMessages,
		callFuncMap:       callFuncMap,
		logger:            p.logger.WithOptions(), // with no option, only for clone
	}
}

func (p *CallToolsProcessor) WithLogger(logger *zap.Logger) processor.Processor {
	p.logger = logger
	return p
}

func (p *CallToolsProcessor) WithMemoryKeyMessages(memoryKeyMessages string) processor.Processor {
	p.memoryKeyMessages = memoryKeyMessages
	return p
}

func (p *CallToolsProcessor) WithToolCallDefinitions(toolCallDefinitions []tools.ToolCallDefinition) processor.Processor {
	callFuncMap := make(map[string]tools.CallFunction)
	for _, definition := range toolCallDefinitions {
		callFuncMap[definition.Function.Name] = definition.CallFunc
	}
	p.callFuncMap = callFuncMap

	return p
}

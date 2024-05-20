package openai_structured_output

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"github.com/invopop/jsonschema"
	"github.com/sashabaranov/go-openai"
	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/community/common/log"
	"go.uber.org/zap"
)

func NewProcessor() *OpenAIStructuredOutputProcessor {
	processor := &OpenAIStructuredOutputProcessor{
		requestConfig: RequestConfig{
			Model:       openai.GPT3Dot5Turbo0125,
			Temperature: 0.7,
			Stream:      false,
		},
		clientConfig: openai.DefaultConfig(os.Getenv("OPENAI_API_KEY")),
		structDefs:   make(map[string]StructDefinition),
		logger:       log.NewDefaultLoggerWithLevel(zap.InfoLevel),
	}

	return processor
}

type OpenAIStructuredOutputProcessor struct { // nolint
	promptTemplate string
	variables      []string

	requestConfig RequestConfig
	clientConfig  openai.ClientConfig
	client        *openai.Client
	structDefs    map[string]StructDefinition

	// internal vars
	prompt  string
	respMsg openai.ChatCompletionMessage

	logger *zap.Logger
}

func (p *OpenAIStructuredOutputProcessor) Process(brain zenmodel.BrainRuntime) error {
	p.logger.Info("openAI structured output processor start processing")

	if err := p.renderPrompt(brain); err != nil {
		return err
	}

	// load struct definitions to requestConfig.Tools
	tools := make([]openai.Tool, 0)
	for name, structDef := range p.structDefs {
		tools = append(tools, openai.Tool{
			Type: openai.ToolType(structDef.Type),
			Function: &openai.FunctionDefinition{
				Name:        name,
				Description: structDef.Function.Description,
				Parameters:  structDef.Function.Parameters,
			},
		})
	}
	p.requestConfig.Tools = tools

	if err := p.chatCompletion(context.Background()); err != nil {
		return err
	}

	if err := p.structureOutput(brain); err != nil {
		return err
	}

	return nil
}

func (p *OpenAIStructuredOutputProcessor) DeepCopy() zenmodel.Processor {
	// variables value copy
	variablesCopy := make([]string, len(p.variables))
	if p.variables != nil {
		copy(variablesCopy, p.variables)
	}
	// structDefs value copy
	structDefsCopy := make(map[string]StructDefinition)
	for k, v := range p.structDefs {
		structDefsCopy[k] = v
	}

	return &OpenAIStructuredOutputProcessor{
		promptTemplate: p.promptTemplate,
		variables:      variablesCopy,
		prompt:         "", // should not be copied
		requestConfig:  p.requestConfig,
		clientConfig:   p.clientConfig,
		client:         nil,
		structDefs:     structDefsCopy,
		logger:         p.logger.WithOptions(), // with no option, only for clone
	}
}

// WithOutputStructDefinition this method Must be called, and can be called multiple times to define multiple output structures
//
//	structObj - empty struct object, like StructABCD{}. struct define should have jsonschema tag for converting to json schema.
//				you can find tag definition rules in https://github.com/invopop/jsonschema
//	name - name of the struct, structured object will be set into brain memory with this name as memory key
//	description - description of the struct, provide to LLM to determine which type of structure to use for output
func (p *OpenAIStructuredOutputProcessor) WithOutputStructDefinition(structObj interface{}, name string, description string) error {
	def := StructDefinition{
		Type: ToolTypeFunction,
		Function: &FunctionDefinition{
			Name:        name,
			Description: description,
		},
		StructObj: structObj,
	}
	params, err := extractStructParameter(structObj)
	if err != nil {
		return err
	}
	def.Function.Parameters = params
	p.structDefs[name] = def

	p.logger.Info("struct definition", zap.Any("definition", def))
	return nil
}

func (p *OpenAIStructuredOutputProcessor) WithPromptTemplate(promptTemplate string) error {
	// 尝试解析模板来检查它是否有效
	_, err := template.New("validate").Parse(promptTemplate)
	if err != nil {
		return err
	}

	// 如果模板有效，将其存储在处理器中
	p.promptTemplate = promptTemplate

	// 用正则表达式提取模板中的变量名
	re := regexp.MustCompile(`{{(.+?)}}`)
	matches := re.FindAllStringSubmatch(p.promptTemplate, -1)

	// 存储提取出的所有变量名
	variables := make([]string, len(matches))
	for _, match := range matches {
		// 清理变量名并存储
		varName := strings.TrimPrefix(strings.TrimSpace(match[1]), `.`)
		variables = append(variables, varName)
	}
	p.variables = variables
	p.logger.Debug("Successfully parsed prompt template and extracted variables", zap.Strings("variables", variables))

	return nil
}

func (p *OpenAIStructuredOutputProcessor) WithLogger(logger *zap.Logger) *OpenAIStructuredOutputProcessor {
	p.logger = logger
	return p
}

func (p *OpenAIStructuredOutputProcessor) WithRequestConfig(requestConfig RequestConfig) *OpenAIStructuredOutputProcessor {
	p.requestConfig = requestConfig
	return p
}

func (p *OpenAIStructuredOutputProcessor) WithClientConfig(clientConfig openai.ClientConfig) *OpenAIStructuredOutputProcessor {
	p.clientConfig = clientConfig
	return p
}

func (p *OpenAIStructuredOutputProcessor) WithClient(client *openai.Client) *OpenAIStructuredOutputProcessor {
	p.client = client
	return p
}

func (p *OpenAIStructuredOutputProcessor) renderPrompt(brain zenmodel.BrainRuntime) error {
	values := make(map[string]string)
	for _, varName := range p.variables {
		values[varName] = fmt.Sprintf("%s", brain.GetMemory(varName))
	}

	tmpl, err := template.New("prompt").Parse(p.promptTemplate)
	if err != nil {
		return err
	}
	var prompt bytes.Buffer
	if err = tmpl.Execute(&prompt, values); err != nil {
		return err
	}
	p.prompt = prompt.String()
	p.logger.Debug("prompt has rendered", zap.String("prompt", p.prompt))

	return nil
}

func (p *OpenAIStructuredOutputProcessor) chatCompletion(ctx context.Context) error {
	if p.client == nil {
		p.client = openai.NewClientWithConfig(p.clientConfig)
	}
	ccr := openai.ChatCompletionRequest{
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

		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: p.prompt,
			},
		},
	}
	p.logger.Debug("Call OpenAI CreateChatCompletion", zap.Any("ChatCompletionRequest", ccr))
	resp, err := p.client.CreateChatCompletion(ctx, ccr)
	if err != nil || len(resp.Choices) != 1 {
		return fmt.Errorf("Completion error: err:%v len(choices):%v\n", err,
			len(resp.Choices))
	}
	msg := resp.Choices[0].Message
	p.respMsg = msg

	p.logger.Debug("LLM respond", zap.Any("response message", msg))
	return nil
}

func (p *OpenAIStructuredOutputProcessor) structureOutput(brain zenmodel.BrainRuntime) error {
	if len(p.respMsg.ToolCalls) == 0 {
		return nil
	}
	for _, call := range p.respMsg.ToolCalls {
		name := call.Function.Name
		args := call.Function.Arguments

		obj, err := unmarshalArgumentsToStruct([]byte(args), p.structDefs[name].StructObj)
		if err != nil {
			return fmt.Errorf("unmarshal arguments to struct object error: %v", err)
		}

		if err = brain.SetMemory(name, obj); err != nil {
			return fmt.Errorf("set memory error: %v", err)
		}
	}

	return nil
}

func unmarshalArgumentsToStruct(data []byte, structObj interface{}) (interface{}, error) {
	t := reflect.TypeOf(structObj)
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("arg is not a struct")
	}

	vp := reflect.New(t)
	v := vp.Elem()
	obj := v.Addr().Interface()

	err := json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func extractStructParameter(structObj interface{}) (map[string]interface{}, error) {
	// Get the type of the input function
	structType := reflect.TypeOf(structObj)

	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct")
	}

	// Create a new instance of the struct type
	structValue := reflect.New(structType).Elem().Interface()

	parameter, err := structAsJSONSchema(structValue)
	if err != nil {
		return nil, err
	}

	return parameter, nil
}

func structAsJSONSchema(v interface{}) (map[string]interface{}, error) {
	r := new(jsonschema.Reflector)
	r.DoNotReference = true
	schema := r.Reflect(v)

	b, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	var jsonSchema map[string]interface{}
	err = json.Unmarshal(b, &jsonSchema)
	if err != nil {
		return nil, err
	}

	delete(jsonSchema, "$schema")

	return jsonSchema, nil
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/zenmodel/zenmodel"
)

func main() {
	bp := zenmodel.NewBrainPrint()
	// add neuron
	llm := bp.AddNeuron(chatLLM)
	action := bp.AddNeuron(callTools)

	// add link
	_, _ = bp.AddEntryLink(llm)
	continueLink, _ := bp.AddLink(llm, action)
	_, _ = bp.AddLink(action, llm)

	endLink, _ := bp.AddEndLink(llm)

	// set cast selection
	_ = bp.AddLinkToCastGroup(llm, "continue", continueLink)
	_ = bp.AddLinkToCastGroup(llm, "end", endLink)
	_ = bp.BindCastGroupSelectFunc(llm, llmNext)

	brain := bp.Build()
	_ = brain.EntryWithMemory(
		"messages", []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "What is the weather in Boston today?"}})

	// block process util brain sleeping
	for brain.GetState() != zenmodel.BrainStateSleeping {
		time.Sleep(1 * time.Second)
	}

	v, found := brain.GetMemory("messages")
	if found {
		messages, _ := json.Marshal(v)
		fmt.Printf("messages: %s\n", messages)
	}
}

// describe the function & its inputs
var tools = []openai.Tool{
	{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_current_weather",
			Description: "Get the current weather in a given location",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"location": {
						Type:        jsonschema.String,
						Description: "The city and state, e.g. San Francisco, CA",
					},
					"unit": {
						Type: jsonschema.String,
						Enum: []string{"celsius", "fahrenheit"},
					},
				},
				Required: []string{"location"},
			},
		},
	},
}

func chatLLM(b zenmodel.Brain) error {
	fmt.Println("run here chatLLM...")

	// get need info form memory
	v, found := b.GetMemory("messages")
	if !found {
		return fmt.Errorf("memory [%s] not found", "messages")
	}
	messages := v.([]openai.ChatCompletionMessage)

	ctx := context.Background()
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo0125,
			Messages: messages,
			Tools:    tools,
		},
	)
	if err != nil || len(resp.Choices) != 1 {
		return fmt.Errorf("Completion error: err:%v len(choices):%v\n", err,
			len(resp.Choices))
	}

	msg := resp.Choices[0].Message
	fmt.Printf("LLM response: %+v\n", msg)
	messages = append(messages, msg)
	_ = b.SetMemory("messages", messages)

	return nil
}

func callTools(b zenmodel.Brain) error {
	fmt.Println("run here callTools...")

	// get need info form memory
	v, found := b.GetMemory("messages")
	if !found {
		return fmt.Errorf("memory [%s] not found", "messages")
	}
	messages := v.([]openai.ChatCompletionMessage)
	lastMsg := messages[len(messages)-1]

	for _, call := range lastMsg.ToolCalls {
		if call.Function.Name == "get_current_weather" {
			// 根据 call.Function.Name 和 call.Function.Arguments 发起函数调用，此处模拟调用并 mock 调用结果
			fmt.Printf("call tool [%s] with arguments [%s]\n", call.Function.Name, call.Function.Arguments)
			// mock 返回 toolCalledResp
			toolCalledResp := "Sunny and 80 degrees."

			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    toolCalledResp,
				Name:       lastMsg.ToolCalls[0].Function.Name,
				ToolCallID: lastMsg.ToolCalls[0].ID,
			})
		}
	}
	_ = b.SetMemory("messages", messages)

	return nil
}

func llmNext(b zenmodel.Brain) string {
	v, found := b.GetMemory("messages")
	if !found {
		return "end"
	}
	messages := v.([]openai.ChatCompletionMessage)
	lastMsg := messages[len(messages)-1]
	if len(lastMsg.ToolCalls) == 0 { // no need to call any tools
		return "end"
	}

	return "continue"
}

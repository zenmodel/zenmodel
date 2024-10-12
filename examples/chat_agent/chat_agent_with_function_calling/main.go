package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/brainlocal"
	"github.com/zenmodel/zenmodel/processor"
)

func main() {
	bp := zenmodel.NewBlueprint()

	// add neuron
	llm := bp.AddNeuron(chatLLM)
	action := bp.AddNeuron(callTools)

	/* This example omits error handling */
	// add entry link
	_, _ = bp.AddEntryLinkTo(llm)

	// add link
	continueLink, _ := bp.AddLink(llm, action)
	_, _ = bp.AddLink(action, llm)

	// add end link
	endLink, _ := bp.AddEndLinkFrom(llm)

	// add link to cast group of a neuron
	_ = llm.AddCastGroup("continue", continueLink)
	_ = llm.AddCastGroup("end", endLink)
	// bind cast group select function for neuron
	llm.BindCastGroupSelectFunc(llmNext)

	// build brain
	brain := brainlocal.BuildBrain(bp)
	// set memory and trig all entry links
	_ = brain.EntryWithMemory(
		"messages", []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "What is the weather in Boston today?"}})

	// block process util brain sleeping
	brain.Wait()

	messages, _ := json.Marshal(brain.GetMemory("messages"))
	fmt.Printf("messages: %s\n", messages)
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

func chatLLM(bc processor.BrainContext) error {
	fmt.Println("run here chatLLM...")

	// get need info form memory
	messages, _ := bc.GetMemory("messages").([]openai.ChatCompletionMessage)

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
	_ = bc.SetMemory("messages", messages)

	return nil
}

func callTools(bc processor.BrainContext) error {
	fmt.Println("run here callTools...")

	// get need info form memory
	messages, _ := bc.GetMemory("messages").([]openai.ChatCompletionMessage)
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
	_ = bc.SetMemory("messages", messages)

	return nil
}

func llmNext(bcr processor.BrainContextReader) string {
	if !bcr.ExistMemory("messages") {
		return "end"
	}
	messages, _ := bcr.GetMemory("messages").([]openai.ChatCompletionMessage)
	lastMsg := messages[len(messages)-1]
	if len(lastMsg.ToolCalls) == 0 { // no need to call any tools
		return "end"
	}

	return "continue"
}

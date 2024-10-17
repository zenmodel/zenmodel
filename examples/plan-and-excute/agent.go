//go:build ignore

package main

import (
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/zenmodel/zenmodel/brainlocal"
	"github.com/zenmodel/zenmodel/community/brain/openai_tool_agent"
	"github.com/zenmodel/zenmodel/community/processor/openaichat"
	"github.com/zenmodel/zenmodel/community/tools"
	"github.com/zenmodel/zenmodel/processor"
)

type PastStep struct {
	Task   string `json:"task"`
	Result string `json:"result"`
}

type PastSteps []PastStep

func (s PastSteps) String() string {
	var builder strings.Builder
	for i, step := range s {
		builder.WriteString(fmt.Sprintf("step%d:\n\ttask: %s\n\tresult: %s\n\n", i+1, step.Task, step.Result))
	}
	return builder.String()
}

func toolAgentProcess(b processor.BrainContext) error {
	plan, ok := b.GetMemory(memKeyPlan).(*Plan)
	if !ok {
		return fmt.Errorf("assert plan error, plan: %+v", b.GetMemory(memKeyPlan))
	}

	task := plan.Steps[0]
	// use tool agent by nested brain
	result, err := SearchAgent(task)
	if err != nil {
		return err
	}

	step := PastStep{
		Task:   task,
		Result: result,
	}

	var steps PastSteps
	if !b.ExistMemory(memKeyPastSteps) {
		steps = PastSteps{step}
	} else {
		steps = b.GetMemory(memKeyPastSteps).(PastSteps)
		steps = append(steps, step)
	}
	if err = b.SetMemory(memKeyPastSteps, steps); err != nil {
		return err
	}

	return nil
}

// SearchAgent by nested brain
func SearchAgent(query string) (result string, err error) {
	reqConfig := openaichat.RequestConfig{
		Model:       openai.GPT3Dot5Turbo0613,
		Temperature: 0.7,
	}
	// clone community shared brainprint, and set some tool cal definitions(support multi definitions)
	cfg := openai_tool_agent.Config{
		ToolCallDefinitions: []tools.ToolCallDefinition{tools.DuckDuckGoSearchToolCallDefinition()},
		RequestConfig:       &reqConfig,
	}
	bp, err := openai_tool_agent.CloneBrainprint(cfg)
	if err != nil {
		return "", err
	}
	// build brain
	brain := brainlocal.BuildBrain(bp)
	// set memory and trig all entry links
	if err = brain.EntryWithMemory(
		"messages", []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: query}},
	); err != nil {
		return "", err
	}

	// block process util brain sleeping
	brain.Wait()

	// get messages finally
	msgs, ok := brain.GetMemory("messages").([]openai.ChatCompletionMessage)
	if !ok {
		return "", fmt.Errorf("assert messages error, messages: %+v", brain.GetMemory("messages"))
	}
	latestMsg := msgs[len(msgs)-1]

	return latestMsg.Content, nil
}

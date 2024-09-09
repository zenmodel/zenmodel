package openai_tool_agent

import (
	"github.com/sashabaranov/go-openai"
	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/brain"
	"github.com/zenmodel/zenmodel/processor"
)

func CloneBrainprint(config Config) (brain.Blueprint, error) {
	bp := zenmodel.NewBlueprint()

	chatProcessor, err := config.newChatProcessor()
	if err != nil {
		return nil, err
	}
	callToolsProcessor, err := config.newCallToolsProcessor()
	if err != nil {
		return nil, err
	}

	// add neuron
	llm := bp.AddNeuronWithProcessor(chatProcessor)
	action := bp.AddNeuronWithProcessor(callToolsProcessor)

	// add entry link
	_, err = bp.AddEntryLinkTo(llm)
	if err != nil {
		return nil, err
	}
	// add link
	continueLink, err := bp.AddLink(llm, action)
	if err != nil {
		return nil, err
	}
	_, err = bp.AddLink(action, llm)
	if err != nil {
		return nil, err
	}

	// add end link
	endLink, err := bp.AddEndLinkFrom(llm)
	if err != nil {
		return nil, err
	}

	// add link to cast group of a neuron
	if err = llm.AddCastGroup("continue", continueLink); err != nil {
		return nil, err
	}
	if err = llm.AddCastGroup("end", endLink); err != nil {
		return nil, err
	}

	// bind cast group select function for neuron
	llm.BindCastGroupSelectFunc(llmNext)

	return bp, nil
}

func llmNext(b processor.BrainContextReader) string {
	if !b.ExistMemory("messages") {
		return "end"
	}
	messages, _ := b.GetMemory("messages").([]openai.ChatCompletionMessage)
	lastMsg := messages[len(messages)-1]
	if len(lastMsg.ToolCalls) == 0 { // no need to call any tools
		return "end"
	}

	return "continue"
}

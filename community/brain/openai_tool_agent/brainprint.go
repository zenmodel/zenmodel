package openai_tool_agent

import (
	"github.com/sashabaranov/go-openai"
	"github.com/zenmodel/zenmodel"
)

func CloneBrainprint(config Config) (*zenmodel.Brainprint, error) {
	bp := zenmodel.NewBrainPrint()

	chatProcessor, err := config.newChatProcessor()
	if err != nil {
		return nil, err
	}
	callToolsProcessor, err := config.newCallToolsProcessor()
	if err != nil {
		return nil, err
	}

	// add neuron
	bp.AddNeuronWithProcessor("llm", chatProcessor)
	bp.AddNeuronWithProcessor("action", callToolsProcessor)

	// add entry link
	_, err = bp.AddEntryLink("llm")
	if err != nil {
		return nil, err
	}
	// add link
	continueLink, err := bp.AddLink("llm", "action")
	if err != nil {
		return nil, err
	}
	_, err = bp.AddLink("action", "llm")
	if err != nil {
		return nil, err
	}

	// add end link
	endLink, err := bp.AddEndLink("llm")
	if err != nil {
		return nil, err
	}

	// add link to cast group of a neuron
	if err = bp.AddLinkToCastGroup("llm", "continue", continueLink); err != nil {
		return nil, err
	}
	if err = bp.AddLinkToCastGroup("llm", "end", endLink); err != nil {
		return nil, err
	}
	// bind cast group select function for neuron
	if err = bp.BindCastGroupSelectFunc("llm", llmNext); err != nil {
		return nil, err
	}

	return bp, nil
}

func llmNext(b zenmodel.BrainRuntime) string {
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

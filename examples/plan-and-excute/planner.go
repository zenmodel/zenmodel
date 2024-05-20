package main

import (
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/zenmodel/zenmodel/community/processor/openai_structured_output"
)

var (
	plannerPromptTemple = fmt.Sprintf(`For the given objective, come up with a simple step by step plan.
This plan should involve individual tasks, that if executed correctly will yield the correct answer. Do not add any superfluous steps.
The result of the final step should be the final answer. Make sure that each step has all the information needed - do not skip steps.

{{.%s}}`,
		memKeyObjective)
)

// Plan to follow in future
type Plan struct {
	Steps []string `json:"steps" jsonschema_description:"different steps to follow should be in sorted order"`
}

func (p *Plan) String() string {
	display := make([]string, len(p.Steps))
	for i, step := range p.Steps {
		display[i] = step
	}
	ret, _ := json.Marshal(display)
	return string(ret)
}

func PlannerProcessor() (*openai_structured_output.OpenAIStructuredOutputProcessor, error) {
	p := openai_structured_output.NewProcessor().
		WithRequestConfig(openai_structured_output.RequestConfig{
			Model:       openai.GPT4Turbo0125,
			Temperature: 0,
		})
	_ = p.WithPromptTemplate(plannerPromptTemple)
	_ = p.WithOutputStructDefinition(Plan{}, memKeyPlan, "Plan to follow in future")

	return p, nil
}

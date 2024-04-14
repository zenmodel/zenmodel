package main

import (
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/zenmodel/zenmodel-contrib/processor/openai_structured_output"
)

var (
	replannerPromptTemple = fmt.Sprintf(`For the given objective, come up with a simple step by step plan.
This plan should involve individual tasks, that if executed correctly will yield the correct answer. Do not add any superfluous steps.
The result of the final step should be the final answer. Make sure that each step has all the information needed - do not skip steps.

Your objective was this:
{{.%s}}

Your original plan was this:
{{.%s}}

You have currently done the follow steps:
{{.%s}}

Update your plan accordingly. If no more steps are needed and you can return to the user, then respond with that. Otherwise, fill out the plan. Only add steps to the plan that still NEED to be done. Do not return previously done steps as part of the plan.
`,
		memKeyObjective,
		memKeyPlan,
		memKeyPastSteps,
	)
)

type Response struct {
	Response string `json:"response" jsonschema:"description=response answer to user"`
}

func RePlannerProcessor() (*openai_structured_output.OpenAIStructuredOutputProcessor, error) {
	p := openai_structured_output.NewProcessor().
		WithRequestConfig(openai_structured_output.RequestConfig{
			Model:       openai.GPT4Turbo0125,
			Temperature: 0,
		})
	_ = p.WithPromptTemplate(replannerPromptTemple)
	_ = p.WithOutputStructDefinition(Plan{}, memKeyPlan, "Plan to follow in future")
	_ = p.WithOutputStructDefinition(Response{}, memKeyResponse, "Response answer to user")

	return p, nil
}

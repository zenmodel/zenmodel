---
title: Multi-Agent
weight: 20
menu:
  notes:
    name: Multi-Agent
    identifier: multi-agent
    parent: agents
    weight: 20
---

<!-- Graph -->
{{< note title="Graph" >}}

<img src="https://github.com/zenmodel/zenmodel/blob/main/examples/multi-agent/agent-supervisor/agent-supervisor.png?raw=true">

{{< /note >}}

<!-- Brain -->
{{< note title="Brain" >}}

```go
package main

import (
	"fmt"

	"github.com/zenmodel/zenmodel"
)

var (
	NeuronLeader     = "Leader"
	NeuronRD         = "RD"
	NeuronQA         = "QA"
	DecisionRD       = "RD"
	DecisionQA       = "QA"
	DecisionResponse = "Response"
)

func main() {
	bp := zenmodel.NewBrainPrint()

	bp.AddNeuron(NeuronLeader, LeaderProcess)
	bp.AddNeuron(NeuronQA, QAProcess)
	bp.AddNeuronWithProcessor(NeuronRD, NewRDProcessor())

	_, _ = bp.AddEntryLink(NeuronLeader)
	// leader out-link
	rdLink, _ := bp.AddLink(NeuronLeader, NeuronRD)
	qaLink, _ := bp.AddLink(NeuronLeader, NeuronQA)
	endLink, _ := bp.AddEndLink(NeuronLeader)

	// leader in-link
	_, _ = bp.AddLink(NeuronRD, NeuronLeader)
	_, _ = bp.AddLink(NeuronQA, NeuronLeader)

	_ = bp.AddLinkToCastGroup(NeuronLeader, DecisionRD, rdLink)
	_ = bp.AddLinkToCastGroup(NeuronLeader, DecisionQA, qaLink)
	_ = bp.AddLinkToCastGroup(NeuronLeader, DecisionResponse, endLink)
	_ = bp.BindCastGroupSelectFunc(NeuronLeader, func(b zenmodel.BrainRuntime) string {
		return b.GetMemory(memKeyDecision).(string)
	})

	brain := bp.Build()
	_ = brain.EntryWithMemory(memKeyDemand, "Help me write a function `func Add (x, y int) int` with golang to implement addition, and implement unit test in a separate _test .go file, at least 3 test cases are required")
	brain.Wait()
	fmt.Printf("Response: %s\n", brain.GetMemory(memKeyResponse).(string))
}


```

{{< /note >}}

<!-- Response -->
{{< note title="Response" >}}

```
Dear Boss:
After the efforts of our RD team and QA team, the final codes and test report are produced as follows:

==========

Codes:

**add.go**

```go
package main

func Add(x, y int) int {
		return x + y
}
```

**add_test.go**

```go
package main

import "testing"

func TestAdd(t *testing.T) {
		cases := []struct {
				x, y, expected int
		}{
				{1, 2, 3},
				{-1, 1, 0},
				{0, 0, 0},
		}

		for _, c := range cases {
				result := Add(c.x, c.y)
				if result != c.expected {
						t.Errorf("Add(%d, %d) == %d, expected %d", c.x, c.y, result, c.expected)
				}
		}
}
```

==========

Test Report:

```shell
#go test -v -run .
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
PASS
ok      gocodetester    0.411s

```


	
```
{{< /note >}}


<!-- Leader -->
{{< note title="Leader" >}}

```go
package main

import (
	"fmt"
	"strings"

	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/community/processor/go_code_tester"
)

const (
	memKeyDemand   = "demand"
	memKeyResponse = "response"
	memKeyTask     = "task"
	memKeyFeedback = "feedback"
	memKeyDecision = "decision"
)

func LeaderProcess(b zenmodel.BrainRuntime) error {
	// if it has no task, disassemble task from demand
	if !b.ExistMemory(memKeyTask) {
		task := rephraseTaskFromDemand(b.GetMemory(memKeyDemand).(string))
		_ = b.SetMemory(memKeyTask, task)
		_ = b.SetMemory(memKeyDecision, DecisionRD)

		return nil
	}
	switch b.GetMemory(memKeyFeedback).(string) {
	case NeuronRD: // feedback from RD
		_ = b.SetMemory(memKeyDecision, DecisionQA) // pass to QA
	case NeuronQA: // feedback from QA
		ok := readTestReport(b.GetMemory(memKeyGoTestResult).(string))
		if !ok {
			// test result not ok, resend to RD
			_ = b.SetMemory(memKeyDecision, DecisionRD)
		} else {
			// pretty response from codes
			resp := genResponse(b)
			_ = b.SetMemory(memKeyResponse, resp)
			_ = b.SetMemory(memKeyDecision, DecisionResponse)
		}
	default:
		return fmt.Errorf("unknown feedback: %v\n", b.GetMemory(memKeyFeedback))
	}

	return nil
}

func rephraseTaskFromDemand(demand string) string {
	// TODO maybe use completion LLM to rephrase demand to task
	task := demand

	return task
}

func readTestReport(testResult string) bool {
	return !strings.Contains(testResult, "FAIL")
}

func genResponse(b zenmodel.BrainRuntime) string {
	codes := b.GetMemory(memKeyCodes).(*go_code_tester.Codes).String()
	testReport := b.GetMemory(memKeyGoTestResult).(string)

	var builder strings.Builder
	builder.WriteString("Dear Boss:  \n")
	builder.WriteString("After the efforts of our RD team and QA team, the final codes and test report are produced as follows:\n\n")
	builder.WriteString("==========\n\nCodes:\n\n")
	builder.WriteString(codes)
	builder.WriteString("==========\n\nTest Report:\n\n")
	builder.WriteString("```shell\n")
	builder.WriteString(testReport)
	builder.WriteString("```")
	builder.WriteString("\n")

	return builder.String()
}

```


{{< /note >}}

<!-- RD -->
{{< note title="RD" >}}

```go
package main

import (
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/community/processor/go_code_tester"
	"github.com/zenmodel/zenmodel/community/processor/openai_structured_output"
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

func (p *CoderProcessor) Process(b zenmodel.BrainRuntime) error {
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

func (p *CoderProcessor) DeepCopy() zenmodel.Processor {
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
	processor := openai_structured_output.NewProcessor()
	_ = processor.WithPromptTemplate(prompt)
	_ = processor.WithOutputStructDefinition(go_code_tester.Codes{}, (go_code_tester.Codes{}).FunctionName(), (go_code_tester.Codes{}).FunctionDescription())
	return processor.WithClientConfig(p.clientConfig).WithRequestConfig(p.requestConfig)
}

```
{{< /note >}}


<!-- QA -->
{{< note title="QA" >}}

```go
package main

import (
	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/community/processor/go_code_tester"
)

const (
	memKeyGoTestResult = "go_test_result"
)

func QAProcess(b zenmodel.BrainRuntime) error {
	p := go_code_tester.NewProcessor().WithTestCodeKeep(true)
	if err := p.Process(b); err != nil {
		return err
	}

	if err := b.SetMemory(memKeyFeedback, b.GetCurrentNeuronID()); err != nil {
		return err
	}

	return nil
}

```
{{< /note >}}
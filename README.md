# ZenModel
[![go report card](https://goreportcard.com/badge/github.com/zenmodel/zenmodel "go report card")](https://goreportcard.com/report/github.com/zenmodel/zenmodel)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/zenmodel/zenmodel)
[![GoDoc](https://pkg.go.dev/badge/github.com/zenmodel/zenmodel?status.svg)](https://pkg.go.dev/github.com/zenmodel/zenmodel?tab=doc)
![GitHub License](https://img.shields.io/github/license/zenmodel/zenmodel)
[![](https://dcbadge.vercel.app/api/server/6YhZquB4zb?compact=true&style=flat)](https://discord.gg/6YhZquB4zb)

[//]: # ([![Sourcegraph]&#40;https://sourcegraph.com/github.com/zenmodel/zenmodel/-/badge.svg&#41;]&#40;https://sourcegraph.com/github.com/zenmodel/zenmodel?badge&#41;)

[//]: # ([![Release]&#40;https://img.shields.io/github/release/zenmodel/zenmodel.svg?style=flat-square&#41;]&#40;https://github.com/zenmodel/zenmodel/releases&#41;)

[中文](./README_zh.md) | [English](./README.md)

***Use Golang to develop Agentic applications with LLMs***

## Overview

[ZenModel](https://github.com/zenmodel/zenmodel) is a workflow programming framework designed for constructing agentic applications with LLMs. It implements by the scheduling of computational units (`Neuron`), that may include loops, by constructing a `Brain` (a directed graph that can have cycles) or support the loop-less DAGs. A `Brain` consists of multiple `Neurons` connected by `Link`s. Inspiration was drawn from [LangGraph](https://github.com/langchain-ai/langgraph).

ZenModel supports multiple implementations of the `Brain` interface:

1. **BrainLocal**: The default implementation. It uses [ristretto](https://github.com/dgraph-io/ristretto) for in-memory `Memory` management.

2. **BrainLite**: A lightweight implementation that uses SQLite for `Memory` management, allowing for persistent storage and potential support for multi-language Processors.

Developers can choose the appropriate Brain implementation based on their specific requirements.

- Developers can build a `Brain` with any process flow:
    - Sequential: Execute `Neuron`s in order.
    - Parallel and Wait: Concurrent execution of `Neuron`s with support for downstream `Neuron`s to wait until all the specified upstream ones have completed before starting.
    - Branch: Execution flow only propagates to certain downstream branches.
    - Looping: Loops are essential for agent-like behaviors, where you would call an LLM in a loop to inquire about the next action to take.
    - With-End: Stops running under specific conditions, such as after obtaining the desired result.
    - Open-Ended: Continuously runs, for instance, in the scenario of a voice call, constantly listening to the user.
- Each `Neuron` is a concrete computational unit, and developers can customize `Neuron` to implement any processing procedure (`Processor`), including LLM calls, other multimodal model invocations, and control mechanisms like timeouts and retries.
- Developers can retrieve the results at any time, typically after the `Brain` has stopped running or when a certain `Memory` has reached an expected value.

## Installation

With [Go module](https://github.com/golang/go/wiki/Modules) support, simply add the following import to your code, and then `go mod [tidy|download]` will automatically fetch the necessary dependencies.

```go
import "github.com/zenmodel/zenmodel"
```

Otherwise, run the following Go command to install the `zenmodel` package:

```sh
$ go get -u github.com/zenmodel/zenmodel
```


## Quick Start
Let's use `zenmodel` to build a `Brain` as shown below.

<img src="examples/chat_agent/chat_agent_with_function_calling/chat-agent-with-tools.png" width="476" height="238">

### Defining a Blueprint

Define the graph's topology by outlining a blueprint.

#### 1. Create a blueprint

```go
bp := zenmodel.NewBlueprint()
```

#### 2. Add `Neuron`s

Bind a processing function to a neuron or custom `Processor`. In this example, a function is bound, and its definition is omitted for brevity. For more details, see [examples/chat_agent_with_function_calling](examples/chat_agent/chat_agent_with_function_calling).

```go
// add neuron with function
llm := bp.AddNeuron(chatLLM)
action := bp.AddNeuron(callTools)
```

#### 3. Add `Link`s

There are three types of `Link`s:

- Normal Links: Include the `source Neuron` and `destination Neuron`
- Entry Links: Only have the `destination Neuron`
- End Links: The `Brain` will automatically go into a `Sleeping` state when there are no active `Neuron`s and `Link`s, but you can also explicitly define end links to set an endpoint for the `Brain` to run. You only need to specify the `source Neuron`, the `destination Neuron` will be END

```go
/* This example omits error handling */
// add entry link
_, _ = bp.AddEntryLinkTo(llm)

// add link
continueLink, _ := bp.AddLink(llm, action)
_, _ = bp.AddLink(action, llm)

// add end link
endLink, _ := bp.AddEndLinkFrom(llm)
```

#### 4. Set cast select at a branch

By default, all outbound links of a `Neuron` will propagate (belonging to the default casting group). To set up branch selections where you only want certain links to propagate, define casting groups (CastGroup) along with a casting selection function (CastGroupSelectFunc). Each cast group contains a set of links, and the return string of the cast group selection function determines which cast group to propagate to.

```go
	// add link to cast group of a neuron
_ = llm.AddCastGroup("continue", continueLink)
_ = llm.AddCastGroup("end", endLink)
// bind cast group select function for neuron
llm.BindCastGroupSelectFunc(llmNext)
```

```go
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
```

### Building a `Brain` from a Blueprint

Build with various withOpts parameters, although it can be done without configuring any, similar to the example below, using default construction parameters.

Use BrainLocal implementation to build Brain, you can also use other implementations
```go
brain := brainlocal.BuildBrain(bp)
// brain := brainlite.BuildBrain(bp)
```

### Running the `Brain`

As long as any `Link` or `Neuron` of `Brain` is activated, it is considered to be running.
The `Brain` can only be triggered to run through `Link`s. You can set initial brain memory `Memory` before the `Brain` runs to store some initial context, but this is an optional step. The following methods are used to trigger `Link`s:

- Use `brain.Entry()` to trigger all entry links.
- Use `brain.EntryWithMemory()` to set initial `Memory` and trigger all entry links.
- Use `brain.TrigLinks()` to trigger specific `Links`.
- You can also use `brain.SetMemory()` + `brain.TrigLinks()` to set initial `Memory` and trigger specific `Links`.

⚠️Note: Once a `Link` is triggered, the program is non-block; the operation of the `Brain` is asynchronous.

```go
// import "github.com/sashabaranov/go-openai" // just for message struct

// set memory and trigger all entry links
_ = brain.EntryWithMemory("messages", []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "What is the weather in Boston today?"}})
```

### Retrieving Results from `Memory`

`Brain` operations are asynchronous and unlimited in terms of timing for fetching results. We typically call `Wait()` to wait for `Brain` to enter `Sleeping` state or for a certain `Memory` to reach the expected value before retrieving results. Results are obtained from `Memory`.

```go
// block process until the brain is sleeping
brain.Wait()

messages, _ := json.Marshal(brain.GetMemory("messages"))
fmt.Printf("messages: %s\n", messages)
```


## Concept

### Link

<details>
<summary>Expand to view</summary>

The connection between Neurons is called a `Link`, and `Link` is directional, having a `source` and a `destination`.
Typically, both the `source` and the `destination` specify a Neuron. The method to add a `regular Link` is as follows:

```go
// add Link, return link object
// bp := zenmodel.NewBlueprint()
linkObj, err := bp.AddLink(srcNeuron, destNeuron)
```

#### Entry Link

You can also add an `Entry Link`, this kind of Link does not have a `source Neuron`, and only specifies a `destination Neuron`; its `source` is the user.

```go
// add Entry Link, return link object
linkObj, err := bp.AddEntryLinkTo(destNeuron)
```

#### End Link

Additionally, you can add an `End Link`. This type of Link only specifies a `source Neuron` and cannot specify a `destination Neuron`, automatically directing to the `End Neuron`.
Adding an `End Link` will also create a unique `End Neuron` for the entire Brain (creating one if it does not exist) and set the Link's destination to the `End Neuron`.
This is the sole method to create an `End Neuron`; it cannot be individually created without connecting it.

```go
// add End Link, return link object
linkObj, err := bp.AddEndLinkFrom(src_neuron)
```

</details>


### Neuron

<details>
<summary>Expand to view</summary>

A `Neuron` is a neural cell in the Brain, which can be understood as a processing unit. It executes processing logic and can read from or write to the Brain's Memory. Memory, as the context of the Brain, can be shared by all Neurons.

#### Processor

When adding a `Neuron`, you need to specify its processing logic, either by directly specifying a process function (ProcessFn) or by assigning a custom Processor.

```go
// add Neuron with process function
neuronObj := bp.AddNeuron(processFn)

// add Neuron with custom processor
neuronObj2 := bp.AddNeuronWithProcessor(processor)
```

The function signature for ProcessFn is as follows, where BrainContext is mainly used for reading and writing to the Brain's Memory, details of which are introduced in the [BrainContext section](#BrainContext).

```go
// processFn signature
func(bc processor.BrainContext) error
```

The interface definition for a Processor is:

```go
type Processor interface {
    Process(bc processor.BrainContext) error
    Clone() Processor
}
```

#### End Neuron

`End Neuron` is a special Neuron with no processing logic, serving only as the unique exit for the entire Brain. Each Brain has only one `End Neuron`, and when it is triggered, the Brain will put all Neurons to sleep, and the Brain itself will enter a Sleeping state.

An `End Neuron` is not mandatory. Without it, the Brain can still enter a Sleeping state when there are no active Neurons and Links.

#### CastGroupSelectFunc

`CastGroupSelectFunc` is a propagation selection function used to determine which CastGroup a Neuron will propagate to, essentially, **branch selection**. Each CastGroup contains a set of `outward links (out-link)`. Typically, binding a CastGroupSelectFunc is used together with adding (dividing) a CastGroup.

```go
// bind cast group select function for neuron
neuronObj.BindCastGroupSelectFunc(selectFn)
```

#### CastGroup

A `CastGroup` is a propagation group used to define the downstream branches of a Neuron. It divides the Neuron's `outward links (out-link)`.
***By default, all of a Neuron's `outward links (out-link)` belong to the same `Default CastGroup`***, and the propagation selection function (CastGroupSelectFunc), if unspecified by default, will choose to propagate to the `Default CastGroup`.

This means that by default, after the execution of a Neuron, all of its `outward links (out-link)` are triggered in parallel (note: this does not imply that downstream Neurons will be activated; it depends on the configuration of the downstream Neurons' TriggerGroup).

If branch selection is required, you need to add a CastGroup and bind a CastGroupSelectFunc. All `outward links (out-link)` of the selected CastGroup will be triggered in parallel (the same applies here, whether downstream Neurons are activated depends on the downstream Neurons' TriggerGroup settings).

```go
// AddLinkToCastGroup add links to a specific named cast group.
// if the group does not exist, create the group. Groups that allow empty links.
// The specified link will be removed from the default group if it originally belonged to the default group.
err := neuronObj.AddCastGroup("group_A", linkObj1, linkObj2)
```

#### TriggerGroup

A `TriggerGroup` is a trigger group used to define which of a Neuron's `inward links (in-link)` must be triggered to activate the Neuron. It divides the Neuron's `inward links (in-link)`.

When any one `TriggerGroup` of a Neuron is triggered (a TriggerGroup is considered triggered only when all `inward links (in-link)` within it are triggered), the Neuron is activated. Inspiration is taken from neurotransmitters, which must accumulate to a certain threshold before opening channels for electrical signal transmission.

***By default, each of a Neuron's `inward links (in-link)` belongs to its own separate `TriggerGroup`***, meaning that, by default, the Neuron gets activated if any of its `inward links (in-link)` are triggered.

If you need to wait for multiple upstream Neurons to finish in parallel before activating this Neuron, you need to add a `TriggerGroup`.

```go
// AddTriggerGroup by default, a single in-link is a group of its own. AddTriggerGroup adds the specified in-link to the same trigger group.
// it also creates the trigger group. If the added trigger group contains the existing trigger group, the existing trigger group will be removed. This can also be deduplicated at the same time(you add an exist named group, the existing group will be removed first).
// add trigger group with links
err := neuronObj.AddTriggerGroup(linkObj1, linkObj2)
```

</details>


### Blueprint

<details>
<summary>Expand to view</summary>

`Blueprint` defining the graph topology structure of the Brain, as well as all Neurons and Links, in addition to the Brain's operational parameters. A runnable `Brain` can be built from the `Blueprint`.
Optionally, specific build configuration parameters can also be defined during construction, such as the size of Memory, the number of concurrent Workers for the Brain runtime, etc.

```go
brain := brainlocal.BuildBrain(bp, brainlocal.WithNeuronWorkerNum(3))
```

</details>

### Brain

<details>
<summary>Expand to view</summary>

`Brain` is an instance that can be triggered for execution. Based on the triggered Links, it conducts signals to various Neurons, each executing its own logic and reading from or writing to Memory.

The operation of the Brain is asynchronous, and it does not block the program waiting for an output of a result after being triggered because zenmodel does not define what is considered an expected outcome,
***all aiming to bring novel imagination to the users***.

Users or developers can wait for certain Memory to reach the expected value, or wait for all Neurons to have executed and for the Brain to enter Sleeping, then read Memory to retrieve results. Alternatively, they can keep the Brain running, continually generating outputs.

Use Brain.Shutdown() to release all resource of the current Brain.

#### Memory

`Memory` is the runtime context of the Brain. It remains intact after the Brain goes to sleep and will not be cleared unless `ClearMemory()` is called.
Users can read from and write to Memory during Brain operation via Neuron Processing functions, preset Memory before operation, or read and write Memory from outside (as opposed to within the Neuron Process function) during or after operation.

#### BrainContext

The `ProcessFn` and `CastGroupSelectFunc` functions both include the `BrainRuntime` as part of their parameters. The `BrainRuntime` encapsulates some information about the Brain's runtime, such as the Memory at the time the current Neuron is running, the ID of the Neuron currently being executed. These pieces of information are commonly used in the logic of function execution, and often involve writing to Memory. There are also cases where it is necessary to maintain the operation of the current Neuron while triggering downstream Neurons. The `BrainRuntime` interface is as follows:

```

type BrainContext interface {
	// SetMemory set memories for brain, one key value pair is one memory.
	// memory will lazy initial util `SetMemory` or any link trig
	SetMemory(keysAndValues ...interface{}) error
	// GetMemory get memory by key
	GetMemory(key interface{}) interface{}
	// ExistMemory indicates whether there is a memory in the brain
	ExistMemory(key interface{}) bool
	// DeleteMemory delete one memory by key
	DeleteMemory(key interface{})
	// ClearMemory clear all memories
	ClearMemory()
	// GetCurrentNeuronID get current neuron id
	GetCurrentNeuronID() string
	// ContinueCast keep current process running, and continue cast
	ContinueCast()
}

type BrainContextReader interface {
	// GetMemory get memory by key
	GetMemory(key interface{}) interface{}
	// ExistMemory indicates whether there is a memory in the brain
	ExistMemory(key interface{}) bool
	// GetCurrentNeuronID get current neuron id
	GetCurrentNeuronID() string
}

```

</details>


## How to

<details>
<summary> Parallel and Waiting: How to Build a Brain with Parallel and Waiting Neurons </summary>

- TrigLinks() or Entry() are for parallel triggering of links
- Links in a Cast group are also triggered in parallel after a Neuron is completed
- A Neuron begins its execution only after all the specified upstream Neurons have been completed. This is defined by setting up a trigger group to denote which upstream completions are to be awaited.

See the complete example here: [examples/flow-topology/parallel](./examples/flow-topology/parallel-and-wait)

```go
func main() {
	bp := zenmodel.NewBlueprint()

	input := bp.AddNeuron(inputFn)
	poetryTemplate := bp.AddNeuron(poetryFn)
	jokeTemplate := bp.AddNeuron(jokeFn)
	generate := bp.AddNeuron(genFn)

	inputIn, _ := bp.AddLink(input, generate)
	poetryIn, _ := bp.AddLink(poetryTemplate, generate)
	jokeIn, _ := bp.AddLink(jokeTemplate, generate)

	entryInput, _ := bp.AddEntryLinkTo(input)
	entryPoetry, _ := bp.AddEntryLinkTo(poetryTemplate)
	entryJoke, _ := bp.AddEntryLinkTo(jokeTemplate)
	entryInput.GetID()
	entryPoetry.GetID()
	entryJoke.GetID()

	_ = generate.AddTriggerGroup(inputIn, poetryIn)
	_ = generate.AddTriggerGroup(inputIn, jokeIn)

	brain := brainlocal.BuildBrain(bp)

	// case 1: entry poetry and input
	// expect: generate poetry
	_ = brain.TrigLinks(entryPoetry)
	_ = brain.TrigLinks(entryInput)

	// case 2:entry joke and input
	// expect: generate joke
	//_ = brain.TrigLinks(entryJoke)
	//_ = brain.TrigLinks(entryInput)

	// case 3: entry poetry and joke
	// expect: keep blocking and waiting for any trigger group triggered
	//_ = brain.TrigLinks(entryPoetry)
	//_ = brain.TrigLinks(entryJoke)

	// case 4: entry only poetry
	// expect: keep blocking and waiting for any trigger group triggered
	//_ = brain.TrigLinks(entryPoetry)

	// case 5: entry all
	// expect: The first done trigger group triggered activates the generated Neuron,
	// and the trigger group triggered later does not activate the generated Neuron again.
	//_ = brain.Entry()

	brain.Wait()
}

func inputFn(b processor.BrainContext) error {
	_ = b.SetMemory("input", "orange")
	return nil
}

func poetryFn(b processor.BrainContext) error {
	_ = b.SetMemory("template", "poetry")
	return nil
}

func jokeFn(b processor.BrainContext) error {
	_ = b.SetMemory("template", "joke")
	return nil
}

func genFn(b processor.BrainContext) error {
	input := b.GetMemory("input").(string)
	tpl := b.GetMemory("template").(string)
	fmt.Printf("Generating %s for %s\n", tpl, input)
	return nil
}

```

</details>

<details>
<summary> Branching: How to Use CastGroup to Build a Branch That Propagates to Multiple Downstreams </summary>

See the complete example here: [examples/flow-topology/branch](./examples/flow-topology/branch/main.go)

```go

func main() {
	bp := zenmodel.NewBlueprint()
	condition := bp.AddNeuron(func(bc processor.BrainContext) error {
		return nil // do nothing
	})
	cellPhone := bp.AddNeuron(func(bc processor.BrainContext) error {
		fmt.Printf("Run here: Cell Phone\n")
		return nil
	})
	laptop := bp.AddNeuron(func(bc processor.BrainContext) error {
		fmt.Printf("Run here: Laptop\n")
		return nil
	})
	ps5 := bp.AddNeuron(func(bc processor.BrainContext) error {
		fmt.Printf("Run here: PS5\n")
		return nil
	})
	tv := bp.AddNeuron(func(bc processor.BrainContext) error {
		fmt.Printf("Run here: TV\n")
		return nil
	})
	printer := bp.AddNeuron(func(bc processor.BrainContext) error {
		fmt.Printf("Run here: Printer\n")
		return nil
	})

	cellPhoneLink, _ := bp.AddLink(condition, cellPhone)
	laptopLink, _ := bp.AddLink(condition, laptop)
	ps5Link, _ := bp.AddLink(condition, ps5)
	tvLink, _ := bp.AddLink(condition, tv)
	printerLink, _ := bp.AddLink(condition, printer)
	// add entry link
	_, _ = bp.AddEntryLinkTo(condition)

	/*
	   Category 1: Electronics
	   - Cell Phone
	   - Laptop
	   - PS5

	   Category 2: Entertainment Devices
	   - Cell Phone
	   - PS5
	   - TV

	   Category 3: Office Devices
	   - Laptop
	   - Printer
	   - Cell Phone
	*/

	_ = condition.AddCastGroup("electronics",
		cellPhoneLink, laptopLink, ps5Link)
	_ = condition.AddCastGroup("entertainment-devices",
		cellPhoneLink, ps5Link, tvLink)
	_ = condition.AddCastGroup("office-devices",
		laptopLink, printerLink, cellPhoneLink)

	condition.BindCastGroupSelectFunc(func(bcr processor.BrainContextReader) string {
		return bcr.GetMemory("category").(string)
	})

	brain := brainlocal.BuildBrain(bp)

	_ = brain.EntryWithMemory("category", "electronics")
	//_ = brain.EntryWithMemory("category", "entertainment-devices")
	//_ = brain.EntryWithMemory("category", "office-devices")
	//_ = brain.EntryWithMemory("category", "NOT-Defined")

	brain.Wait()
}
```

</details>

<details>

<summary> Nesting: How to Use a Brain as a Neuron within Another Brain </summary>

You can refer to the agent neuron in [plan-and-excute](./examples/plan-and-excute/agent.go), which is a nested brain: [openai_tool_agent](https://github.com/zenmodel/zenmodel/community/tree/main/brain/openai_tool_agent)

You can also refer to the example [nested](./examples/flow-topology/nested/main.go) as follows:

```go
func main() {
	bp := zenmodel.NewBlueprint()
	nested := bp.AddNeuron(nestedBrain)

	_, _ = bp.AddEntryLinkTo(nested)

	brain := brainlocal.BuildBrain(bp)
	_ = brain.Entry()
	brain.Wait()

	fmt.Printf("nested result: %s\n", brain.GetMemory("nested_result").(string))
	
    // nested result: run here neuron: nested.run
}

func nestedBrain(outerBrain processor.BrainContext) error {
	bp := zenmodel.NewBlueprint()
	run := bp.AddNeuron(func(curBrain processor.BrainContext) error {
		_ = curBrain.SetMemory("result", fmt.Sprintf("run here neuron: %s.%s", outerBrain.GetCurrentNeuronID(), curBrain.GetCurrentNeuronID()))
		return nil
	})

	_, _ = bp.AddEntryLinkTo(run)

	brain := brainlocal.BuildBrain(bp)

	// run nested brain
	_ = brain.Entry()
	brain.Wait()
	// get nested brain result
	result := brain.GetMemory("result").(string)
	// pass nested brain result to outer brain
	_ = outerBrain.SetMemory("nested_result", result)

	return nil
}

```

</details>

<details>
<summary> How to Reuse Other Processors within a Processor </summary>

The [zenmodel-contrib](https://github.com/zenmodel/zenmodel/community) community offers many full-featured Processors, or your project's code may have implemented other Processors. Sometimes you need to utilize the functionalities of these Processors, use a combination of multiple Processors, or add extra functionality to existing Processors.
In these cases, you can reuse other Processors within your current Processor or ProcessFn by simply passing the `BrainRuntime` of your current Processor or ProcessFn as a parameter to the other Processor or ProcessFn.

For example, in the `QAProcess` function of [multi-agent/agent-supervisor](./examples/multi-agent/agent-supervisor/qa.go), it reuses the [GoCodeTestProcessor](https://github.com/zenmodel/zenmodel/community/blob/main/processor/go_code_tester/processor.go) from the [zenmodel-contrib](https://github.com/zenmodel/zenmodel/community) community and adds extra functionality after reusing the Processor.

```go
func QAProcess(b processor.BrainContext) error {
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

</details>

## Agent Examples

### Tool Use Agent

#### ChatAgent: With `Function Calling`

ChatAgent takes a list of chat messages as input and outputs new messages to this list. In this example, OpenAI's `function calling` feature is utilized. It is recommended to use in models facilitated with the `function calling` feature.

- [Chat Agent With Tools](./examples/chat_agent/chat_agent_with_function_calling): An example of creating a chat agent from scratch using tools.

### Reflection / Self-Critique

When output quality becomes a primary concern, a combination of self-reflection, self-critique, and external validation is often used to optimize the system output. The example below shows how to implement such a design.

- [Basic Reflection](./examples/reflection): Adds a simple "reflect" step in the `Brain` to prompt your system for output modification.

### Plan and Execute

The examples below implement a typical "plan and execute" style of agent architecture, where an LLM planner decomposes user requests into a program, an executor executes the program, and the LLM synthesizes responses (and/or dynamically replans) based on the program’s output.

- [Plan & Execute](./examples/plan-and-excute): A simple agent with a Planner that generates a multistep task list, an Executing Agent that invokes tools from the plan, and a replanner that responds or creates an updated plan.

### Multi-Agent

Multi-agent systems consist of multiple decision-making agents that interact in a shared environment to achieve common or conflicting goals.

- [agent-supervisor](./examples/multi-agent/agent-supervisor): An example of a multi-agent system with an agent supervisor to help delegate tasks. In the example, the Leader delegates tasks to RD (Research and Development) and QA (Quality Assurance), if the code doesn’t pass the test, it is sent back to RD for rewriting and then tested again, and the Leader makes corresponding decisions based on feedback, finally returning the tested code.

## 🎉 One More Thing

Here we introduce the [zenmodel-contrib](https://github.com/zenmodel/zenmodel/community) repository, a community-driven collection of `Brain` and `Processor` contributions.
At [zenmodel-contrib](https://github.com/zenmodel/zenmodel/community), every line of code is a testament to ideas and innovation. Go ahead, unleash your creativity, and build your `Brain` like assembling Lego bricks. Also, you can find other members' creative ideas here, expanding the boundaries of your thoughts.

Let's have a look at the current list of resources, awaiting your discovery and innovation:

#### Brain

| Brain                                                                                      | Introduction                                                |
| ------------------------------------------------------------------------------------------ | ----------------------------------------------------------- |
| [openai_tool_agent](https://github.com/zenmodel/zenmodel/community/tree/main/brain/openai_tool_agent) | A chat agent based on the OpenAI model, with tool support and calling |

#### Processor

| Processor                                                                                                                         | Introduction                                                 |
| --------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| [calltools](https://github.com/zenmodel/zenmodel/community/tree/main/processor/calltools)                                           | A Processor that calls tools, with tool support and calling  |
| [openaichat](https://github.com/zenmodel/zenmodel/community/tree/main/processor/openaichat)                                         | A chat Processor based on the OpenAI model                   |
| [openai_structured_output](https://github.com/zenmodel/zenmodel/community/tree/main/processor/openai_structured_output)             | A structured output Processor based on OpenAI Function Calling |
| [go_code_tester](https://github.com/zenmodel/zenmodel/community/tree/main/processor/go_code_tester)                                 | A Go unit test runner, often used for testing code generated by LLM |

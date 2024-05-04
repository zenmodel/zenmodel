---
title: "Get Started"
date: 2020-06-08T08:06:25+06:00
description: Get Started of ZenModel
menu:
  sidebar:
    name: "Get Started"
    identifier: get-started
    weight: 10
tags: ["Basic"]
categories: ["Basic"]
---

## Overview

[ZenModel](https://github.com/zenmodel/zenmodel) is a workflow programming framework designed for constructing agentic applications with LLMs. It implements by the scheduling of computational units (`Neuron`), that may include loops, by constructing a `Brain` (a directed graph that can have cycles) or support the loop-less DAGs. A `Brain` consists of multiple `Neurons` connected by `Link`s. Inspiration was drawn from [LangGraph](https://github.com/langchain-ai/langgraph). The `Memory` of a `Brain` leverages [ristretto](https://github.com/dgraph-io/ristretto) for its implementation.

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

<img src="https://github.com/zenmodel/zenmodel/blob/main/examples/chat_agent/chat_agent_with_function_calling/chat-agent-with-tools.png?raw=true" width="476" height="238">


### Defining a Brainprint

Define the graph's topology by outlining a brainprint (a shorthand for brain blueprint).

#### 1. Create a brainprint

```go
bp := zenmodel.NewBrainPrint()
```

#### 2. Add `Neuron`s

Bind a processing function to a neuron or custom `Processor`. In this example, a function is bound, and its definition is omitted for brevity. For more details, see [examples/chat_agent_with_function_calling](https://github.com/zenmodel/zenmodel/blob/main/examples/chat_agent/chat_agent_with_function_calling).

```go
// add neuron with function
bp.AddNeuron("llm", chatLLM)
bp.AddNeuron("action", callTools)
```

#### 3. Add `Link`s

There are three types of `Link`s:

- Normal Links: Include the `source Neuron` and `destination Neuron`
- Entry Links: Only have the `destination Neuron`
- End Links: The `Brain` will automatically go into a `Sleeping` state when there are no active `Neuron`s and `Link`s, but you can also explicitly define end links to set an endpoint for the `Brain` to run. You only need to specify the `source Neuron`, the `destination Neuron` will be END

```go
/* This example omits error handling */
// add entry link
_, _ = bp.AddEntryLink("llm")

// add link
continueLink, _ := bp.AddLink("llm", "action")
_, _ = bp.AddLink("action", "llm")

// add end link
endLink, _ := bp.AddEndLink("llm")
```

#### 4. Set cast select at a branch

By default, all outbound links of a `Neuron` will propagate (belonging to the default casting group). To set up branch selections where you only want certain links to propagate, define casting groups (CastGroup) along with a casting selection function (CastGroupSelectFunc). Each cast group contains a set of links, and the return string of the cast group selection function determines which cast group to propagate to.

```go
// add link to cast group of a neuron
_ = bp.AddLinkToCastGroup("llm", "continue", continueLink)
_ = bp.AddLinkToCastGroup("llm", "end", endLink)
// bind cast group select function for neuron
_ = bp.BindCastGroupSelectFunc("llm", llmNext)
```

```go
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
```

### Building a `Brain` from a Brainprint

Build with various withOpts parameters, although it can be done without configuring any, similar to the example below, using default construction parameters.

```go
brain := bp.Build()
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


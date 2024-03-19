# ZenModel

## 概述
[ZenModel](https://github.com/zenmodel/zenmodel) 是一个用于构建大模型应用的工作流编程框架。它通过构建 `Brain`(一个有向的、允许有环的图)
来支持调度存在环路的计算单元（`Neuron`）或者无环 DAG 的运行。`Brain` 由多个 `Neuron` 组成，`Neuron` 之间通过 `Link`
连接。它的灵感来自 [LangGraph](https://github.com/langchain-ai/langgraph) 。 Brain `Memory` 引用 [ristretto](https://github.com/dgraph-io/ristretto) 实现。

- 开发者可以构建出任意执行流程的 `Brain` 。
  - 串行：按顺序执行 `Neuron`。
  - 并行与等待：并发的执行 `Neuron`，并且支持下游 `Neuron` 等待指定的上游全都执行完成后才开始执行。
  - 分支：执行流程只传播到某一或某些下游分支。
  - 循环：循环对于类似代理（Agent）的行为很重要，您在循环中调用 LLM，询问它下一步要采取什么行动。
  - 有终点：在特定条件下结束运行。比如得到了想要的结果后结束运行。
  - 无终点：持续运行。例如语音通话的场景，持续监听用户说话。
- 每个 `Neuron` 是实际的计算单元，开发者可以自定义 `Neuron` 来实现包括 LLM 调用、其他多模态模型调用等任意处理过程（`Processor`）以及处理的超时、重试等控制机制。
- 开发者可以在任意时机获取运行的结果，通常我们可以等待 `Brain` 停止运行后或者是某个 `Memory` 达到预期值之后去获取结果。

## 安装

```go
import "github.com/zenmodel/zenmodel"
```

## 快速入门

### 定义蓝图 brainprint 

通过定义 brainprint (brain blueprint 大脑蓝图的简称) 来定义图的拓扑结构

#### 1. 创建 brainprint

```go
bp := zenmodel.NewBrainPrint()
```

#### 2. 添加神经元 `Neuron`

可以为 neuron 绑定的处理函数，或自定义 `Processor`，此示例为绑定函数，函数的定义省略，详见 [examples/chat_agent_with_function_calling](examples/chat_agent/chat_agent_with_function_calling/main.go)。

```go
// add neuron with function
bp.AddNeuron("llm", chatLLM)
bp.AddNeuron("action", callTools)
```

#### 2. 添加连接 `Link`

`Link` 有 3 类：
- 普通连接 (Link): 包含 `源 Neuron` 和 `目的 Nueron`
- 入口连接 (Entry Link): 只有 `目的 Nueron`
- 终点连接 (End Link): 当 `Brain` 不存在活跃的 `Neuron` 和 `Link` 时会自动休眠，但也可以显式的定义终点连接来为 `Brain` 指定运行的终点。只需要指定  `源 Neuron`，  `目的 Nueron` 为 END

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

#### 3. 设置分支处的传播选择

默认情况下 `Neuron` 的出向连接全都会进行传播（属于默认传播组），如果要设置分支选择，希望只有某些连接会进行传播，那就需要设置传播组（CastGroup）和传播选择函数（CastGroupSelectFunc）。每个传播组包含一组连接，传播选择函数的返回字符串决定传播到哪个传播组。

```go
	// add link to cast group of a neuron
_ = bp.AddLinkToCastGroup("llm", "continue", continueLink)
_ = bp.AddLinkToCastGroup("llm", "end", endLink)
// bind cast group select function for neuron
_ = bp.BindCastGroupSelectFunc("llm", llmNext)
```

```go
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
```

### 从蓝图构建 `Brain`
构建时可以携带各种 withOpts 参数，当然也可以像示例中一样不配置，使用默认构建参数。
```go
brain := bp.Build()
```

### 运行 `Brain`
只要 `Brain` 有任何 `Link` 或 `Neuron` 激活，就处于运行状态。  
仅可以通过触发 `Link` 来运行 `Brain`。在 `Brain` 运行之前也可以设置初始大脑记忆 `Memory` 来存入一些初始上下文，但这是可选的步骤。下面方法用来触发 `Link` :

- 通过 `brain.Entry()` 来触发所有入口连接
- 通过 `brain.EntryWithMemory()` 来设置初始 `Memory` 并且触发所有入口连接
- 通过 `brain.TrigLinks()` 来触发指定 `Links`
- 也可以通过 `brain.SetMemory()` + `brain.TrigLinks()` 来设置初始 `Memory` 并且触发指定 `Links`

⚠️注意：触发 `Link` 之后，程序不会阻塞，`Brain` 的运行是异步的。
```go
// import "github.com/sashabaranov/go-openai" // just for message struct

// set memory and trig all entry links
_ = brain.EntryWithMemory("messages", []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "What is the weather in Boston today?"}})
```


### 从 `Memory` 中获取结果

`Brain` 的运行是异步的，我们获取运行结果的时机也是是没有限制的，通常我们可以等待 `Brain` 状态变为 `Sleeping` 或者是某个 `Memory` 达到预期值之后去获取结果。结果是从 `Memory` 中获取的。
```go
// block process util brain sleeping
for brain.GetState() != zenmodel.BrainStateSleeping {
    time.Sleep(1 * time.Second)
}

v, found := brain.GetMemory("messages")
if found {
    messages, _ := json.Marshal(v)
    fmt.Printf("messages: %s\n", messages)
}
```

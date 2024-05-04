---
title: "概念"
date: 2020-06-08T08:06:25+06:00
description: ZenModel 概念
menu:
  sidebar:
    name: 概念
    identifier: concept
    weight: 20
tags: ["Concept", "Basic"]
categories: ["Basic"]
---

### Link


Neuron 之间的连接是 `Link`，`Link` 是有方向的，具备`源`和`目的` 。
通常情况下，`源`和`目的`都指定了 Neuron。添加`普通 Link` 的方法如下：

```go
// add Link, return link ID
// bp := zenmodel.NewBrainPrint()
id, err := bp.AddLink("src_neuron", "dest_neuron")
```

#### Entry Link

也可以添加 `Entry Link`, 这种 Link 没有 `源 Neuron`，仅指定了 `目的 Neuron`，它的`源`是用户。

```go
// add Entry Link, return link ID
id, err := bp.AddEntryLink("dest_neuron")
```

#### End Link

也可以添加 `End Link`, 这种 Link 仅指定了 `源 Neuron`，不可指定 `目的 Neuron`，固定为 `End Neuron` 。
添加 `End Link` 的同时，也会创建全 Brain 唯一的 `End Neuron`（如果不存在则创建），并将 Link 的目的地指向 `End Neuron`。
这也是唯一的途径创建 `End Neuron`，无法单独创建一个 `End Neuron` 而不去连接它。

```go
// add End Link, return link ID
id, err := bp.AddEndLink("src_neuron")
```



### Neuron



`Neuron` 是 Brain 中的神经元，可以理解为一个处理单元，它执行处理逻辑，并且可以读写 Brain 的 Memory。Memory 作为 Brain
的上下文可以被所有 Neuron 共享。

#### Processor

添加 `Neuron` 时需要指定此 `Neuron` 的处理逻辑，可以直接指定处理函数(ProcessFn) 或者指定自定义的 Processor 。

```go
// add Neuron with process function
bp.AddNeuron("neuron_id", processFn)

// add Neuron with custom processor
bp.AddNeuronWithProcessor("neuron_id", processor)
```

ProcessFn 的函数签名如下，其中 BrainRuntime 是主要用来读写 Brain 的 Memory 的，细节在 [BrainRuntime 小节](#BrainRuntime)
介绍。

```go
// processFn signature
func(runtime BrainRuntime) error
```

Processor 的接口定义如下:

```go
type Processor interface {
    Process(brain BrainRuntime) error
    DeepCopy() Processor
}
```

#### End Neuron

`End Neuron` 是一种特殊的 Neuron，它没有处理逻辑，仅作为全 Brain 唯一的出口。 `End Neuron` 是每个 Brain
唯一的，当 `End Neuron` 被触发时，Brain 就会休眠所有 Neuron 并且自身也会处于 Sleeping 状态。

`End Neuron` 不是必须的，没有`End Neuron` Brain 也可以运转到 Sleeping 状态，当没有任何活跃的 Neuron 和 Link 时也会进入
Sleeping 状态。

#### CastGroupSelectFunc

`CastGroupSelectFunc` 传播选择函数，用来判定当前 Neuron 将会传播到哪个 CastGroup，也就是**分支选择**。 每个 CastGroup
包含一组 `出向连接(out-link)`。通常绑定 CastGroupSelectFunc 会和添加（划分） CastGroup 一起使用。

```go
// bind cast group select function for neuron
err := bp.BindCastGroupSelectFunc("neuron_id", selectFn)
```

#### CastGroup

`CastGroup` 传播组是用来定义 Neuron 下游分支的。它划分了 Neuron 的 `出向连接(out-link)`。
***默认情况下 Neuron 的所有`出向连接(out-link)`  都属于同一个 `Default CastGroup`***
，并且传播选择函数（CastGroupSelectFunc）如果不指定，默认会选择传播到 `Default CastGroup` 。

也就是说默认情况下，在 Neuron 执行完成后，当前 Neuron 的所有 `出向连接(out-link)` 都是并行触发的(注意：这不代表下游所有
Neuron 都会被激活，还需要看下游 Neuron 的 TriggerGroup 配置)。

如果需要分支选择，那就需要添加 CastGroup 并且绑定 CastGroupSelectFunc，被选中的 CastGroup 中的所有 `出向连接(out-link)`
都将会并行触发（同上，下游 Neuron 是否被激活还需看下游 Neuron 的 TriggerGroup 配置）。

```go
// AddLinkToCastGroup add links to specific named cast group.
// if group not exist, create the group. Groups that allow empty links.
// The specified link will remove from the default group, if it originally belonged to the default group.
err := bp.AddLinkToCastGroup("neuron_id", "group_A", linkID1, linkID2)
```

#### TriggerGroup

`TriggerGroup` 触发组是用来定义 Neuron 的哪些 `入向连接(in-link)` 被触发之后就激活此 Neuron 的。它划分了 Neuron
的 `入向连接(in-link)`。

当 Neuron 的任意一个 `TriggerGroup` 被触发时（某个 `TriggerGroup` 中所有 `入向连接(in-link)` 都被触发则此 TriggerGroup
才被触发），Neuron 就会被激活。灵感来自于神经递质累积到一定阈值才会打开通道进行电信号传递。

***默认情况下 Neuron 的每一条`入向连接(in-link)` 都各自单独属于一个 `TriggerGroup`*** 。也就是说默认情况下，Neuron
只要有任意一条 `入向连接(in-link)` 被触发，Neuron 就会被激活。

如果需要等待上游多个 Neuron 并行完成之后，再激活此 Neuron，那就需要添加 `TriggerGroup` 。

```go
// AddTriggerGroup by default, a single in-link is a group of its own. AddTriggerGroup adds the specified in-link to the same trigger group.
// it also creates the trigger group. If the added trigger group contains the existing trigger group, the existing trigger group will be removed. This can also be deduplicated at the same time(you add an exist named group, existing group will be removed first).
// add trigger group with links
err := bp.AddTriggerGroup("neuron_id", "group_B", linkID1, linkID2)
```



### Brainprint


`Brainprint` 是大脑蓝图(Brain Blueprint) 的简称，它定义了 Brain 的图拓扑结构以及所有 Neuron 和 Link 以及 Brain
的运行参数。可以通过 `Brainprint` 构建出可运行的 `Brain`。
在构建时也可选的能够指定构建的配置参数，例如 Memory 大小，Brain 运行时的并发 Worker 数等。

```go
brain := bp.Build(zenmodel.WithWorkerNum(3), )
```


### Brain


`Brain` 是可触发运行的实例。根据触发的 Link 传导到各个 Neuron，每个 Neuron 执行各自的逻辑并且读写 Memory。

Brain 的运行是异步的，触发后不会阻塞程序直到输出一个结果，因为 zenmodel 不去定义何为预期的结果，
***旨在给用户带来新的想象力***。

用户或者开发者可以等待某个 Memory 到达预期值，或者等待所有 Neuron 执行完毕 Brain Sleeping，然后去读取 Memory 获取到结果。
也可以使 Brain 保持运行，持续输出结果。

#### Memory

`Memory` 是 Brain 运行时的上下文，在 Brain Sleeping 之后，也不会被清除，除非调用了 ClearMemory() 。
用户可以在运行时通过 Neuron 的 Process 函数读写 Memory，也可以在运行前预设 Memory，当然也可以在运行结束后或者运行期间在外部（相较于
Neuron Process 函数的内部）读写 Memory。

#### BrainRuntime

`ProcessFn` 和 `CastGroupSelectFunc` 这些函数的参数中都有 `BrainRuntime`,
`BrainRuntime` 包含了 Brain 运行时的一些信息，例如运行到当前 Neuron 时的 Memory， 当前执行的 Neuron 的
ID，函数执行的逻辑中通常会使用到这些信息，也会进行 Memory 的写入，也有情况会需要保持当前 Neuron 运行的同时触发下游 Neuron。
`BrainRuntime` 接口如下：

```go
type BrainRuntime interface {
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
```


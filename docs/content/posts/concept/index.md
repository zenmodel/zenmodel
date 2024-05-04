---
title: "Concept"
date: 2020-06-08T08:06:25+06:00
description: Concept Of ZenModel
menu:
  sidebar:
    name: Concept
    identifier: concept
    weight: 20
tags: ["Concept", "Basic"]
categories: ["Basic"]
---


### Link


The connection between Neurons is called a `Link`, and `Link` is directional, having a `source` and a `destination`.
Typically, both the `source` and the `destination` specify a Neuron. The method to add a `regular Link` is as follows:

```go
// add Link, return link ID
// bp := zenmodel.NewBrainPrint()
id, err := bp.AddLink("src_neuron", "dest_neuron")
```

#### Entry Link

You can also add an `Entry Link`, this kind of Link does not have a `source Neuron`, and only specifies a `destination Neuron`; its `source` is the user.

```go
// add Entry Link, return link ID
id, err := bp.AddEntryLink("dest_neuron")
```

#### End Link

Additionally, you can add an `End Link`. This type of Link only specifies a `source Neuron` and cannot specify a `destination Neuron`, automatically directing to the `End Neuron`.
Adding an `End Link` will also create a unique `End Neuron` for the entire Brain (creating one if it does not exist) and set the Link's destination to the `End Neuron`.
This is the sole method to create an `End Neuron`; it cannot be individually created without connecting it.

```go
// add End Link, return link ID
id, err := bp.AddEndLink("src_neuron")
```


### Neuron


A `Neuron` is a neural cell in the Brain, which can be understood as a processing unit. It executes processing logic and can read from or write to the Brain's Memory. Memory, as the context of the Brain, can be shared by all Neurons.

#### Processor

When adding a `Neuron`, you need to specify its processing logic, either by directly specifying a process function (ProcessFn) or by assigning a custom Processor.

```go
// add Neuron with process function
bp.AddNeuron("neuron_id", processFn)

// add Neuron with custom processor
bp.AddNeuronWithProcessor("neuron_id", processor)
```

The function signature for ProcessFn is as follows, where BrainRuntime is mainly used for reading and writing to the Brain's Memory, details of which are introduced in the [BrainRuntime section](#BrainRuntime).

```go
// processFn signature
func(runtime BrainRuntime) error
```

The interface definition for a Processor is:

```go
type Processor interface {
    Process(brain BrainRuntime) error
    DeepCopy() Processor
}
```

#### End Neuron

`End Neuron` is a special Neuron with no processing logic, serving only as the unique exit for the entire Brain. Each Brain has only one `End Neuron`, and when it is triggered, the Brain will put all Neurons to sleep, and the Brain itself will enter a Sleeping state.

An `End Neuron` is not mandatory. Without it, the Brain can still enter a Sleeping state when there are no active Neurons and Links.

#### CastGroupSelectFunc

`CastGroupSelectFunc` is a propagation selection function used to determine which CastGroup a Neuron will propagate to, essentially, **branch selection**. Each CastGroup contains a set of `outward links (out-link)`. Typically, binding a CastGroupSelectFunc is used together with adding (dividing) a CastGroup.

```go
// bind cast group select function for neuron
err := bp.BindCastGroupSelectFunc("neuron_id", selectFn)
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
err := bp.AddLinkToCastGroup("neuron_id", "group_A", linkID1, linkID2)
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
err := bp.AddTriggerGroup("neuron_id", "group_B", linkID1, linkID2)
```



### Brainprint


`Brainprint` is an abbreviation for Brain Blueprint, defining the graph topology structure of the Brain, as well as all Neurons and Links, in addition to the Brain's operational parameters. A runnable `Brain` can be built from the `Brainprint`.
Optionally, specific build configuration parameters can also be defined during construction, such as the size of Memory, the number of concurrent Workers for the Brain runtime, etc.

```go
brain := bp.Build(zenmodel.WithWorkerNum(3), )
```


### Brain


`Brain` is an instance that can be triggered for execution. Based on the triggered Links, it conducts signals to various Neurons, each executing its own logic and reading from or writing to Memory.

The operation of the Brain is asynchronous, and it does not block the program waiting for an output of a result after being triggered because zenmodel does not define what is considered an expected outcome,
***all aiming to bring novel imagination to the users***.

Users or developers can wait for certain Memory to reach the expected value, or wait for all Neurons to have executed and for the Brain to enter Sleeping, then read Memory to retrieve results. Alternatively, they can keep the Brain running, continually generating outputs.

#### Memory

`Memory` is the runtime context of the Brain. It remains intact after the Brain goes to sleep and will not be cleared unless `ClearMemory()` is called.
Users can read from and write to Memory during Brain operation via Neuron Processing functions, preset Memory before operation, or read and write Memory from outside (as opposed to within the Neuron Process function) during or after operation.

#### BrainRuntime

The `ProcessFn` and `CastGroupSelectFunc` functions both include the `BrainRuntime` as part of their parameters. The `BrainRuntime` encapsulates some information about the Brain's runtime, such as the Memory at the time the current Neuron is running, the ID of the Neuron currently being executed. These pieces of information are commonly used in the logic of function execution, and often involve writing to Memory. There are also cases where it is necessary to maintain the operation of the current Neuron while triggering downstream Neurons. The `BrainRuntime` interface is as follows:

```go
type BrainRuntime interface {
    // SetMemory sets memories for the brain, one key-value pair is one memory.
    // Memory will lazily initialize until `SetMemory` or any link is triggered
    SetMemory(keysAndValues ...interface{}) error
    // GetMemory retrieves memory by key
    GetMemory(key interface{}) interface{}
    // ExistMemory indicates whether there is a memory in the brain
    ExistMemory(key interface{}) bool
    // DeleteMemory deletes a single memory by key
    DeleteMemory(key interface{})
    // ClearMemory clears all memories
    ClearMemory()
    // GetCurrentNeuronID gets the current neuron's ID
    GetCurrentNeuronID() string
    // ContinueCast keeps the current process running, and continues casting
    ContinueCast()
}
```


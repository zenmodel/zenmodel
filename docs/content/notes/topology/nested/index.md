---
title: Nested
weight: 40
menu:
  notes:
    name: Nested
    identifier: nested
    parent: topology
    weight: 40
---

<!-- Nested -->

{{< note title="Nested" >}}

```go
package main

import (
	"fmt"

	"github.com/zenmodel/zenmodel"
)

func main() {
	bp := zenmodel.NewBrainPrint()
	bp.AddNeuron("nested", nestedBrain)
	_, _ = bp.AddEntryLink("nested")

	brain := bp.Build()
	_ = brain.Entry()
	brain.Wait()

	fmt.Printf("nested result: %s\n", brain.GetMemory("nested_result").(string))
}

func nestedBrain(outerBrain zenmodel.BrainRuntime) error {
	bp := zenmodel.NewBrainPrint()
	bp.AddNeuron("run", func(curBrain zenmodel.BrainRuntime) error {
		_ = curBrain.SetMemory("result", fmt.Sprintf("run here neuron: %s.%s", outerBrain.GetCurrentNeuronID(), curBrain.GetCurrentNeuronID()))
		return nil
	})
	_, _ = bp.AddEntryLink("run")

	brain := bp.Build()

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

{{< /note >}}

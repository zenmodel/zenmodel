package main

import (
	"fmt"

	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/brainlocal"
	"github.com/zenmodel/zenmodel/processor"
)

func main() {
	bp := zenmodel.NewBlueprint()
	nested := bp.AddNeuron(nestedBrain)

	_, _ = bp.AddEntryLinkTo(nested)

	brain := brainlocal.BuildBrain(bp)
	_ = brain.Entry()
	brain.Wait()

	fmt.Printf("nested result: %s\n", brain.GetMemory("nested_result").(string))
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

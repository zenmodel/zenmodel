package tests

import (
	"fmt"
	"testing"

	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/brainlocal"
	"github.com/zenmodel/zenmodel/processor"
)

func TestNested(t *testing.T) {
	bp := zenmodel.NewBlueprint()
	nested := bp.AddNeuron(nestedBrain)

	_, _ = bp.AddEntryLinkTo(nested)

	brain := brainlocal.BuildBrain(bp)
	
	fmt.Println("-----\nTesting Nested Brain:")
	_ = brain.Entry()
	brain.Wait()

	result := brain.GetMemory("nested_result").(string)
	fmt.Printf("Nested result: %s\n", result)

	brain.Shutdown()
}

func nestedBrain(outerBrain processor.BrainContext) error {
	bp := zenmodel.NewBlueprint()
	run := bp.AddNeuron(func(curBrain processor.BrainContext) error {
		result := fmt.Sprintf("run here neuron: %s.%s", outerBrain.GetCurrentNeuronID(), curBrain.GetCurrentNeuronID())
		fmt.Printf("Inner Brain: %s\n", result)
		_ = curBrain.SetMemory("result", result)
		return nil
	})

	_, _ = bp.AddEntryLinkTo(run)

	brain := brainlocal.BuildBrain(bp)

	_ = brain.Entry()
	brain.Wait()
	result := brain.GetMemory("result").(string)
	_ = outerBrain.SetMemory("nested_result", result)
	
	brain.Shutdown()

	return nil
}

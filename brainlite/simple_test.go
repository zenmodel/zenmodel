package brainlite

import (
	"fmt"
	"testing"

	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/processor"
)

func TestSimpleBrain(t *testing.T) {
	bp := zenmodel.NewBlueprint()
	n1 := bp.AddNeuron(fn1)
	n2 := bp.AddNeuron(fn2)
	
	_, err := bp.AddLink(n1, n2)
	if err != nil {
		fmt.Printf("add link error: %s\n", err)
		return
	}
	_, err = bp.AddEntryLinkTo(n1)
	if err != nil {
		fmt.Printf("add entry link error: %s\n", err)
		return
	}

	brain := BuildBrain(bp)

	_ = brain.Entry()

	brain.Wait()

	name := brain.GetMemory("name").(string)
	fmt.Printf("result: my name is %s.\n", name)
	brain.Shutdown()
}

func fn1(b processor.BrainContext) error {
	fmt.Println("start fn1 ..............")

	if err := b.SetMemory("name", "Clay"); err != nil {
		return err
	}

	return nil
}

func fn2(b processor.BrainContext) error {
	fmt.Println("start fn2 ..............")

	firstName := b.GetMemory("name").(string)

	name := firstName + " Zhang"
	if err := b.SetMemory("name", name); err != nil {
		return err
	}
	return nil
}

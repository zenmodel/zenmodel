package main

import (
	"fmt"
	"time"

	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/brainlite"
	"github.com/zenmodel/zenmodel/processor"
)

func main() {
	bp := zenmodel.NewMultiLangBlueprint()
	n0 := bp.AddNeuron(date)
	n1 := bp.AddNeuronWithPyProcessor("a/b/c","setname", "SetNameProcessor", map[string]interface{}{"lastname": "Zhang"})
	n2 := bp.AddNeuronWithPyProcessor("d/e/f/","add", "AddProcessor", map[string]interface{}{"a": 1, "b": 2})

	_, err := bp.AddEntryLinkTo(n0)
	if err != nil {
		fmt.Printf("add entry link error: %s\n", err)
		return
	}
	_, err = bp.AddLink(n0, n1)
	if err != nil {
		fmt.Printf("add link error: %s\n", err)
		return
	}
	_, err = bp.AddLink(n1, n2)
	if err != nil {
		fmt.Printf("add link error: %s\n", err)
		return
	}
	_, err = bp.AddEndLinkFrom(n2)
	if err != nil {
		fmt.Printf("add end link error: %s\n", err)
		return
	}

	brain := brainlite.BuildMultiLangBrain(bp)

	err = brain.EntryWithMemory("name", "Clay")
	if err != nil {
		fmt.Printf("entry error: %s\n", err)
		return
	}
	brain.Wait()

	answer := brain.GetMemory("answer").(string)
	fmt.Printf("answer: %s\n", answer)
	brain.Shutdown()
}

func date(b processor.BrainContext) error {
	fmt.Println("start date ..............")

	if err := b.SetMemory("date", time.Now().Format("2006-01-02")); err != nil {
		return err
	}

	return nil
}
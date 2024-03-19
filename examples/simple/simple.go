package main

import (
	"fmt"
	"time"

	"github.com/zenmodel/zenmodel"
)

func main() {
	bp := zenmodel.NewBrainPrint()
	bp.AddNeuron("n1", fn1)
	bp.AddNeuron("n2", fn2)
	_, err := bp.AddLink("n1", "n2")
	if err != nil {
		fmt.Printf("add link error: %s\n", err)
		return
	}
	_, err = bp.AddEntryLink("n1")
	if err != nil {
		fmt.Printf("add entry link error: %s\n", err)
		return
	}

	//bp.AddLinkToCastGroup()
	//bp.AddTriggerGroup()
	brain := bp.Build()

	brain.Entry()

	time.Sleep(10 * time.Second)

	name, found := brain.GetMemory("name")
	if !found {
		fmt.Println("name not found")
		return
	}
	fmt.Printf("result: my name is %s.\n", name)
	//brain.GetStatus()
}

func fn1(b zenmodel.Brain) error {
	fmt.Println("start fn1 ..............")

	if err := b.SetMemory("name", "Clay"); err != nil {
		return err
	}

	return nil
}

func fn2(b zenmodel.Brain) error {
	fmt.Println("start fn2 ..............")

	firstName, found := b.GetMemory("name")
	if !found {
		return nil
	}

	name := firstName.(string) + " Zhang"
	if err := b.SetMemory("name", name); err != nil {
		return err
	}
	return nil
}

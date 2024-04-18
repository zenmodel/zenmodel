package main

import (
	"fmt"

	"github.com/zenmodel/zenmodel"
)

func main() {
	bp := zenmodel.NewBrainPrint()
	bp.AddNeuron("condition", func(runtime zenmodel.BrainRuntime) error {
		return nil // do nothing
	})
	bp.AddNeuron("cell-phone", func(runtime zenmodel.BrainRuntime) error {
		fmt.Printf("Run here: Cell Phone\n")
		return nil
	})
	bp.AddNeuron("laptop", func(runtime zenmodel.BrainRuntime) error {
		fmt.Printf("Run here: Laptop\n")
		return nil
	})
	bp.AddNeuron("ps5", func(runtime zenmodel.BrainRuntime) error {
		fmt.Printf("Run here: PS5\n")
		return nil
	})
	bp.AddNeuron("tv", func(runtime zenmodel.BrainRuntime) error {
		fmt.Printf("Run here: TV\n")
		return nil
	})
	bp.AddNeuron("printer", func(runtime zenmodel.BrainRuntime) error {
		fmt.Printf("Run here: Printer\n")
		return nil
	})

	cellPhone, _ := bp.AddLink("condition", "cell-phone")
	laptop, _ := bp.AddLink("condition", "laptop")
	ps5, _ := bp.AddLink("condition", "ps5")
	tv, _ := bp.AddLink("condition", "tv")
	printer, _ := bp.AddLink("condition", "printer")
	// add entry link
	_, _ = bp.AddEntryLink("condition")

	/*
	   Category 1: Electronics
	   - Cell Phone
	   - Laptop
	   - PS5

	   Category 2: Entertainment Devices
	   - Cell Phone
	   - PS5
	   - TV

	   Category 3: Office Devices
	   - Laptop
	   - Printer
	   - Cell Phone
	*/
	_ = bp.AddLinkToCastGroup("condition", "electronics",
		cellPhone, laptop, ps5)
	_ = bp.AddLinkToCastGroup("condition",
		"entertainment-devices",
		cellPhone, ps5, tv)
	_ = bp.AddLinkToCastGroup(
		"condition", "office-devices",
		laptop, printer, cellPhone)

	_ = bp.BindCastGroupSelectFunc("condition", func(brain zenmodel.BrainRuntime) string {
		return brain.GetMemory("category").(string)
	})

	brain := bp.Build()

	_ = brain.EntryWithMemory("category", "electronics")
	//_ = brain.EntryWithMemory("category", "entertainment-devices")
	//_ = brain.EntryWithMemory("category", "office-devices")
	//_ = brain.EntryWithMemory("category", "NOT-Defined")

	brain.Wait()
}

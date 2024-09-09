package main

import (
	"fmt"

	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/brainlocal"
	"github.com/zenmodel/zenmodel/processor"
)

func main() {
	bp := zenmodel.NewBlueprint()
	condition := bp.AddNeuron(func(bc processor.BrainContext) error {
		return nil // do nothing
	})
	cellPhone := bp.AddNeuron(func(bc processor.BrainContext) error {
		fmt.Printf("Run here: Cell Phone\n")
		return nil
	})
	laptop := bp.AddNeuron(func(bc processor.BrainContext) error {
		fmt.Printf("Run here: Laptop\n")
		return nil
	})
	ps5 := bp.AddNeuron(func(bc processor.BrainContext) error {
		fmt.Printf("Run here: PS5\n")
		return nil
	})
	tv := bp.AddNeuron(func(bc processor.BrainContext) error {
		fmt.Printf("Run here: TV\n")
		return nil
	})
	printer := bp.AddNeuron(func(bc processor.BrainContext) error {
		fmt.Printf("Run here: Printer\n")
		return nil
	})

	cellPhoneLink, _ := bp.AddLink(condition, cellPhone)
	laptopLink, _ := bp.AddLink(condition, laptop)
	ps5Link, _ := bp.AddLink(condition, ps5)
	tvLink, _ := bp.AddLink(condition, tv)
	printerLink, _ := bp.AddLink(condition, printer)
	// add entry link
	_, _ = bp.AddEntryLinkTo(condition)

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

	_ = condition.AddCastGroup("electronics",
		cellPhoneLink, laptopLink, ps5Link)
	_ = condition.AddCastGroup("entertainment-devices",
		cellPhoneLink, ps5Link, tvLink)
	_ = condition.AddCastGroup("office-devices",
		laptopLink, printerLink, cellPhoneLink)

	condition.BindCastGroupSelectFunc(func(bcr processor.BrainContextReader) string {
		return bcr.GetMemory("category").(string)
	})

	brain := brainlocal.NewBrainLocal(bp)

	_ = brain.EntryWithMemory("category", "electronics")
	//_ = brain.EntryWithMemory("category", "entertainment-devices")
	//_ = brain.EntryWithMemory("category", "office-devices")
	//_ = brain.EntryWithMemory("category", "NOT-Defined")

	brain.Wait()
}

package main

import (
	"fmt"

	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/brainlocal"
	"github.com/zenmodel/zenmodel/processor"
)

func main() {
	bp := zenmodel.NewBlueprint()

	input := bp.AddNeuron(inputFn)
	poetryTemplate := bp.AddNeuron(poetryFn)
	jokeTemplate := bp.AddNeuron(jokeFn)
	generate := bp.AddNeuron(genFn)

	inputIn, _ := bp.AddLink(input, generate)
	poetryIn, _ := bp.AddLink(poetryTemplate, generate)
	jokeIn, _ := bp.AddLink(jokeTemplate, generate)

	entryInput, _ := bp.AddEntryLinkTo(input)
	entryPoetry, _ := bp.AddEntryLinkTo(poetryTemplate)
	entryJoke, _ := bp.AddEntryLinkTo(jokeTemplate)
	entryInput.GetID()
	entryPoetry.GetID()
	entryJoke.GetID()

	_ = generate.AddTriggerGroup(inputIn, poetryIn)
	_ = generate.AddTriggerGroup(inputIn, jokeIn)

	brain := brainlocal.BuildBrain(bp)

	// case 1: entry poetry and input
	// expect: generate poetry
	_ = brain.TrigLinks(entryPoetry)
	_ = brain.TrigLinks(entryInput)

	// case 2:entry joke and input
	// expect: generate joke
	//_ = brain.TrigLinks(entryJoke)
	//_ = brain.TrigLinks(entryInput)

	// case 3: entry poetry and joke
	// expect: keep blocking and waiting for any trigger group triggered
	//_ = brain.TrigLinks(entryPoetry)
	//_ = brain.TrigLinks(entryJoke)

	// case 4: entry only poetry
	// expect: keep blocking and waiting for any trigger group triggered
	//_ = brain.TrigLinks(entryPoetry)

	// case 5: entry all
	// expect: The first done trigger group triggered activates the generated Neuron,
	// and the trigger group triggered later does not activate the generated Neuron again.
	//_ = brain.Entry()

	brain.Wait()
}

func inputFn(b processor.BrainContext) error {
	_ = b.SetMemory("input", "orange")
	return nil
}

func poetryFn(b processor.BrainContext) error {
	_ = b.SetMemory("template", "poetry")
	return nil
}

func jokeFn(b processor.BrainContext) error {
	_ = b.SetMemory("template", "joke")
	return nil
}

func genFn(b processor.BrainContext) error {
	input := b.GetMemory("input").(string)
	tpl := b.GetMemory("template").(string)
	fmt.Printf("Generating %s for %s\n", tpl, input)
	return nil
}

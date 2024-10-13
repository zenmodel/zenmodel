package tests

import (
	"fmt"
	"testing"

	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/brainlite"
	"github.com/zenmodel/zenmodel/processor"
)

func TestParallelAndWait(t *testing.T) {
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

	_ = generate.AddTriggerGroup(inputIn, poetryIn)
	_ = generate.AddTriggerGroup(inputIn, jokeIn)

	brain := brainlite.BuildBrain(bp)

	fmt.Println("-----\nTesting Poetry and Input:")
	_ = brain.TrigLinks(entryPoetry)
	_ = brain.TrigLinks(entryInput)
	brain.Wait()

	fmt.Println("\n-----\nTesting Joke and Input:")
	_ = brain.TrigLinks(entryJoke)
	_ = brain.TrigLinks(entryInput)
	brain.Wait()

	brain.Shutdown()
}

func inputFn(b processor.BrainContext) error {
	fmt.Println("Input function called")
	_ = b.SetMemory("input", "orange")
	return nil
}

func poetryFn(b processor.BrainContext) error {
	fmt.Println("Poetry function called")
	_ = b.SetMemory("template", "poetry")
	return nil
}

func jokeFn(b processor.BrainContext) error {
	fmt.Println("Joke function called")
	_ = b.SetMemory("template", "joke")
	return nil
}

func genFn(b processor.BrainContext) error {
	input := b.GetMemory("input").(string)
	tpl := b.GetMemory("template").(string)
	result := fmt.Sprintf("Generating %s for %s", tpl, input)
	fmt.Println(result)
	return nil
}

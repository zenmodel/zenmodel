package main

import (
	"fmt"

	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/brainlocal"
	"github.com/zenmodel/zenmodel/processor"
)

const (
	memKeyObjective = "objective"
	memKeyPlan      = "plan"
	memKeyPastSteps = "past_steps"
	memKeyResponse  = "response"
)

func main() {
	bp := zenmodel.NewBlueprint()

	// add planner neuron
	pp, _ := PlannerProcessor()
	neuronPlanner := bp.AddNeuronWithProcessor(pp)

	// add tool agent neuron
	neuronAgent := bp.AddNeuron(toolAgentProcess)

	// add replanner neuron
	rpp, _ := RePlannerProcessor()
	neuronRePlanner := bp.AddNeuronWithProcessor(rpp)

	// add link
	_, _ = bp.AddEntryLinkTo(neuronPlanner)
	_, _ = bp.AddLink(neuronPlanner, neuronAgent)
	_, _ = bp.AddLink(neuronAgent, neuronRePlanner)
	continueLink, _ := bp.AddLink(neuronRePlanner, neuronAgent)
	endLink, _ := bp.AddEndLinkFrom(neuronRePlanner)

	// add link to cast group of a neuron
	_ = neuronPlanner.AddCastGroup("continue", continueLink)
	_ = neuronPlanner.AddCastGroup("end", endLink)

	// bind cast group select function for neuron
	neuronPlanner.BindCastGroupSelectFunc(replanerNext)

	// build brain
	brain := brainlocal.NewBrainLocal(bp)
	// set memory and trig all entry links
	_ = brain.EntryWithMemory(memKeyObjective, "what is the hometown of the 2024 Australia open winner?")
	// block process util brain sleeping
	brain.Wait()

	fmt.Printf("past_steps: %s\n", brain.GetMemory(memKeyPastSteps).(PastSteps))
	fmt.Printf("response: %s\n", brain.GetMemory(memKeyResponse).(*Response))

	/*
		past_steps:
		step1:
		        task: Identify the winner of the 2024 Australia Open.
		        result: The winner of the 2024 Australia Open was Jannik Sinner. He won his first ever Grand Slam title with an epic comeback victory against Daniil Medvedev. Sinner climbed back from a two-set deficit to win the match with a score of 3-6, 3-6, 6-4, 6-4, 6-3.

		step2:
		        task: Research Jannik Sinner's biographical details to find his hometown.
		        result: Jannik Sinner was born on August 16, 2001, in the San Candido region in northern Italy. His parents are Johann Sinner and Siglinde Sinner. He hails from the northern Italian region of South Tyrol, which borders Austria to the east and west with the Swiss canton of Graubünden to the west. [Source](https://www.tennis365.com/tennis-features/who-are-jannik-sinners-parents)


		response: &{The hometown of the 2024 Australia Open winner, Jannik Sinner, is San Candido in northern Italy. No further steps are needed as this information conclusively answers the objective.}

	*/
}

func replanerNext(b processor.BrainContextReader) string {
	// if we got response, turn end
	if b.ExistMemory(memKeyResponse) {
		return "end"
	}

	return "continue"
}

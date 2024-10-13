package main

import (
	"fmt"

	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/brainlocal"
	"github.com/zenmodel/zenmodel/processor"
)

var (
	FeedBackRD       = "RD"
	FeedBackQA       = "QA"
	DecisionRD       = "RD"
	DecisionQA       = "QA"
	DecisionResponse = "Response"
)

func main() {
	bp := zenmodel.NewBlueprint()

	neuronLeader := bp.AddNeuron(LeaderProcess)
	neuronQA := bp.AddNeuron(QAProcess)
	neuronRD := bp.AddNeuronWithProcessor(NewRDProcessor())

	_, _ = bp.AddEntryLinkTo(neuronLeader)
	// leader out-link
	rdLink, _ := bp.AddLink(neuronLeader, neuronRD)
	qaLink, _ := bp.AddLink(neuronLeader, neuronQA)
	endLink, _ := bp.AddEndLinkFrom(neuronLeader)

	// leader in-link
	_, _ = bp.AddLink(neuronRD, neuronLeader)
	_, _ = bp.AddLink(neuronQA, neuronLeader)

	_ = neuronLeader.AddCastGroup(DecisionRD, rdLink)
	_ = neuronLeader.AddCastGroup(DecisionQA, qaLink)
	_ = neuronLeader.AddCastGroup(DecisionResponse, endLink)
	neuronLeader.BindCastGroupSelectFunc(func(bcr processor.BrainContextReader) string {
		return bcr.GetMemory(memKeyDecision).(string)
	})

	brain := brainlocal.BuildBrain(bp)
	_ = brain.EntryWithMemory(memKeyDemand, "Help me write a function `func Add (x, y int) int` with golang to implement addition, and implement unit test in a separate _test .go file, at least 3 test cases are required")
	brain.Wait()
	fmt.Printf("Response: %s\n", brain.GetMemory(memKeyResponse).(string))

	/*
		Response: Dear Boss:
		After the efforts of our RD team and QA team, the final codes and test report are produced as follows:

		==========

		Codes:

		**add.go**

		```go
		package main

		func Add(x, y int) int {
		        return x + y
		}
		```

		**add_test.go**

		```go
		package main

		import "testing"

		func TestAdd(t *testing.T) {
		        cases := []struct {
		                x, y, expected int
		        }{
		                {1, 2, 3},
		                {-1, 1, 0},
		                {0, 0, 0},
		        }

		        for _, c := range cases {
		                result := Add(c.x, c.y)
		                if result != c.expected {
		                        t.Errorf("Add(%d, %d) == %d, expected %d", c.x, c.y, result, c.expected)
		                }
		        }
		}
		```

		==========

		Test Report:

		```shell
		#go test -v -run .
		=== RUN   TestAdd
		--- PASS: TestAdd (0.00s)
		PASS
		ok      gocodetester    0.411s

		```


	*/
}

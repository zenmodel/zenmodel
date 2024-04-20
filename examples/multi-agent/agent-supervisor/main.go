package main

import (
	"fmt"

	"github.com/zenmodel/zenmodel"
)

var (
	NeuronLeader     = "Leader"
	NeuronRD         = "RD"
	NeuronQA         = "QA"
	DecisionRD       = "RD"
	DecisionQA       = "QA"
	DecisionResponse = "Response"
)

func main() {
	bp := zenmodel.NewBrainPrint()

	bp.AddNeuron(NeuronLeader, LeaderProcess)
	bp.AddNeuron(NeuronQA, QAProcess)
	bp.AddNeuronWithProcessor(NeuronRD, NewRDProcessor())

	_, _ = bp.AddEntryLink(NeuronLeader)
	// leader out-link
	rdLink, _ := bp.AddLink(NeuronLeader, NeuronRD)
	qaLink, _ := bp.AddLink(NeuronLeader, NeuronQA)
	endLink, _ := bp.AddEndLink(NeuronLeader)

	// leader in-link
	_, _ = bp.AddLink(NeuronRD, NeuronLeader)
	_, _ = bp.AddLink(NeuronQA, NeuronLeader)

	_ = bp.AddLinkToCastGroup(NeuronLeader, DecisionRD, rdLink)
	_ = bp.AddLinkToCastGroup(NeuronLeader, DecisionQA, qaLink)
	_ = bp.AddLinkToCastGroup(NeuronLeader, DecisionResponse, endLink)
	_ = bp.BindCastGroupSelectFunc(NeuronLeader, func(b zenmodel.BrainRuntime) string {
		return b.GetMemory(memKeyDecision).(string)
	})

	brain := bp.Build()
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

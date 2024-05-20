package main

import (
	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/community/processor/go_code_tester"
)

const (
	memKeyGoTestResult = "go_test_result"
)

func QAProcess(b zenmodel.BrainRuntime) error {
	p := go_code_tester.NewProcessor().WithTestCodeKeep(true)
	if err := p.Process(b); err != nil {
		return err
	}

	if err := b.SetMemory(memKeyFeedback, b.GetCurrentNeuronID()); err != nil {
		return err
	}

	return nil
}

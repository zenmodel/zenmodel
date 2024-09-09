package main

import (
	"fmt"
	"strings"

	"github.com/zenmodel/zenmodel/community/processor/go_code_tester"
	"github.com/zenmodel/zenmodel/processor"
)

const (
	memKeyDemand   = "demand"
	memKeyResponse = "response"
	memKeyTask     = "task"
	memKeyFeedback = "feedback"
	memKeyDecision = "decision"
)

func LeaderProcess(b processor.BrainContext) error {
	// if it has no task, disassemble task from demand
	if !b.ExistMemory(memKeyTask) {
		task := rephraseTaskFromDemand(b.GetMemory(memKeyDemand).(string))
		_ = b.SetMemory(memKeyTask, task)
		_ = b.SetMemory(memKeyDecision, DecisionRD)

		return nil
	}
	switch b.GetMemory(memKeyFeedback).(string) {
	case FeedBackRD: // feedback from RD
		_ = b.SetMemory(memKeyDecision, DecisionQA) // pass to QA
	case FeedBackQA: // feedback from QA
		ok := readTestReport(b.GetMemory(memKeyGoTestResult).(string))
		if !ok {
			// test result not ok, resend to RD
			_ = b.SetMemory(memKeyDecision, DecisionRD)
		} else {
			// pretty response from codes
			resp := genResponse(b)
			_ = b.SetMemory(memKeyResponse, resp)
			_ = b.SetMemory(memKeyDecision, DecisionResponse)
		}
	default:
		return fmt.Errorf("unknown feedback: %v\n", b.GetMemory(memKeyFeedback))
	}

	return nil
}

func rephraseTaskFromDemand(demand string) string {
	// TODO maybe use completion LLM to rephrase demand to task
	task := demand

	return task
}

func readTestReport(testResult string) bool {
	return !strings.Contains(testResult, "FAIL")
}

func genResponse(b processor.BrainContextReader) string {
	codes := b.GetMemory(memKeyCodes).(*go_code_tester.Codes).String()
	testReport := b.GetMemory(memKeyGoTestResult).(string)

	var builder strings.Builder
	builder.WriteString("Dear Boss:  \n")
	builder.WriteString("After the efforts of our RD team and QA team, the final codes and test report are produced as follows:\n\n")
	builder.WriteString("==========\n\nCodes:\n\n")
	builder.WriteString(codes)
	builder.WriteString("==========\n\nTest Report:\n\n")
	builder.WriteString("```shell\n")
	builder.WriteString(testReport)
	builder.WriteString("```")
	builder.WriteString("\n")

	return builder.String()
}

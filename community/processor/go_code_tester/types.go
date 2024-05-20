package go_code_tester

import (
	"fmt"
	"strings"
)

type CodeFile struct {
	Language string `json:"language" jsonschema:"description=the programming language of code,enum=go,enum=shell,enum=python"`
	Path     string `json:"path" jsonschema_description:"code file path, include file name"`
	Content  string `json:"content" jsonschema_description:"code content"`
}

func (c CodeFile) FunctionName() string {
	return "code_file"
}
func (c CodeFile) FunctionDescription() string {
	return "Code file including file path and code content"
}

func (c CodeFile) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("**%s**\n\n", c.Path))
	builder.WriteString(fmt.Sprintf("```%s\n", c.Language))
	builder.WriteString(c.Content)
	builder.WriteString(fmt.Sprintf("\n```\n\n"))

	return builder.String()
}

type Codes struct {
	Description string     `json:"description,omitempty" jsonschema:"description=description of codes"`
	CodeFiles   []CodeFile `json:"code_files" jsonschema:"description=some dependent code files"`
}

func (c Codes) FunctionName() string {
	return "codes"
}

func (c Codes) FunctionDescription() string {
	return "some dependent code files, each with its path and code content"
}

func (c Codes) String() string {
	var builder strings.Builder
	if c.Description != "" {
		builder.WriteString(c.Description)
		builder.WriteString("\n\n")
	}
	for _, code := range c.CodeFiles {
		builder.WriteString(code.String())
	}

	return builder.String()
}

package go_code_tester

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zenmodel/zenmodel"
	"github.com/zenmodel/zenmodel/community/common/log"
	"go.uber.org/zap"
)

const (
	workspacePath = "./.zenmodel/processor/go-code-tester"
	defaultGoBin  = "go"
	goModName     = "gocodetester"

	memKeyGoTestResult = "go_test_result"
)

var (
	defaultTestFlag = []string{"-v", "-run"}

	memKeyCodes = (&Codes{}).FunctionName()
)

func NewProcessor() *GoCodeTestProcessor {
	processor := &GoCodeTestProcessor{
		testFlag:     defaultTestFlag,
		goBin:        defaultGoBin,
		keepTestCode: false,
		logger:       log.NewDefaultLoggerWithLevel(zap.InfoLevel),
	}

	return processor
}

type GoCodeTestProcessor struct { // nolint
	testFlag     []string
	goBin        string
	keepTestCode bool

	logger *zap.Logger
}

func (p *GoCodeTestProcessor) WithLogger(logger *zap.Logger) *GoCodeTestProcessor {
	p.logger = logger
	return p
}

// WithTestFlag will reset test flag, default test flag is '-v', '-run'
func (p *GoCodeTestProcessor) WithTestFlag(testFlag []string) *GoCodeTestProcessor {
	p.testFlag = testFlag
	return p
}

// WithGoBin will reset go bin, default go bin is 'go'
func (p *GoCodeTestProcessor) WithGoBin(goBin string) *GoCodeTestProcessor {
	p.goBin = goBin
	return p
}

// WithTestCodeKeep will keep test code after test, default is false
func (p *GoCodeTestProcessor) WithTestCodeKeep(keep bool) *GoCodeTestProcessor {
	p.keepTestCode = keep
	return p
}

func (p *GoCodeTestProcessor) Process(brain zenmodel.BrainRuntime) error {
	p.logger.Info("go code test processor start processing")

	codes, ok := brain.GetMemory(memKeyCodes).(*Codes)
	if !ok {
		return fmt.Errorf("invalid memory type of key %s", memKeyCodes)
	}
	if err := cleanWorkspace(); err != nil {
		return err
	}
	if err := createCodeFiles(codes); err != nil {
		return err
	}
	defer func() {
		if !p.keepTestCode {
			_ = cleanWorkspace()
		}
	}()
	if err := p.goModInitAndTidy(); err != nil {
		return err
	}

	out, err := p.goCodeTest()
	if err != nil {
		return err
	}

	p.logger.Debug("go test result", zap.String("result", out))
	if err = brain.SetMemory(memKeyGoTestResult, out); err != nil {
		return err
	}

	return nil
}

func (p *GoCodeTestProcessor) DeepCopy() zenmodel.Processor {
	testFlagCopy := make([]string, len(p.testFlag))
	if p.testFlag != nil {
		copy(testFlagCopy, p.testFlag)
	}

	return &GoCodeTestProcessor{
		testFlag: testFlagCopy,
		goBin:    p.goBin,
		logger:   p.logger.WithOptions(), // with no option, only for clone
	}
}

func cleanWorkspace() error {
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		err = os.MkdirAll(workspacePath, 0755)
		return err
	}

	dir, err := os.Open(workspacePath)
	if err != nil {
		return err
	}
	defer dir.Close()

	objects, err := dir.Readdir(-1) // 返回 workspacePath 文件夹下所有的文件和文件夹
	if err != nil {
		return err
	}

	for _, obj := range objects {
		objPath := filepath.Join(workspacePath, obj.Name())
		err := os.RemoveAll(objPath) // 删除文件或文件夹
		if err != nil {
			return err
		}
	}
	return nil
}

func createCodeFiles(codes *Codes) error {
	for _, codeFile := range codes.CodeFiles {
		if codeFile.Language != "go" {
			continue
		}
		completePath := filepath.Join(workspacePath, codeFile.Path)
		dirPath := filepath.Dir(completePath)

		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(completePath, []byte(codeFile.Content), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *GoCodeTestProcessor) goModInitAndTidy() error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s mod init %s", p.goBin, goModName))
	cmd.Dir = workspacePath
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("%s mod tidy", p.goBin))
	cmd.Dir = workspacePath
	err = cmd.Run()

	return err
}

func (p *GoCodeTestProcessor) goCodeTest() (string, error) {
	goTestCmd := fmt.Sprintf("%s test %s . ", p.goBin, strings.Join(p.testFlag, " "))
	cmd := exec.Command("/bin/sh", "-c", goTestCmd)
	cmd.Dir = workspacePath

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("#%s\n%s\n", goTestCmd, out.String()), nil
}

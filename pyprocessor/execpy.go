package pyprocessor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/zenmodel/zenmodel/processor"
)

func LoadPythonProcessor(pyCodePath, moduleName, processorClassName string, constructorArgs map[string]interface{}) *ExecPyProcessor {
	return &ExecPyProcessor{
		pyCodePath:         pyCodePath,
		moduleName:         moduleName,
		processorClassName: processorClassName,
		constructorArgs:    constructorArgs,
		scriptPath:         "temp_exec_script.py",
		pythonCmd:          "/usr/bin/python3",
	}
}

type ExecPyProcessor struct {
	pyCodePath         string
	moduleName         string
	processorClassName string
	constructorArgs    map[string]interface{}
	scriptPath         string
	pythonCmd          string
}

func (p *ExecPyProcessor) Process(ctx processor.BrainContext) error {
	if l := ctx.GetCurrentNeuronLabels(); l != nil {
		if v, ok := l["python_cmd"]; ok {
			p.pythonCmd = v
		}
	}

	err := p.createTempPythonScript()
	if err != nil {
		return fmt.Errorf("create temp python script failed: %s", err)
	}
	defer os.Remove(p.scriptPath)

	return p.execPythonScript(fmt.Sprintf("%s.db", ctx.GetBrainID()))
}

func (p *ExecPyProcessor) Clone() processor.Processor {
	return &ExecPyProcessor{
		pyCodePath:         p.pyCodePath,
		moduleName:         p.moduleName,
		processorClassName: p.processorClassName,
		constructorArgs:    p.constructorArgs,
		scriptPath:         p.scriptPath,
		pythonCmd:          p.pythonCmd,
	}
}

func (p *ExecPyProcessor) createTempPythonScript() error {
	// 替换路径分隔符为点号
	importPath := strings.ReplaceAll(p.pyCodePath, string(os.PathSeparator), ".")
	// 移除开头结尾的点号（如果存在）
	importPath = strings.TrimSuffix(importPath, ".")
	importPath = strings.TrimPrefix(importPath, ".")

	content := fmt.Sprintf(`
import sys
import json
import os

from %s.%s import %s
from zenmodel import BrainContext

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python script.py <db_path> <params_json>")
        sys.exit(1)

    db_path = sys.argv[1]
    params_json = sys.argv[2]

    params = json.loads(params_json)

    processor = %s(**params)
    ctx = BrainContext(db_path)
    processor.process(ctx)
`, importPath, p.moduleName, p.processorClassName, p.processorClassName)

	return os.WriteFile(p.scriptPath, []byte(content), 0644)
}

func (p *ExecPyProcessor) execPythonScript(sqliteDBPath string) error {
	// 将参数转换为JSON字符串
	paramsJSON, err := json.Marshal(p.constructorArgs)
	if err != nil {
		return fmt.Errorf("参数序列化错误: %s", err)
	}

	// 构造Python命令
	cmd := exec.Command(p.pythonCmd, p.scriptPath, sqliteDBPath, string(paramsJSON))

	// 获取标准错误和标准输出管道
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("无法获取标准输出管道: %s", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("无法获取标准错误输出管道: %s", err)
	}

	// 启动Python进程
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("启动Python进程失败: %s", err)
	}

	// 创建一个多路复用的reader
	reader := io.MultiReader(stdoutPipe, stderrPipe)

	// 读取所有输出
	fmt.Println("python processor:")
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Println("  " + scanner.Text())
	}

	// 等待Python进程结束
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("Python进程执行错误: %s", err)
	}

	return nil
}
